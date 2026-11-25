package weibodirect

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestBuildQuery(t *testing.T) {
	// 宏替换后经 r.Get 解码得到的原始 markid
	mark := "4_reallog_mark_ad:8|4_11111111111111111111"
	now := time.UnixMilli(1581645600000)
	c := credentials{ServiceProvider: "某服务商", TechnicalProvider: "某技术商"}
	q := buildQuery(mark, "2001", "cb.example.com", c, now, rand.New(rand.NewSource(1)))

	// 服务端 decode 一次应还原为原值 (只 encode 一次)
	v, err := url.ParseQuery(q)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"time": "1581645600000", "behavior": "2001", "mark_id": mark,
		"host": "cb.example.com", "technical_provider": "某技术商",
	}
	for k, w := range want {
		if v.Get(k) != w {
			t.Errorf("%s = %q, want %q", k, v.Get(k), w)
		}
	}
	var o itemOrderPay
	if err := json.Unmarshal([]byte(v.Get("item_order_pay")), &o); err != nil {
		t.Fatalf("item_order_pay: %v", err)
	}
	if o.ServiceProvider != "某服务商" || o.OrderStatus != "已付款" || o.OrderPayTime != "1581645600000" {
		t.Errorf("order = %+v", o)
	}

	// 非 2001 不带订单, 空 host/technical_provider 不传
	v, _ = url.ParseQuery(buildQuery(mark, "1", "", credentials{}, now, rand.New(rand.NewSource(1))))
	for _, k := range []string{"item_order_pay", "host", "technical_provider"} {
		if v.Has(k) {
			t.Errorf("unexpected %s", k)
		}
	}
}

func TestGenOrder(t *testing.T) {
	cfg := OrderConfig{ItemNames: []string{"A", "B"}, Categories: []string{"C"}, PriceMin: 10, PriceMax: 20, QtyMin: 1, QtyMax: 3}
	now := time.Now()
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		o := genOrder(cfg, "sp", now, r)
		if o.ItemPrice < 10 || o.ItemPrice > 20 || o.QuantityOfItem < 1 || o.QuantityOfItem > 3 {
			t.Fatalf("out of range: %+v", o)
		}
		if o.PaymentAmount != float64(int64(o.ItemPrice*float64(o.QuantityOfItem)*100+0.5))/100 {
			t.Fatalf("amount %v != %v*%d", o.PaymentAmount, o.ItemPrice, o.QuantityOfItem)
		}
		if (o.ItemName != "A" && o.ItemName != "B") || o.Category != "C" || len(o.OrderID) != 19 {
			t.Fatalf("bad order: %+v", o)
		}
		ts := func(s string) int64 { n, _ := strconv.ParseInt(s, 10, 64); return n }
		if !(ts(o.ClickTime) < ts(o.OrderCreateTime) && ts(o.OrderCreateTime) < ts(o.OrderPayClickTime) && ts(o.OrderPayClickTime) < ts(o.OrderPayTime)) {
			t.Fatalf("time order wrong: %+v", o)
		}
	}
	// 空配置走默认值
	o := genOrder(OrderConfig{}, "", now, r)
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
