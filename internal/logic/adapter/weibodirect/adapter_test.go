package weibodirect

import (
	"encoding/json"
	"math/rand/v2"
	"net/url"
	"strconv"
	"testing"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/model/bo"
)

// 宏替换后经 r.Get 解码得到的原始 markid
const mark = "4_reallog_mark_ad:8|4_11111111111111111111"

func orderOf(t *testing.T, q string) (url.Values, itemOrderPay) {
	t.Helper()
	v, err := url.ParseQuery(q)
	if err != nil {
		t.Fatal(err)
	}
	var o itemOrderPay
	if s := v.Get("item_order_pay"); s != "" {
		if err := json.Unmarshal([]byte(s), &o); err != nil {
			t.Fatalf("item_order_pay: %v", err)
		}
	}
	return v, o
}

func TestBuildQuery(t *testing.T) {
	p := &bo.CallbackPayload{
		ClickID: mark, MappedType: "2001", ConvTS: 1581645600000,
		Devices: map[consts.InternalField]string{consts.FieldTS: "1581645000000"},
	}
	c := credentials{ServiceProvider: "某服务商", TechnicalProvider: "某技术商"}
	now := time.UnixMilli(1581645600500)

	// 服务端 decode 一次应还原为原值 (只 encode 一次)
	v, o := orderOf(t, buildQuery(p, "cb.example.com", c, now))
	want := map[string]string{
		"time": "1581645600500", "behavior": "2001", "mark_id": mark,
		"host": "cb.example.com", "technical_provider": "某技术商",
	}
	for k, w := range want {
		if v.Get(k) != w {
			t.Errorf("%s = %q, want %q", k, v.Get(k), w)
		}
	}
	if o.ServiceProvider != "某服务商" || o.OrderStatus != "已付款" ||
		o.OrderPayTime != "1581645600000" || o.ClickTime != "1581645000000" {
		t.Errorf("order = %+v", o)
	}

	// 重试: 调用时间变了, time 跟着变, 订单不变
	v2, o2 := orderOf(t, buildQuery(p, "cb.example.com", c, now.Add(time.Minute)))
	if v2.Get("time") == v.Get("time") || o2 != o {
		t.Errorf("retry: time %s→%s, order %+v → %+v", v.Get("time"), v2.Get("time"), o, o2)
	}
	// 另一笔转化 (转化时间不同) 订单号不同
	p2 := *p
	p2.ConvTS++
	if _, o3 := orderOf(t, buildQuery(&p2, "", c, now)); o3.OrderID == o.OrderID {
		t.Error("different conversion got same order_id")
	}

	// 非 2001 不带订单, 空 host/technical_provider 不传
	v, _ = orderOf(t, buildQuery(&bo.CallbackPayload{ClickID: mark, MappedType: "1"}, "", credentials{}, now))
	for _, k := range []string{"item_order_pay", "host", "technical_provider"} {
		if v.Has(k) {
			t.Errorf("unexpected %s", k)
		}
	}
}

func TestGenOrder(t *testing.T) {
	cfg := OrderConfig{ItemNames: []string{"A", "B"}, Categories: []string{"C"}, PriceMin: 10, PriceMax: 20, QtyMin: 1, QtyMax: 3}
	pay := time.Now().UnixMilli()
	r := rand.New(rand.NewPCG(42, 0))
	ts := func(s string) int64 { n, _ := strconv.ParseInt(s, 10, 64); return n }
	ordered := func(o itemOrderPay) bool {
		return ts(o.ClickTime) < ts(o.OrderCreateTime) && ts(o.OrderCreateTime) < ts(o.OrderPayClickTime) &&
			ts(o.OrderPayClickTime) < ts(o.OrderPayTime) && ts(o.OrderPayTime) == pay
	}
	for i := 0; i < 1000; i++ {
		o := genOrder(cfg, "sp", 0, pay, r)
		if o.ItemPrice < 10 || o.ItemPrice > 20 || o.QuantityOfItem < 1 || o.QuantityOfItem > 3 {
			t.Fatalf("out of range: %+v", o)
		}
		if o.PaymentAmount != float64(int64(o.ItemPrice*float64(o.QuantityOfItem)*100+0.5))/100 {
			t.Fatalf("amount %v != %v*%d", o.PaymentAmount, o.ItemPrice, o.QuantityOfItem)
		}
		if (o.ItemName != "A" && o.ItemName != "B") || o.Category != "C" || len(o.OrderID) != 19 {
			t.Fatalf("bad order: %+v", o)
		}
		if !ordered(o) {
			t.Fatalf("time order wrong: %+v", o)
		}
	}
	// 联调模式: 点击即转化, 四个时间都等于点击时间
	if o := genOrder(cfg, "sp", pay, pay, r); o.ClickTime != o.OrderPayTime || o.OrderCreateTime != o.OrderPayTime || o.OrderPayClickTime != o.OrderPayTime || ts(o.OrderPayTime) != pay {
		t.Errorf("click==pay: %+v", o)
	}
	// 真实点击: 点击后 50ms 回传 / 正常模式 3 天后才回传
	for _, click := range []int64{pay - 50, pay - 3*86400*1000} {
		o := genOrder(cfg, "sp", click, pay, r)
		if ts(o.ClickTime) != click || !ordered(o) {
			t.Errorf("click=%d: %+v", click, o)
		}
	}
	// 点击时间晚于转化 (时钟异常) 退回随机
	if o := genOrder(cfg, "sp", pay+1, pay, r); !ordered(o) {
		t.Errorf("future click: %+v", o)
	}
	// 空配置走默认值
	o := genOrder(OrderConfig{}, "", 0, pay, r)
	if o.ItemName != "淘宝商品" || o.Category != "日用百货" || o.QuantityOfItem != 1 || o.ItemPrice != 10 {
		t.Errorf("defaults: %+v", o)
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
