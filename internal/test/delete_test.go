package test_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/biz"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/model/do"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func TestIntegration_ManagementDelete(t *testing.T) {
	ctx := context.Background()
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	count := func(m *gdb.Model, want int) {
		t.Helper()
		n, err := m.Count()
		check(err)
		if n != want {
			t.Fatalf("count = %d, want %d", n, want)
		}
	}
	fixture := func() (int, int) {
		t.Helper()
		pid, err := dao.AdPlatform.Ctx(ctx).Data(do.AdPlatform{Name: fmt.Sprintf("delete_%d", time.Now().UnixNano()), DisplayName: "delete test"}).InsertAndGetId()
		check(err)
		t.Cleanup(func() {
			if err := biz.PlatformDelete(ctx, int(pid)); err != nil {
				t.Error(err)
			}
		})
		aid, err := dao.AdAccount.Ctx(ctx).Data(do.AdAccount{PlatformId: pid, Name: "delete test"}).InsertAndGetId()
		check(err)
		_, err = dao.ActionMapping.Ctx(ctx).Data(do.ActionMapping{AccountId: aid, TaolianType: 5, PlatformType: "1"}).Insert()
		check(err)
		_, err = dao.ReportStats.Ctx(ctx).Data(do.ReportStats{AccountId: aid, PlatformId: pid, StatDate: gtime.New("2026-01-01"), ClickCount: 7}).Insert()
		check(err)
		_, err = dao.PlatformCsvStats.Ctx(ctx).Data(do.PlatformCsvStats{PlatformId: pid, StatDate: gtime.New("2026-01-01")}).Insert()
		check(err)
		_, err = cache.GetAccount(ctx, int(aid))
		check(err)
		_, err = cache.GetPlatform(ctx, int(pid))
		check(err)
		_, err = cache.GetMappings(ctx, int(aid))
		check(err)
		return int(pid), int(aid)
	}
	rollbackPID, rollbackAID := fixture()
	rollback := errors.New("force rollback")
	err := dao.AdPlatform.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := biz.PlatformDelete(ctx, rollbackPID); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("expected rollback error, got %v", err)
	}
	count(dao.AdPlatform.Ctx(ctx).WherePri(rollbackPID), 1)
	count(dao.AdAccount.Ctx(ctx).WherePri(rollbackAID), 1)
	count(dao.ActionMapping.Ctx(ctx).Where("account_id", rollbackAID), 1)
	count(dao.ReportStats.Ctx(ctx).Where("account_id", rollbackAID), 1)
	count(dao.PlatformCsvStats.Ctx(ctx).Where("platform_id", rollbackPID), 1)
	pid, aid := fixture()
	otherPID, otherAID := fixture()
	_, _ = g.Redis().LPush(ctx, consts.KeyReqLog(aid), "x")
	_, _ = g.Redis().LPush(ctx, consts.KeyCbLog(aid), "x")
	check(biz.AccountDelete(ctx, aid))
	check(biz.AccountDelete(ctx, aid))
	if n, _ := g.Redis().Exists(ctx, consts.KeyReqLog(aid), consts.KeyCbLog(aid)); n != 0 {
		t.Fatal("request/callback log remains after account delete")
	}
	count(dao.AdAccount.Ctx(ctx).WherePri(aid), 0)
	count(dao.ActionMapping.Ctx(ctx).Where("account_id", aid), 0)
	count(dao.ReportStats.Ctx(ctx).Where("account_id", aid), 0)
	count(dao.AdPlatform.Ctx(ctx).WherePri(pid), 1)
	count(dao.PlatformCsvStats.Ctx(ctx).Where("platform_id", pid), 1)
	for _, key := range []string{consts.KeyCacheAccount(aid), consts.KeyCacheMappings(aid)} {
		v, err := g.Redis().Get(ctx, key)
		check(err)
		if !v.IsNil() {
			t.Fatalf("cache remains: %s", key)
		}
	}
	if _, err := cache.GetAccount(ctx, aid); err == nil {
		t.Fatal("deleted account still readable")
	}
	check(biz.PlatformDelete(ctx, otherPID))
	check(biz.PlatformDelete(ctx, otherPID))
	count(dao.AdPlatform.Ctx(ctx).WherePri(otherPID), 0)
	count(dao.AdAccount.Ctx(ctx).WherePri(otherAID), 0)
	count(dao.ActionMapping.Ctx(ctx).Where("account_id", otherAID), 0)
	count(dao.ReportStats.Ctx(ctx).Where("platform_id", otherPID), 0)
	count(dao.PlatformCsvStats.Ctx(ctx).Where("platform_id", otherPID), 0)
	count(dao.AdPlatform.Ctx(ctx).WherePri(pid), 1)
	for _, key := range []string{consts.KeyCachePlatform(otherPID), consts.KeyCacheAccount(otherAID), consts.KeyCacheMappings(otherAID)} {
		v, err := g.Redis().Get(ctx, key)
		check(err)
		if !v.IsNil() {
			t.Fatalf("cache remains: %s", key)
		}
	}
	if err := biz.AccountDelete(ctx, 0); err == nil {
		t.Fatal("accepted invalid account ID")
	}
	if err := biz.PlatformDelete(ctx, -1); err == nil {
		t.Fatal("accepted invalid platform ID")
	}
}
