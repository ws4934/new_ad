package domain


type TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO struct {
    /*
        优惠名称，如“商品券”、“跨店满减”、“单品直降”等     */
    PromotionTitle  *string `json:"promotion_title,omitempty" `

    /*
        预热优惠利益点文案，如“1件7.92折”、“每200减20”等     */
    PromotionDesc  *string `json:"promotion_desc,omitempty" `

    /*
        预热实际优惠金额（元）     */
    PromotionFee  *string `json:"promotion_fee,omitempty" `

    /*
        优惠开始时间     */
    PromotionStartTime  *string `json:"promotion_start_time,omitempty" `

    /*
        优惠结束时间     */
    PromotionEndTime  *string `json:"promotion_end_time,omitempty" `

}

func (s *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO) SetPromotionTitle(v string) *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO {
    s.PromotionTitle = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO) SetPromotionDesc(v string) *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO {
    s.PromotionDesc = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO) SetPromotionFee(v string) *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO {
    s.PromotionFee = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO) SetPromotionStartTime(v string) *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO {
    s.PromotionStartTime = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO) SetPromotionEndTime(v string) *TaobaoTbkItemInfoUpgradeGetFutureActivityPromotionDTO {
    s.PromotionEndTime = &v
    return s
}
