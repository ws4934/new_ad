package domain


type TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo struct {
    /*
        优惠名称，如“官方淘礼金"     */
    PromotionTitle  *string `json:"promotion_title,omitempty" `

    /*
        优惠金额（元）     */
    PromotionFee  *string `json:"promotion_fee,omitempty" `

}

func (s *TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo) SetPromotionTitle(v string) *TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo {
    s.PromotionTitle = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo) SetPromotionFee(v string) *TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo {
    s.PromotionFee = &v
    return s
}
