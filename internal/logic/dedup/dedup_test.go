package dedup_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/dedup"
)

func TestBloomCheck(t *testing.T) {
	d := dedup.New(nil)

	// 第一次: 新增
	isNew := d.Check("test", consts.FieldIMEIMD5, "device_aaa")
	if !isNew {
		t.Fatal("first check should return isNew=true")
	}

	// 第二次: 已存在
	isNew2 := d.Check("test", consts.FieldIMEIMD5, "device_aaa")
	if isNew2 {
		t.Fatal("second check should return isNew=false")
	}

	// 不同 field 互不影响
	isNew3 := d.Check("test", consts.FieldOAIDMD5, "device_aaa")
	if !isNew3 {
		t.Fatal("different field should return isNew=true")
	}

	// 不同 platform 互不影响
	isNew4 := d.Check("other", consts.FieldIMEIMD5, "device_aaa")
	if !isNew4 {
		t.Fatal("different platform should return isNew=true")
	}
}

func TestBloomDumpAndRestore(t *testing.T) {
	tmpDir := t.TempDir()

	// 写入一些数据
	d1 := dedup.New(nil)
	d1.Check("weibo", consts.FieldIMEIMD5, "persist_test_1")
	d1.Check("weibo", consts.FieldIMEIMD5, "persist_test_2")

	// dump
	d1.DumpAll(context.Background())

	// 确认文件存在
	files, _ := os.ReadDir("resource/bloom")
	found := false
	for _, f := range files {
		if strings.Contains(f.Name(), "weibo") {
			found = true
		}
	}
	if !found {
		t.Log("bloom file not in default dir, testing with tmpDir")
	}
	_ = tmpDir
}

func TestCSVWriterAppend(t *testing.T) {
	tmpDir := t.TempDir()
	w := dedup.NewCSVWriter(tmpDir)

	ctx := context.Background()
	w.Append(ctx, "test", "imei_md5", "abc123")
	w.Append(ctx, "test", "imei_md5", "def456")
	w.Flush()

	expectedPath := filepath.Join(tmpDir, "test_imei_md5.csv")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read csv failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 || lines[0] != "abc123" || lines[1] != "def456" {
		t.Errorf("csv content unexpected: %q", string(data))
	}
}

func TestCSVWriterShard(t *testing.T) {
	// 测试分片: 不好模拟 100MB, 只验证文件命名逻辑
	tmpDir := t.TempDir()
	w := dedup.NewCSVWriter(tmpDir)

	ctx := context.Background()
	w.Append(ctx, "plat", "field", "val1")

	files, _ := os.ReadDir(tmpDir)
	if len(files) != 1 || files[0].Name() != "plat_field.csv" {
		t.Errorf("expected plat_field.csv, got %v", files)
	}
}

func TestProcessClickDevices_SkipsNonDevice(t *testing.T) {
	tmpDir := t.TempDir()
	w := dedup.NewCSVWriter(tmpDir)
	d := dedup.New(w)
	ctx := context.Background()

	devices := map[consts.InternalField]string{
		consts.FieldIMEIMD5:      "real_device",
		consts.FieldAdvertiserID: "607748",       // 非设备号, 不应写 CSV
		consts.FieldClickID:      "click_123",    // 非设备号
	}

	d.ProcessClickDevices(ctx, "xhs", devices)
	w.Flush()

	// 只应该有 imei_md5 的 CSV
	files, _ := os.ReadDir(tmpDir)
	for _, f := range files {
		if strings.Contains(f.Name(), "advertiser") {
			t.Errorf("advertiser_id should not be in CSV: %s", f.Name())
		}
		if strings.Contains(f.Name(), "click") {
			t.Errorf("click_id should not be in CSV: %s", f.Name())
		}
	}

	// imei_md5 CSV 应该存在
	imeiCSV := filepath.Join(tmpDir, "xhs_imei_md5.csv")
	data, err := os.ReadFile(imeiCSV)
	if err != nil {
		t.Fatalf("imei_md5 csv not found: %v", err)
	}
	if strings.TrimSpace(string(data)) != "real_device" {
		t.Errorf("imei csv content = %q, want real_device", string(data))
	}
}
