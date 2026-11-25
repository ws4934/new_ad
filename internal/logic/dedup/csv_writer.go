package dedup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	maxFileSize  = 100 * 1024 * 1024 // 100MB 分片阈值
	writeBufSize = 64 * 1024         // 写缓冲大小, 减少锁内 syscall 次数
)

// CSVWriter 按大小分片追加写 CSV 文件 (带写缓冲, 由 gcron 定时 Flush)
// 文件命名: {platform}_{field}.csv, {platform}_{field}_1.csv, ...
// 当文件达到 100MB 时自动创建下一个分片
type CSVWriter struct {
	baseDir string
	locks   sync.Map // map[prefix]*sync.Mutex
	files   sync.Map // map[prefix]*writerState
}

type writerState struct {
	file    *os.File
	buf     *bufio.Writer
	size    int64
	shardNo int
	prefix  string
	baseDir string
}

// NewCSVWriter 构造
func NewCSVWriter(baseDir string) *CSVWriter {
	_ = os.MkdirAll(baseDir, 0755)
	return &CSVWriter{baseDir: baseDir}
}

// BaseDir 返回 CSV 目录
func (w *CSVWriter) BaseDir() string {
	return w.baseDir
}

// Append 追加一条设备号到 CSV (按大小分片, 不按天)
func (w *CSVWriter) Append(ctx context.Context, platform, field, value string) {
	prefix := platform + "_" + field

	lockI, _ := w.locks.LoadOrStore(prefix, &sync.Mutex{})
	lock := lockI.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	ws := w.getOrCreateWriter(ctx, prefix)
	if ws == nil {
		return
	}

	// 达到 100MB 则切换分片
	if ws.size >= maxFileSize {
		_ = ws.buf.Flush()
		_ = ws.file.Close()
		ws.shardNo++
		ws.file = nil
		ws.buf = nil
		ws.size = 0
		ws = w.openShard(ctx, ws)
		if ws == nil {
			return
		}
	}

	n, err := ws.buf.WriteString(value + "\n")
	if err != nil {
		g.Log().Errorf(ctx, "write csv failed: %v", err)
		return
	}
	ws.size += int64(n)
}

func (w *CSVWriter) getOrCreateWriter(ctx context.Context, prefix string) *writerState {
	if v, ok := w.files.Load(prefix); ok {
		ws := v.(*writerState)
		if ws.file != nil {
			return ws
		}
		return w.openShard(ctx, ws)
	}

	ws := &writerState{prefix: prefix, baseDir: w.baseDir}
	ws.shardNo = w.findLatestShard(prefix)
	ws = w.openShard(ctx, ws)
	if ws != nil {
		w.files.Store(prefix, ws)
	}
	return ws
}

func (w *CSVWriter) openShard(ctx context.Context, ws *writerState) *writerState {
	path := ws.filePath()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		g.Log().Errorf(ctx, "open csv shard failed: %v", err)
		return nil
	}
	info, _ := f.Stat()
	ws.file = f
	ws.buf = bufio.NewWriterSize(f, writeBufSize)
	ws.size = info.Size()
	w.files.Store(ws.prefix, ws)
	return ws
}

func (ws *writerState) filePath() string {
	name := ws.prefix + ".csv"
	if ws.shardNo > 0 {
		name = fmt.Sprintf("%s_%d.csv", ws.prefix, ws.shardNo)
	}
	return filepath.Join(ws.baseDir, name)
}

func (w *CSVWriter) findLatestShard(prefix string) int {
	entries, err := os.ReadDir(w.baseDir)
	if err != nil {
		return 0
	}
	maxShard := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".csv") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".csv")
		if name == prefix {
			continue
		}
		if rest, ok := strings.CutPrefix(name, prefix+"_"); ok {
			var n int
			if _, err := fmt.Sscanf(rest, "%d", &n); err == nil && n > maxShard {
				maxShard = n
			}
		}
	}
	return maxShard
}

