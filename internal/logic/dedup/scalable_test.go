package dedup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	bloomfilter "github.com/bits-and-blooms/bloom/v3"
)

// ============ scalableFilter 基础行为 ============

func TestScalable_BasicAddAndTest(t *testing.T) {
	sf := newScalableFilter(1000, 0.01)

	if existed := sf.TestAndAdd("device_a"); existed {
		t.Fatal("first add should return existed=false")
	}
	if existed := sf.TestAndAdd("device_a"); !existed {
		t.Fatal("second add of same value should return existed=true")
	}
	if existed := sf.TestAndAdd("device_b"); existed {
		t.Fatal("different value should return existed=false")
	}
	if sf.Layers() != 1 {
		t.Fatalf("layers = %d, want 1 (far below capacity)", sf.Layers())
	}
	if sf.Count() != 2 {
		t.Fatalf("count = %d, want 2", sf.Count())
	}
}

func TestScalable_AutoGrow(t *testing.T) {
	const baseN = 100
	sf := newScalableFilter(baseN, 0.01)

	// 写入 10 倍基础容量
	const total = 1000
	for i := 0; i < total; i++ {
		sf.TestAndAdd(fmt.Sprintf("grow_item_%d", i))
	}

	// 必须发生过扩容
	if sf.Layers() < 2 {
		t.Fatalf("layers = %d, want >= 2 after 10x overflow", sf.Layers())
	}

	// 层容量按 2 倍递增
	for i := 1; i < len(sf.caps); i++ {
		if sf.caps[i] != sf.caps[i-1]*sbfGrowthFactor {
			t.Errorf("caps[%d] = %d, want %d (x%d growth)", i, sf.caps[i], sf.caps[i-1]*sbfGrowthFactor, sbfGrowthFactor)
		}
	}

	// 每层误判率按 0.5 倍递减
	if got, want := sf.layerP(1), 0.01*sbfTighteningRate; got != want {
		t.Errorf("layerP(1) = %v, want %v", got, want)
	}

	// 扩容触发条件: 除最后一层外, 每层写入数应达到容量
	for i := 0; i < len(sf.counts)-1; i++ {
		if sf.counts[i] < sf.caps[i] {
			t.Errorf("layer %d count = %d < cap %d, should be full before growing", i, sf.counts[i], sf.caps[i])
		}
	}
}

func TestScalable_NoFalseNegatives(t *testing.T) {
	// Bloom Filter 的核心保证: 加过的元素一定能查到 (跨层也成立)
	sf := newScalableFilter(50, 0.01)

	const total = 2000
	for i := 0; i < total; i++ {
		sf.TestAndAdd(fmt.Sprintf("fn_item_%d", i))
	}
	if sf.Layers() < 3 {
		t.Fatalf("expect >=3 layers for this test, got %d", sf.Layers())
	}

	// 全部重查: 必须 100% 命中
	for i := 0; i < total; i++ {
		if existed := sf.TestAndAdd(fmt.Sprintf("fn_item_%d", i)); !existed {
			t.Fatalf("item %d added but reported as new (false negative!)", i)
		}
	}
}

func TestScalable_FalsePositiveRateBounded(t *testing.T) {
	// 写入 10 倍于基础容量后, 误判率仍应有界 (理论上 ≤ 2×baseP)
	// 对比: 固定容量 filter 在 10 倍超量时误判率会接近 100%
	const baseN = 1000
	const baseP = 0.01
	sf := newScalableFilter(baseN, baseP)

	for i := 0; i < baseN*10; i++ {
		sf.TestAndAdd(fmt.Sprintf("added_%d", i))
	}

	// 用从未加过的值测误判率
	const probes = 20000
	falsePositives := 0
	for i := 0; i < probes; i++ {
		data := []byte(fmt.Sprintf("never_added_%d", i))
		hit := false
		for _, bf := range sf.layers {
			if bf.Test(data) {
				hit = true
				break
			}
		}
		if hit {
			falsePositives++
		}
	}

	rate := float64(falsePositives) / float64(probes)
	theoretical := baseP / (1 - sbfTighteningRate) // = 2×baseP = 2%
	if rate > theoretical*2 {                      // 留 2 倍余量防偶发抖动
		t.Errorf("false positive rate = %.4f, want <= %.4f (theoretical bound %.4f)",
			rate, theoretical*2, theoretical)
	}
	t.Logf("10x overflow: layers=%d fpRate=%.4f (theoretical bound %.4f)", sf.Layers(), rate, theoretical)
}

