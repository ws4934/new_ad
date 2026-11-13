package domain


type TaobaoTbkItemInfoUpgradeGetZgBtPromotionInfo struct {
    /*
        优惠名称，如“官方淘礼金"     */
    PromotionTitle  *string `json:"promotion_title,omitempty" `

    /*
        优惠金额（元）     */
    PromotionFee  *string `json:"promotion_fee,omitempty" `

}

func (s *TaobaoTbkItemInfoUpgradeGetZgBtPromotionInfo) SetPromotionTitle(v string) *TaobaoTbkItemInfoUpgradeGetZgBtPromotionInfo {
    s.PromotionTitle = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetZgBtPromotionInfo) SetPromotionFee(v string) *TaobaoTbkItemInfoUpgradeGetZgBtPromotionInfo {
    s.PromotionFee = &v
    return s
}