// Flush 把所有写缓冲落盘
// 由 gcron 定时调用 + 导出列表/下载前调用, 保证文件内容可见
func (w *CSVWriter) Flush() {
	w.files.Range(func(k, v any) bool {
		lockI, _ := w.locks.LoadOrStore(k.(string), &sync.Mutex{})
		lock := lockI.(*sync.Mutex)
		lock.Lock()
		ws := v.(*writerState)
		if ws.buf != nil {
			_ = ws.buf.Flush()
		}
		lock.Unlock()
		return true
	})
}

// Close 刷缓冲并关闭所有打开的 CSV 文件句柄, 进程退出时调用
func (w *CSVWriter) Close() {
	w.files.Range(func(k, v any) bool {
		lockI, _ := w.locks.LoadOrStore(k.(string), &sync.Mutex{})
		lock := lockI.(*sync.Mutex)
		lock.Lock()
		ws := v.(*writerState)
		if ws.buf != nil {
			_ = ws.buf.Flush()
			ws.buf = nil
		}
		if ws.file != nil {
			_ = ws.file.Close()
			ws.file = nil
		}
		lock.Unlock()
		return true
	})
}

// RemoveFile 删除 CSV 文件 (导出页删除入口)
// 若删除的是正在写入的活跃分片, 先丢弃写缓冲并关闭句柄,
// 避免 writer 继续往已 unlink 的 inode 写 (新设备会静默消失直到分片轮转)。
// 整个过程持有该前缀的锁, 防止删除与并发 Append 竞争。
func (w *CSVWriter) RemoveFile(filename string) error {
	prefix := prefixOfFile(filename)
	lockI, _ := w.locks.LoadOrStore(prefix, &sync.Mutex{})
	lock := lockI.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	if v, ok := w.files.Load(prefix); ok {
		ws := v.(*writerState)
		if ws.file != nil && filepath.Base(ws.filePath()) == filename {
			// 不 flush: 缓冲里的数据属于要删除的文件, 直接丢弃
			ws.buf = nil
			_ = ws.file.Close()
			ws.file = nil
			ws.size = 0
			// 句柄置空后, 下次 Append 会按同分片号重建空文件继续写
		}
	}
	return os.Remove(filepath.Join(w.baseDir, filename))
}

// prefixOfFile 从文件名反推 writer 前缀: 去掉 .csv 和末尾的纯数字分片号
// "weibo_imei_md5_3.csv" → "weibo_imei_md5"; "weibo_imei_md5.csv" → "weibo_imei_md5"
func prefixOfFile(filename string) string {
	name := strings.TrimSuffix(filename, ".csv")
	if i := strings.LastIndex(name, "_"); i > 0 {
		if _, err := strconv.Atoi(name[i+1:]); err == nil {
			return name[:i]
		}
	}
	return name
}

func isDeviceField(s string) bool {
	for _, f := range consts.DeviceFields {
		if string(f) == s {
			return true
		}
	}
	return false
}

// ListFiles 列出 CSV 文件, 支持按 platform/field 筛选
func (w *CSVWriter) ListFiles(platform, field, date string) ([]string, error) {
	w.Flush() // 先落盘, 保证文件大小/内容是最新的
	entries, err := os.ReadDir(w.baseDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".csv") {
			continue
		}
		name := e.Name()
		// 平台名精确匹配: 文件名 {平台}_{设备号类型}[_{分片}].csv, 前缀匹配会让 weibo 混进 weibo_direct
		if platform != "" {
			if rest, ok := strings.CutPrefix(prefixOfFile(name), platform+"_"); !ok || !isDeviceField(rest) {
				continue
			}
		}
		// 设备号类型也精确匹配: caid 不能混进 caid_md5
		if field != "" && !strings.HasSuffix(prefixOfFile(name), "_"+field) {
			continue
		}
		out = append(out, name)
	}
	return out, nil
}