// ============ 序列化 / 反序列化 ============

func TestScalable_SerializeRoundtrip(t *testing.T) {
	sf := newScalableFilter(100, 0.01)
	const total = 500 // 触发多层
	for i := 0; i < total; i++ {
		sf.TestAndAdd(fmt.Sprintf("ser_item_%d", i))
	}

	var buf bytes.Buffer
	if err := sf.writeTo(&buf); err != nil {
		t.Fatalf("writeTo failed: %v", err)
	}

	restored, err := readScalableFilter(&buf, 0.01)
	if err != nil {
		t.Fatalf("readScalableFilter failed: %v", err)
	}

	// 元数据一致
	if restored.Layers() != sf.Layers() {
		t.Fatalf("restored layers = %d, want %d", restored.Layers(), sf.Layers())
	}
	if restored.Count() != sf.Count() {
		t.Fatalf("restored count = %d, want %d", restored.Count(), sf.Count())
	}
	if restored.baseP != sf.baseP {
		t.Fatalf("restored baseP = %v, want %v", restored.baseP, sf.baseP)
	}
	for i := range sf.caps {
		if restored.caps[i] != sf.caps[i] || restored.counts[i] != sf.counts[i] {
			t.Fatalf("layer %d meta mismatch: caps %d/%d counts %d/%d",
				i, restored.caps[i], sf.caps[i], restored.counts[i], sf.counts[i])
		}
	}

	// 已加元素全部命中 (无假阴性)
	for i := 0; i < total; i++ {
		if existed := restored.TestAndAdd(fmt.Sprintf("ser_item_%d", i)); !existed {
			t.Fatalf("item %d lost after roundtrip", i)
		}
	}

	// 恢复后能继续写入并扩容
	layersBefore := restored.Layers()
	for i := 0; i < 1000; i++ {
		restored.TestAndAdd(fmt.Sprintf("post_restore_%d", i))
	}
	if restored.Layers() <= layersBefore {
		t.Errorf("restored filter should keep growing: layers %d -> %d", layersBefore, restored.Layers())
	}
}

func TestScalable_LegacyFormatCompat(t *testing.T) {
	// 旧版格式: 单个裸 BloomFilter 直接 WriteTo
	legacy := bloomfilter.NewWithEstimates(1000, 0.01)
	const total = 800
	for i := 0; i < total; i++ {
		legacy.AddString(fmt.Sprintf("legacy_item_%d", i))
	}
	var buf bytes.Buffer
	if _, err := legacy.WriteTo(&buf); err != nil {
		t.Fatalf("legacy WriteTo failed: %v", err)
	}

	sf, err := readScalableFilter(&buf, 0.01)
	if err != nil {
		t.Fatalf("read legacy format failed: %v", err)
	}
	if sf.Layers() != 1 {
		t.Fatalf("legacy should load as single layer, got %d", sf.Layers())
	}

	// 旧数据全部命中
	for i := 0; i < total; i++ {
		if existed := sf.TestAndAdd(fmt.Sprintf("legacy_item_%d", i)); !existed {
			t.Fatalf("legacy item %d lost", i)
		}
	}

	// 容量反推应接近原 1000 (m/k×ln2), 继续写满后能扩容
	if sf.caps[0] < 500 || sf.caps[0] > 2000 {
		t.Errorf("inferred legacy capacity = %d, want ~1000", sf.caps[0])
	}
	for i := 0; i < 2000; i++ {
		sf.TestAndAdd(fmt.Sprintf("legacy_overflow_%d", i))
	}
	if sf.Layers() < 2 {
		t.Errorf("legacy filter should grow after overflow, layers = %d", sf.Layers())
	}
}

