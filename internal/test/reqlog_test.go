package test_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/stats"

	"github.com/gogf/gf/v2/frame/g"
)

// TestIntegration_RequestLog 每账户只保留最新 50 条, 表头最新; 监测入口会记录原始 query
func TestIntegration_RequestLog(t *testing.T) {
	ctx := context.Background()
	const aid = 987655
	key := consts.KeyReqLog(aid)
	g.Redis().Del(ctx, key)
	defer g.Redis().Del(ctx, key)

	// 两批写入 (跨 flush), 共 70 条
	for i := 0; i < 30; i++ {
		stats.RecordRequest(aid, fmt.Sprintf("q=%d", i))
	}
	stats.FlushRequestLog(ctx)
	for i := 30; i < 70; i++ {
		stats.RecordRequest(aid, fmt.Sprintf("q=%d", i))
	}
	stats.FlushRequestLog(ctx)

	lines, err := g.Redis().LRange(ctx, key, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if ttl, _ := g.Redis().TTL(ctx, key); ttl <= 0 {
		t.Fatalf("request log has no TTL: %d", ttl)
	}
	if len(lines) != 50 || !strings.HasSuffix(lines[0].String(), " q=69") || !strings.HasSuffix(lines[49].String(), " q=20") {
		t.Fatalf("len=%d first=%q last=%q", len(lines), lines[0], lines[len(lines)-1])
	}

	// 监测入口: 真实请求的原始 query 被记录
	s := setupTestServer(t)
	defer s.Shutdown()
	real := getTestAccountID(ctx, t)
	realKey := consts.KeyReqLog(real)
	g.Redis().Del(ctx, realKey)
	defer g.Redis().Del(ctx, realKey)
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/m/%d?IMP=REQLOG_1&imei_md5=__IMEI__", s.GetListenedPort(), real))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	stats.FlushRequestLog(ctx)
	v, _ := g.Redis().LIndex(ctx, realKey, 0)
	if !strings.HasSuffix(v.String(), " IMP=REQLOG_1&imei_md5=__IMEI__") {
		t.Fatalf("monitor not logged: %q", v.String())
	}
	t.Logf("sample: %s", v.String())
}
