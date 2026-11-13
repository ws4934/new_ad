package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.2.3", "v1.2.3", 0},
		{"v1.2.10", "v1.2.9", 1},
		{"v1.2.3", "v1.2.3.1", -1},
		{"v0.0.1", "dev", 1},
		{"dev", "v0.0.1", -1},
		{"v1.2", "v1.0.0", -1}, // 非法视为最小
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	h := "ABCDEF0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	sums := h + "  ad-monitor_v1.0.0_linux_amd64.tar.gz\n" + h[:63] + "0 *other\n"
	if got := checksumFor(sums, "ad-monitor_v1.0.0_linux_amd64.tar.gz"); got != "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" {
		t.Fatalf("got %q", got)
	}
	if checksumFor(sums, "missing") != "" {
		t.Fatal("missing should be empty")
	}
}

func TestExtract(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	add := func(name, body string) {
		_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg})
		_, _ = tw.Write([]byte(body))
	}
	add("README", "x")
	add("ad-monitor", "BINARY")
	tw.Close()
	gz.Close()

	dir := t.TempDir()
	p, err := extract(buf.Bytes(), dir)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	st, _ := os.Stat(p)
	if string(data) != "BINARY" || st.Mode().Perm() != 0o755 {
		t.Fatalf("data=%q mode=%v", data, st.Mode())
	}
}