func TestScalable_CorruptedMetadata(t *testing.T) {
	// magic 完好但元数据非法 (baseP/cap 病态值) 也必须返回 error
	sf := newScalableFilter(100, 0.01)
	sf.TestAndAdd("meta_item")
	var buf bytes.Buffer
	if err := sf.writeTo(&buf); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()

	// baseP 在 magic(4B) 之后 8 字节, 篡改为 0
	badP := append([]byte{}, raw...)
	for i := 4; i < 12; i++ {
		badP[i] = 0
	}
	if _, err := readScalableFilter(bytes.NewReader(badP), 0.01); err == nil {
		t.Error("baseP=0 should be rejected")
	}

	// 层 cap 在 magic(4)+baseP(8)+layers(4)+count(8) 之后 8 字节, 篡改为 0
	badCap := append([]byte{}, raw...)
	for i := 24; i < 32; i++ {
		badCap[i] = 0
	}
	if _, err := readScalableFilter(bytes.NewReader(badCap), 0.01); err == nil {
		t.Error("cap=0 should be rejected")
	}
}

func TestScalable_CorruptedFile(t *testing.T) {
	// 损坏数据应返回 error 而不是 panic
	cases := map[string][]byte{
		"empty":        {},
		"garbage":      []byte("not a bloom filter at all"),
		"magic only":   {0x53, 0x42, 0x46, 0x31},
		"bad layercnt": append([]byte{0x53, 0x42, 0x46, 0x31}, make([]byte, 12)...), // baseP=0, layers=0
	}
	for name, data := range cases {
		if _, err := readScalableFilter(bytes.NewReader(data), 0.01); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

// ============ CSVWriter 删除活跃分片 ============

func TestCSVWriter_RemoveActiveShard(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewCSVWriter(tmpDir)
	ctx := context.Background()

	w.Append(ctx, "weibo", "imei_md5", "before_delete_1")
	w.Append(ctx, "weibo", "imei_md5", "before_delete_2")
	w.Flush()

	path := filepath.Join(tmpDir, "weibo_imei_md5.csv")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("csv should exist before delete: %v", err)
	}

	// 删除活跃分片
	if err := w.RemoveFile("weibo_imei_md5.csv"); err != nil {
		t.Fatalf("RemoveFile failed: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("file should be deleted")
	}

	// 关键: 删除后继续写入, 新数据必须落到磁盘上的新文件 (而不是已 unlink 的旧 inode)
	w.Append(ctx, "weibo", "imei_md5", "after_delete_1")
	w.Flush()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file should be recreated after delete+append: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "after_delete_1" {
		t.Errorf("recreated file content = %q, want only after_delete_1", content)
	}
}

func TestCSVWriter_RemoveFile_BufferDiscarded(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewCSVWriter(tmpDir)
	ctx := context.Background()

	// 写入但不 flush: 数据还在缓冲里
	w.Append(ctx, "weibo", "oaid_md5", "buffered_value")
	if err := w.RemoveFile("weibo_oaid_md5.csv"); err != nil {
		t.Fatalf("RemoveFile failed: %v", err)
	}

	// 删除后 Flush 不应 panic, 也不应把旧缓冲写到任何地方
	w.Flush()
	path := filepath.Join(tmpDir, "weibo_oaid_md5.csv")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("discarded buffer should not recreate the file")
	}
}

func TestCSVWriter_RemoveNonActiveFile(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewCSVWriter(tmpDir)
	ctx := context.Background()

	// 模拟历史状态: 主文件是已写满的旧分片, _1 是最新分片
	// findLatestShard 会让 writer 活跃在 _1 上, 主文件即非活跃文件
	if err := os.WriteFile(filepath.Join(tmpDir, "weibo_imei_md5.csv"), []byte("old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "weibo_imei_md5_1.csv"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	w.Append(ctx, "weibo", "imei_md5", "active_value")

	// 删除非活跃的主文件, 活跃分片 _1 不受影响
	if err := w.RemoveFile("weibo_imei_md5.csv"); err != nil {
		t.Fatalf("remove old shard failed: %v", err)
	}
	w.Flush()
	data, err := os.ReadFile(filepath.Join(tmpDir, "weibo_imei_md5_1.csv"))
	if err != nil || strings.TrimSpace(string(data)) != "active_value" {
		t.Errorf("active shard should be untouched, content=%q err=%v", string(data), err)
	}

	// 删除不存在的文件返回 error
	if err := w.RemoveFile("not_exist.csv"); err == nil {
		t.Error("removing nonexistent file should return error")
	}
}

func TestPrefixOfFile(t *testing.T) {
	cases := map[string]string{
		"weibo_imei_md5.csv":   "weibo_imei_md5",
		"weibo_imei_md5_1.csv": "weibo_imei_md5",
		"weibo_imei_md5_12.csv": "weibo_imei_md5",
		"xhs_caid_md5.csv":     "xhs_caid_md5",
		"plain.csv":            "plain",
	}
	for in, want := range cases {
		if got := prefixOfFile(in); got != want {
			t.Errorf("prefixOfFile(%q) = %q, want %q", in, got, want)
		}
	}
}

// ============ Dedup 集成 (dump / restore / dirty) ============

func newTestDedup(t *testing.T, baseN uint) *Dedup {
	t.Helper()
	d := New(nil)
	d.dataDir = t.TempDir()
	d.expectedItems = baseN
	d.falsePositive = 0.01
	return d
}

func TestDedup_DumpRestore_WithGrowth(t *testing.T) {
	d1 := New(nil)
	dir := t.TempDir()
	d1.dataDir = dir
	d1.expectedItems = 100
	d1.falsePositive = 0.01

	// 写入 5 倍容量, 触发扩容
	// 注意: 不逐个断言 isNew — bloom 固有误判会让极少数新值被判已存在, 只做统计断言
	const total = 500
	newCount := 0
	for i := 0; i < total; i++ {
		if d1.Check("weibo", "imei_md5", fmt.Sprintf("dump_dev_%d", i)) {
			newCount++
		}
	}
	if newCount < total*95/100 {
		t.Fatalf("newCount = %d, want >= %d (too many false positives on insert)", newCount, total*95/100)
	}
	d1.DumpAll(context.Background())

	// 新实例从同一目录恢复: 加过的设备必须全部命中 (无假阴性, 严格断言)
	d2 := New(nil)
	d2.dataDir = dir
	d2.expectedItems = 100
	d2.falsePositive = 0.01
	for i := 0; i < total; i++ {
		if isNew := d2.Check("weibo", "imei_md5", fmt.Sprintf("dump_dev_%d", i)); isNew {
			t.Fatalf("device %d should exist after restore (false negative!)", i)
		}
	}
}

func TestDedup_DirtyFlag_SkipsCleanDump(t *testing.T) {
	d := newTestDedup(t, 1000)
	ctx := context.Background()

	d.Check("weibo", "imei_md5", "dirty_dev_1")
	d.DumpAll(ctx)

	path := filepath.Join(d.dataDir, "weibo:imei_md5.bloom")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("bloom file should exist after first dump: %v", err)
	}

	// 删掉文件后再 DumpAll: 无新增 → 跳过写盘, 文件不应重建
	_ = os.Remove(path)
	d.DumpAll(ctx)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("clean filter should skip dump (file should not be recreated)")
	}

	// 有新增后再 DumpAll: 文件应重建
	d.Check("weibo", "imei_md5", "dirty_dev_2")
	d.DumpAll(ctx)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("dirty filter should be dumped: %v", err)
	}
}

