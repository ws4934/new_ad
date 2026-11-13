// Package updater 基于 GitHub Release 的在线更新。
//
// 流程: 查最新 Release → 下载 ad-monitor_<tag>_linux_amd64.tar.gz + SHA256SUMS → 校验 →
// 解出新二进制并用 `--version` 自检 → 原地替换当前可执行文件 (旧版留 .bak) →
// 给自己发 SIGTERM 走正常优雅关闭 → cmd 在 s.Run() 返回后 syscall.Exec 新二进制 (PID 不变)。
// 裸机与 Docker 一致; Docker 下可执行文件位于 data 卷 (见 docker-entrypoint.sh), 重建容器不丢。
package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Version 由 CI 通过 -ldflags "-X ad-monitor/internal/logic/updater.Version=vX.Y.Z" 注入。
var Version = "dev"

// apiBase / githubBase 仅供端到端测试通过 -ldflags -X 指向模拟服务。
var (
	apiBase    = "https://api.github.com"
	githubBase = "https://github.com"
)

// proxyPrefix 配置 adMonitor.update.githubProxy (如 https://gh-proxy.com/), 国内服务器直连 GitHub 失败时使用。
// 注意: 代理方能看到并可篡改下载内容 (校验文件也经同一代理), 只填信任的代理。
func proxyPrefix(ctx context.Context) string {
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "adMonitor.update.githubProxy").String())
}

const (
	Repo       = "ws4934/new_ad"
	binaryName = "ad-monitor"
	checkTTL   = 10 * time.Minute
	maxBinary  = 256 << 20
)

// Phase 更新阶段
const (
	PhaseIdle        = "idle"
	PhaseDownloading = "downloading"
	PhaseVerifying   = "verifying"
	PhaseRestarting  = "restarting"
	PhaseFailed      = "failed"
)

// Status 返回给前端的更新状态。
type Status struct {
	Supported         bool   `json:"supported"`
	UnsupportedReason string `json:"unsupportedReason"`
	CurrentVersion    string `json:"currentVersion"`
	LatestVersion     string `json:"latestVersion"`
	ReleaseNotes      string `json:"releaseNotes"`
	ReleaseUrl        string `json:"releaseUrl"`
	PublishedAt       string `json:"publishedAt"`
	UpdateAvailable   bool   `json:"updateAvailable"`
	CheckedAt         string `json:"checkedAt"`
	Phase             string `json:"phase"`
	Message           string `json:"message"`
	Error             string `json:"error"`
}

type release struct {
	TagName     string `json:"tag_name"`
	Body        string `json:"body"`
	HtmlUrl     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	Assets      []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

var (
	mu        sync.Mutex
	latest    *release
	checkedAt time.Time
	phase     = PhaseIdle
	message   string
	lastErr   string
	pending   string // 已替换完成、等待 exec 的可执行文件路径

	httpClient = &http.Client{Timeout: 5 * time.Minute}
)

// unsupportedReason 返回空串表示当前进程可在线更新。
func unsupportedReason() string {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		return "仅支持 linux/amd64"
	}
	if parseVersion(Version) == nil {
		return "开发版本 (" + Version + ") 不支持在线更新, 请使用 Release 构建"
	}
	if _, err := executable(); err != nil {
		return err.Error()
	}
	return ""
}

func executable() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if p, err = filepath.EvalSymlinks(p); err != nil {
		return "", err
	}
	dir := filepath.Dir(p)
	f, err := os.CreateTemp(dir, ".write-test-*")
	if err != nil {
		return "", fmt.Errorf("程序目录 %s 不可写: %v", dir, err)
	}
	f.Close()
	os.Remove(f.Name())
	return p, nil
}

// GetStatus 当前状态 (不发网络请求)。
func GetStatus() Status {
	mu.Lock()
	defer mu.Unlock()
	return statusLocked()
}

func statusLocked() Status {
	reason := unsupportedReason()
	s := Status{
		Supported: reason == "", UnsupportedReason: reason, CurrentVersion: Version,
		Phase: phase, Message: message, Error: lastErr,
	}
	if latest != nil {
		s.LatestVersion = latest.TagName
		s.ReleaseNotes = latest.Body
		s.ReleaseUrl = latest.HtmlUrl
		s.PublishedAt = latest.PublishedAt
		s.UpdateAvailable = Compare(latest.TagName, Version) > 0
	}
	if !checkedAt.IsZero() {
		s.CheckedAt = checkedAt.Format("2006-01-02 15:04:05")
	}
	return s
}

