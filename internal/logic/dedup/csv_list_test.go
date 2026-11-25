package dedup

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestListFiles_PlatformExact 平台筛选精确匹配: weibo 不能混进 weibo_direct 的文件
func TestListFiles_PlatformExact(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"weibo_imei_md5.csv", "weibo_imei_md5_2.csv", "weibo_direct_imei_md5.csv", "weibo_direct_caid_md5_3.csv", "淘客订单支付_imei_md5.csv"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w := NewCSVWriter(dir)
	cases := map[string][]string{
		"weibo":        {"weibo_imei_md5.csv", "weibo_imei_md5_2.csv"},
		"weibo_direct": {"weibo_direct_caid_md5_3.csv", "weibo_direct_imei_md5.csv"},
		"":             {"weibo_direct_caid_md5_3.csv", "weibo_direct_imei_md5.csv", "weibo_imei_md5.csv", "weibo_imei_md5_2.csv", "淘客订单支付_imei_md5.csv"},
	}
	for platform, want := range cases {
		got, err := w.ListFiles(platform, "", "")
		if err != nil {
			t.Fatal(err)
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("platform %q: got %v, want %v", platform, got, want)
		}
	}
}