func TestDedup_AtomicDump_NoTmpLeftover(t *testing.T) {
	d := newTestDedup(t, 1000)
	d.Check("weibo", "oaid_md5", "atomic_dev")
	d.DumpAll(context.Background())

	entries, _ := os.ReadDir(d.dataDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("tmp file leftover: %s", e.Name())
		}
	}
}

func TestDedup_LoadCorruptedFile_StartsFresh(t *testing.T) {
	d := newTestDedup(t, 1000)
	path := filepath.Join(d.dataDir, "weibo:imei_md5.bloom")
	if err := os.WriteFile(path, []byte("corrupted"), 0644); err != nil {
		t.Fatal(err)
	}

	// 损坏文件应被忽略, 当作全新 filter, 不 panic
	if isNew := d.Check("weibo", "imei_md5", "fresh_dev"); !isNew {
		t.Fatal("corrupted file should be discarded, device should be new")
	}
}

func TestDedup_DumpFailure_RetainsDirty(t *testing.T) {
	d := newTestDedup(t, 1000)
	ctx := context.Background()

	d.Check("weibo", "imei_md5", "retry_dev")

	// 让写盘失败: dataDir 指向一个普通文件
	goodDir := d.dataDir
	badDir := filepath.Join(t.TempDir(), "not_a_dir")
	if err := os.WriteFile(badDir, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	d.dataDir = badDir
	d.DumpAll(ctx) // 失败, 应保留脏标记

	// 恢复目录后再 dump: 脏标记还在 → 文件应写出
	d.dataDir = goodDir
	d.DumpAll(ctx)
	if _, err := os.Stat(filepath.Join(goodDir, "weibo:imei_md5.bloom")); err != nil {
		t.Fatalf("dirty flag should survive a failed dump, file not written: %v", err)
	}
}

func TestDedup_DumpAllConcurrentWithCheck(t *testing.T) {
	// DumpAll (锁内序列化+锁外写盘) 与热路径 Check 并发执行不竞态 (配合 -race)
	d := newTestDedup(t, 200)
	ctx := context.Background()

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
					d.Check("weibo", "imei_md5", fmt.Sprintf("dump_race_g%d_%d", gid, i))
				}
			}
		}(g)
	}
	// 并发跑多轮 DumpAll (模拟 cron 与 shutdown 撞车)
	var dumpWg sync.WaitGroup
	for i := 0; i < 2; i++ {
		dumpWg.Add(1)
		go func() {
			defer dumpWg.Done()
			for j := 0; j < 10; j++ {
				d.DumpAll(ctx)
			}
		}()
	}
	dumpWg.Wait()
	close(stop)
	wg.Wait()

	// dump 后无 tmp 残留
	entries, _ := os.ReadDir(d.dataDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("tmp leftover after concurrent dumps: %s", e.Name())
		}
	}
}