// Check 查询 GitHub 最新 Release; force=false 时 10 分钟内复用缓存。
func Check(ctx context.Context, force bool) (Status, error) {
	mu.Lock()
	fresh := !force && time.Since(checkedAt) < checkTTL
	mu.Unlock()
	if fresh {
		return GetStatus(), nil
	}
	rel, err := fetchLatest(ctx)
	mu.Lock()
	defer mu.Unlock()
	if err != nil {
		return statusLocked(), err
	}
	latest, checkedAt = rel, time.Now()
	return statusLocked(), nil
}

// fetchLatest 未配置代理: 先走 API (含更新说明), 失败 (限流/网络) 再走 latest/download/SHA256SUMS;
// 配置了代理: 只走代理 + SHA256SUMS (代理一般不支持 api.github.com)。
func fetchLatest(ctx context.Context) (*release, error) {
	if p := proxyPrefix(ctx); p != "" {
		return fetchLatestSums(ctx, p)
	}
	rel, err := fetchLatestAPI(ctx)
	if err == nil {
		return rel, nil
	}
	if rel2, err2 := fetchLatestSums(ctx, ""); err2 == nil {
		return rel2, nil
	}
	return nil, err
}

var sumsTagRe = regexp.MustCompile(binaryName + `_(v\d+\.\d+\.\d+(?:\.\d+)?)_linux_amd64\.tar\.gz`)

// releaseFromSums 由 SHA256SUMS 内容推出版本号, 资产地址按 GitHub 固定格式拼出。
func releaseFromSums(sums string) (*release, error) {
	m := sumsTagRe.FindStringSubmatch(sums)
	if m == nil {
		return nil, errors.New("SHA256SUMS 中没有可识别的安装包")
	}
	tag := m[1]
	dl := githubBase + "/" + Repo + "/releases/download/" + tag + "/"
	rel := &release{TagName: tag, HtmlUrl: githubBase + "/" + Repo + "/releases/tag/" + tag}
	for _, name := range []string{m[0], "SHA256SUMS"} {
		rel.Assets = append(rel.Assets, struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		}{name, dl + name})
	}
	return rel, nil
}

func fetchLatestSums(ctx context.Context, proxy string) (*release, error) {
	data, err := download(proxy+githubBase+"/"+Repo+"/releases/latest/download/SHA256SUMS", 1<<20)
	if err != nil {
		return nil, fmt.Errorf("获取最新版本失败: %w", err)
	}
	return releaseFromSums(string(data))
}

func fetchLatestAPI(ctx context.Context) (*release, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		apiBase+"/repos/"+Repo+"/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ad-monitor-updater")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("仓库还没有发布任何 Release")
	}
	if resp.Header.Get("X-RateLimit-Remaining") == "0" {
		reset, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
		return nil, fmt.Errorf("GitHub API 访问次数超限 (未登录每小时 60 次/IP), 请在 %s 后重试",
			time.Unix(reset, 0).Format("15:04"))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub 返回 HTTP %d", resp.StatusCode)
	}
	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	if parseVersion(rel.TagName) == nil {
		return nil, fmt.Errorf("Release 版本号无效: %q", rel.TagName)
	}
	return &rel, nil
}

// Start 异步执行更新。restart 在新二进制就位后被调用, 负责触发优雅关闭。
func Start(ctx context.Context, restart func()) (Status, error) {
	if _, err := Check(ctx, true); err != nil {
		return GetStatus(), err
	}
	mu.Lock()
	defer mu.Unlock()
	if reason := unsupportedReason(); reason != "" {
		return statusLocked(), errors.New(reason)
	}
	switch phase {
	case PhaseDownloading, PhaseVerifying, PhaseRestarting:
		return statusLocked(), errors.New("更新正在进行中")
	}
	if latest == nil || Compare(latest.TagName, Version) <= 0 {
		return statusLocked(), errors.New("已是最新版本")
	}
	rel := latest
	phase, message, lastErr = PhaseDownloading, "正在下载 "+rel.TagName, ""
	go func() {
		path, err := install(rel)
		mu.Lock()
		if err != nil {
			phase, message, lastErr = PhaseFailed, "", err.Error()
			mu.Unlock()
			return
		}
		phase, message, pending = PhaseRestarting, "已安装 "+rel.TagName+", 正在重启", path
		mu.Unlock()
		time.Sleep(500 * time.Millisecond) // 让前端轮询拿到 restarting 状态
		restart()
	}()
	return statusLocked(), nil
}

func setMessage(p, m string) {
	mu.Lock()
	phase, message = p, m
	mu.Unlock()
}

