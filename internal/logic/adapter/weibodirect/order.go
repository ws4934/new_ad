package weibodirect

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"time"
)

// OrderConfig 账户凭证里的随机订单配置 (credentials.order)
// 淘联回调不带订单信息, behavior=2001 必填的 item_order_pay 按此随机生成
type OrderConfig struct {
	ItemNames  []string `json:"item_names"`
	Categories []string `json:"categories"`
	PriceMin   float64  `json:"price_min"`
	PriceMax   float64  `json:"price_max"`
	QtyMin     int      `json:"qty_min"`
	QtyMax     int      `json:"qty_max"`
}

// itemOrderPay item_order_pay 字段 (文档: time 相关传 string, price 相关传 float)
type itemOrderPay struct {
	ClickTime         string  `json:"click_time"`
	OrderCreateTime   string  `json:"order_create_time"`
	OrderPayTime      string  `json:"order_pay_time"`
	OrderPayClickTime string  `json:"order_pay_click_time"`
	ItemName          string  `json:"item_name"`
	ItemPrice         float64 `json:"item_price"`
	PaymentAmount     float64 `json:"payment_amount"`
	QuantityOfItem    int     `json:"quantity_of_item"`
	Category          string  `json:"category"`
	OrderStatus       string  `json:"order_status"`
	OrderID           string  `json:"order_id"`
	ServiceProvider   string  `json:"service_provider"`
}

// genOrder 按配置随机生成一笔已付款订单; 未配置项用默认值
// 时间线: 点击 → (1~30 分钟) 下单 → (10~120 秒) 付款点击 → (1~10 秒) 付款 = 转化时间 payMs。
// 有真实点击时间 clickMs 时点击用真实值, 下单/付款点击压缩在点击与付款之间;
// 联调模式点击即转化 (clickMs == payMs) 时四个时间相同, 与文档示例一致
func genOrder(cfg OrderConfig, serviceProvider string, clickMs, payMs int64, r *rand.Rand) itemOrderPay {
	pick := func(list []string, def string) string {
		if len(list) == 0 {
			return def
		}
		return list[r.IntN(len(list))]
	}
	priceMin, priceMax := cfg.PriceMin, cfg.PriceMax
	if priceMin <= 0 {
		priceMin = 10
	}
	if priceMax < priceMin {
		priceMax = priceMin
	}
	qtyMin, qtyMax := cfg.QtyMin, cfg.QtyMax
	if qtyMin <= 0 {
		qtyMin = 1
	}
	if qtyMax < qtyMin {
		qtyMax = qtyMin
	}
	round2 := func(f float64) float64 { return math.Round(f*100) / 100 }

	price := round2(priceMin + r.Float64()*(priceMax-priceMin))
	qty := qtyMin + r.IntN(qtyMax-qtyMin+1)
	ms := func(t time.Time) string { return strconv.FormatInt(t.UnixMilli(), 10) }

	payGap := time.Duration(1+r.IntN(10)) * time.Second
	createGap := time.Duration(10+r.IntN(111)) * time.Second
	clickGap := time.Duration(1+r.IntN(30)) * time.Minute
	pay := time.UnixMilli(payMs)
	realClick := clickMs > 0 && clickMs <= payMs
	if realClick {
		span := pay.Sub(time.UnixMilli(clickMs)) / 3
		payGap, createGap = min(payGap, span), min(createGap, span)
	}
	payClick := pay.Add(-payGap)
	create := payClick.Add(-createGap)
	click := create.Add(-clickGap)
	if realClick {
		click = time.UnixMilli(clickMs)
	}

	return itemOrderPay{
		ClickTime:         ms(click),
		OrderCreateTime:   ms(create),
		OrderPayTime:      ms(pay),
		OrderPayClickTime: ms(payClick),
		ItemName:          pick(cfg.ItemNames, "淘宝商品"),
		ItemPrice:         price,
		PaymentAmount:     round2(price * float64(qty)),
		QuantityOfItem:    qty,
		Category:          pick(cfg.Categories, "日用百货"),
		OrderStatus:       "已付款",
		// 淘宝订单号形态: 19 位数字
		OrderID:         fmt.Sprintf("%d%018d", 1+r.IntN(9), r.Int64N(1e18)),
		ServiceProvider: serviceProvider,
	}
}