func TestDedup_ConcurrentCheck(t *testing.T) {
	// 并发去重正确性: N 个 goroutine 同时 Check 重叠的设备号集合,
	// 每个设备号应恰好被判定为"新增"一次 (互斥锁保证 TestAndAdd 原子)
	d := newTestDedup(t, 10000)

	const goroutines = 8
	const devices = 5000

	var newCount int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := 0
			for i := 0; i < devices; i++ {
				if d.Check("xhs", "imei_md5", fmt.Sprintf("conc_dev_%d", i)) {
					local++
				}
			}
			mu.Lock()
			newCount += int64(local)
			mu.Unlock()
		}()
	}
	wg.Wait()

	// 每个设备号全局恰好新增一次; 允许极少量 bloom 误判导致的"漏新增"
	if newCount > devices {
		t.Fatalf("newCount = %d > %d, duplicate 'new' verdicts (race!)", newCount, devices)
	}
	if newCount < devices-5 {
		t.Fatalf("newCount = %d, want ~%d (too many false positives)", newCount, devices)
	}
}

func TestDedup_ConcurrentCheckWithGrowth(t *testing.T) {
	// 并发写入触发扩容的同时不能丢数据/竞态 (配合 -race 跑)
	d := newTestDedup(t, 100) // 小容量, 并发中必然扩容多次

	const goroutines = 8
	const devices = 2000

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < devices; i++ {
				// 每个 goroutine 写自己的号段, 无重叠
				d.Check("weibo", "caid_md5", fmt.Sprintf("g%d_dev_%d", gid, i))
			}
		}(g)
	}
	wg.Wait()

	// 全部重查必须命中 (无假阴性)
	for g := 0; g < goroutines; g++ {
		for i := 0; i < devices; i++ {
			if isNew := d.Check("weibo", "caid_md5", fmt.Sprintf("g%d_dev_%d", g, i)); isNew {
				t.Fatalf("g%d_dev_%d lost after concurrent growth", g, i)
			}
		}
	}
}
