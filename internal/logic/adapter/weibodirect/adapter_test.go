package weibodirect

import (
	"net/url"
	"testing"
	"time"
)

func TestBuildQuery(t *testing.T) {
	// 宏替换后经 r.Get 解码得到的原始 markid
	mark := "4_reallog_mark_ad:8|4_11111111111111111111"
	q := buildQuery(mark, "2001", time.UnixMilli(1581645600000))
	if q != "behavior=2001&mark_id=4_reallog_mark_ad%3A8%7C4_11111111111111111111&time=1581645600000" {
		t.Fatalf("query = %s", q)
	}
	// 服务端 decode 一次应还原为原值 (只 encode 一次)
	v, _ := url.ParseQuery(q)
	if v.Get("mark_id") != mark {
		t.Fatalf("mark_id round trip = %q", v.Get("mark_id"))
	}
}

func TestIsOK(t *testing.T) {
	cases := map[string]bool{
		`{"code":0,"message":"OK"}`:                                    true,
		`{"code":10002,"message":"param time must be unix timestamp"}`: false,
		`{"message":"OK"}`:                                             false,
		``:                                                             false,
		`OK`:                                                           false,
	}
	for body, want := range cases {
		if got := isOK(body); got != want {
			t.Errorf("isOK(%q) = %v, want %v", body, got, want)
		}
	}
}
