package dedup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
)

// guardedFilter 带互斥锁的可扩容 Bloom Filter, 保证并发安全
// dirty 标记自上次 dump 后是否有新增, 无新增则跳过写盘
type guardedFilter struct {
	mu    sync.Mutex
	sf    *scalableFilter
	dirty bool
}

// Dedup 设备号去重 + CSV 追加
// 使用进程内可扩容 Bloom Filter (Scalable Bloom Filter):
// 基础层写满后自动追加「容量×2, 误判率×0.5」的新层, 总误判率有界 (≤ 2×基础误判率),
// 内存按实际设备量增长。定期 dump 到本地文件, 启动时 load 恢复。
type Dedup struct {
	writer        *CSVWriter
	filters       sync.Map // map[filterKey]*guardedFilter, key = "{platform}:{field}"
	mu            sync.Mutex
	dataDir       string // bloom filter 持久化目录
	expectedItems uint
	falsePositive float64
}

// 默认基础层容量参数。自动扩容后这只是起点:
// 200 万/0.01% ≈ 5MB, 写满自动翻倍扩容, 误判率始终 ≤ 2×0.01%。
const (
	defaultBloomExpectedItems = 2_000_000 // 基础层预期元素数
	defaultBloomFalsePositive = 0.0001    // 基础层误判率 0.01%
)

// New 构造 Dedup
// 容量可通过配置覆盖: adMonitor.dedup.bloomExpectedItems / bloomFalsePositive
func New(writer *CSVWriter) *Dedup {
	d := &Dedup{
		writer:        writer,
		dataDir:       "resource/bloom", // bloom filter 持久化目录
		expectedItems: defaultBloomExpectedItems,
		falsePositive: defaultBloomFalsePositive,
	}
	ctx := context.Background()
	if v, err := g.Cfg().Get(ctx, "adMonitor.dedup.bloomExpectedItems"); err == nil && v.Uint() > 0 {
		d.expectedItems = v.Uint()
	}
	if v, err := g.Cfg().Get(ctx, "adMonitor.dedup.bloomFalsePositive"); err == nil && v.Float64() > 0 {
		d.falsePositive = v.Float64()
	}
	_ = os.MkdirAll(d.dataDir, 0755)
	return d
}

// getFilter 获取或创建指定 platform+field 的 guardedFilter
// 首次创建时尝试从本地文件恢复
func (d *Dedup) getFilter(platform string, field consts.InternalField) *guardedFilter {
	key := platform + ":" + string(field)
	if v, ok := d.filters.Load(key); ok {
		return v.(*guardedFilter)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// 二次检查
	if v, ok := d.filters.Load(key); ok {
		return v.(*guardedFilter)
	}

	// 尝试从文件恢复
	sf := d.loadFromFile(key)
	if sf == nil {
		sf = newScalableFilter(d.expectedItems, d.falsePositive)
	}
	gf := &guardedFilter{sf: sf}
	d.filters.Store(key, gf)
	return gf
}

// Check 对单个设备号字段做去重, 返回是否新增
// 加锁保证并发安全, TestAndAdd: 如果之前不存在返回 false (= 新增)
func (d *Dedup) Check(platform string, field consts.InternalField, value string) bool {
	gf := d.getFilter(platform, field)
	gf.mu.Lock()
	layersBefore := gf.sf.Layers()
	existed := gf.sf.TestAndAdd(value)
	if !existed {
		gf.dirty = true
	}
	grown := gf.sf.Layers() > layersBefore
	count := gf.sf.Count()
	gf.mu.Unlock()
	if grown {
		g.Log().Infof(context.Background(),
			"bloom filter %s:%s grew to %d layers (items=%d)",
			platform, field, layersBefore+1, count)
	}
	return !existed // 返回 true = 新增
}

// ProcessClickDevices 监测入口调用的统一入口
// 只处理 DeviceFields 列表里的字段 (非设备号字段如 advertiser_id 不参与去重)
func (d *Dedup) ProcessClickDevices(ctx context.Context, platform string, devices map[consts.InternalField]string) {
	for _, field := range consts.DeviceFields {
		value, ok := devices[field]
		if !ok || value == "" {
			continue
		}
		isNew := d.Check(platform, field, value)
		if isNew && d.writer != nil {
			d.writer.Append(ctx, platform, string(field), value)
		}
	}
}

// DumpAll 持久化所有 Bloom Filter 到本地文件
// 由 gcron 每 5 分钟调用 + 进程退出时由 ShutdownPlugin 调用
// 自上次 dump 后无新增的 filter 跳过写盘 (全量序列化, 体积随层数增长)
//
// 锁内只做内存序列化 (memcpy 级, 毫秒内), 磁盘 IO 放在锁外 —
// 避免持锁写盘几百 ms 期间阻塞热路径的 Check, 导致点击处理 goroutine 堆积。
func (d *Dedup) DumpAll(ctx context.Context) {
	dumped := 0
	d.filters.Range(func(k, v any) bool {
		key := k.(string)
		gf := v.(*guardedFilter)

		gf.mu.Lock()
		if !gf.dirty {
			gf.mu.Unlock()
			return true
		}
		var buf bytes.Buffer
		err := gf.sf.writeTo(&buf)
		if err == nil {
			gf.dirty = false
		}
		gf.mu.Unlock()

		if err != nil {
			g.Log().Errorf(ctx, "serialize bloom filter %s failed: %v", key, err)
			return true
		}
		if err := d.writeFileAtomic(key, buf.Bytes()); err != nil {
			g.Log().Errorf(ctx, "dump bloom filter %s failed: %v", key, err)
			// 写盘失败重新置脏, 下个周期重试
			gf.mu.Lock()
			gf.dirty = true
			gf.mu.Unlock()
			return true
		}
		dumped++
		return true
	})
	g.Log().Debugf(ctx, "bloom filters dumped: %d changed", dumped)
}

// writeFileAtomic 临时文件 + 原子改名落盘
// 临时文件名唯一 (CreateTemp): cron 的 DumpAll 和退出时 ShutdownPlugin 的 DumpAll
// 可能并发执行, 固定 tmp 名会互相写花
func (d *Dedup) writeFileAtomic(key string, data []byte) error {
	path := filepath.Join(d.dataDir, key+".bloom")
	f, err := os.CreateTemp(d.dataDir, key+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// loadFromFile 从文件恢复可扩容 bloom filter (兼容旧版单 filter 格式)
func (d *Dedup) loadFromFile(key string) *scalableFilter {
	path := filepath.Join(d.dataDir, key+".bloom")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	sf, err := readScalableFilter(f, d.falsePositive)
	if err != nil {
		g.Log().Warningf(context.Background(), "load bloom filter %s failed, starting fresh: %v", key, err)
		return nil
	}
	return sf
}