// install 下载校验并原地替换当前可执行文件, 返回其路径。
func install(rel *release) (string, error) {
	exe, err := executable()
	if err != nil {
		return "", err
	}
	archiveName := fmt.Sprintf("%s_%s_linux_amd64.tar.gz", binaryName, rel.TagName)
	var archiveURL, sumsURL string
	for _, a := range rel.Assets {
		switch a.Name {
		case archiveName:
			archiveURL = a.URL
		case "SHA256SUMS":
			sumsURL = a.URL
		}
	}
	if archiveURL == "" || sumsURL == "" {
		return "", fmt.Errorf("Release %s 缺少 %s 或 SHA256SUMS", rel.TagName, archiveName)
	}
	p := proxyPrefix(context.Background())
	archiveURL, sumsURL = p+archiveURL, p+sumsURL

	sums, err := download(sumsURL, 1<<20)
	if err != nil {
		return "", fmt.Errorf("下载 SHA256SUMS: %w", err)
	}
	want := checksumFor(string(sums), archiveName)
	if want == "" {
		return "", fmt.Errorf("SHA256SUMS 中没有 %s", archiveName)
	}
	archive, err := download(archiveURL, maxBinary)
	if err != nil {
		return "", fmt.Errorf("下载安装包: %w", err)
	}

	setMessage(PhaseVerifying, "正在校验 "+rel.TagName)
	sum := sha256.Sum256(archive)
	if hex.EncodeToString(sum[:]) != want {
		return "", errors.New("安装包 SHA256 校验失败")
	}

	candidate, err := extract(archive, filepath.Dir(exe))
	if err != nil {
		return "", err
	}
	defer os.Remove(candidate) // 成功 rename 后此文件已不存在

	out, err := exec.Command(candidate, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("新版本自检失败: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != rel.TagName {
		return "", fmt.Errorf("新版本自检版本号不符: 期望 %s, 实际 %s", rel.TagName, got)
	}

	if err := os.Rename(exe, exe+".bak"); err != nil {
		return "", fmt.Errorf("备份旧程序: %w", err)
	}
	if err := os.Rename(candidate, exe); err != nil {
		_ = os.Rename(exe+".bak", exe)
		return "", fmt.Errorf("替换程序: %w", err)
	}
	return exe, nil
}

func download(url string, limit int64) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "ad-monitor-updater")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("文件过大")
	}
	return data, nil
}

// checksumFor 解析 sha256sum 输出 ("<hex>  <name>")。
func checksumFor(sums, name string) string {
	for _, line := range strings.Split(sums, "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name && len(f[0]) == 64 {
			return strings.ToLower(f[0])
		}
	}
	return ""
}

// extract 从 tar.gz 中取出 ad-monitor 写到 dir 下的临时文件 (与目标同目录, 保证 rename 原子)。
func extract(archive []byte, dir string) (string, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return "", err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return "", fmt.Errorf("安装包中没有 %s", binaryName)
		}
		if err != nil {
			return "", err
		}
		if h.Typeflag != tar.TypeReg || filepath.Base(h.Name) != binaryName {
			continue
		}
		f, err := os.CreateTemp(dir, ".ad-monitor-new-*")
		if err != nil {
			return "", err
		}
		_, err = io.Copy(f, io.LimitReader(tr, maxBinary))
		if err == nil {
			err = f.Chmod(0o755)
		}
		if err == nil {
			err = f.Sync()
		}
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			os.Remove(f.Name())
			return "", err
		}
		return f.Name(), nil
	}
}

// RestartIfPending 在 HTTP 服务优雅关闭后调用: 有待生效的新版本则 exec 替换当前进程。
func RestartIfPending() error {
	mu.Lock()
	path := pending
	mu.Unlock()
	if path == "" {
		return nil
	}
	return syscall.Exec(path, os.Args, os.Environ())
}

// SelfTerminate 触发与 Ctrl+C / docker stop 相同的优雅关闭流程。
func SelfTerminate() {
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)
}

// parseVersion 解析 v1.2.3 / v1.2.3.4, 非法返回 nil。
func parseVersion(v string) []int {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".")
	if len(parts) < 3 || len(parts) > 4 {
		return nil
	}
	out := make([]int, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil
		}
		out[i] = n
	}
	return out
}

// Compare a>b 返回 1, a<b 返回 -1, 相等 0; 非法版本视为最小。
func Compare(a, b string) int {
	pa, pb := parseVersion(a), parseVersion(b)
	switch {
	case pa == nil && pb == nil:
		return 0
	case pa == nil:
		return -1
	case pb == nil:
		return 1
	}
	for i := range pa {
		if pa[i] != pb[i] {
			if pa[i] > pb[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
