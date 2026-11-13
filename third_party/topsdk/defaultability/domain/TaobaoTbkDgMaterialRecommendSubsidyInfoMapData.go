package domain


type TaobaoTbkDgMaterialRecommendSubsidyInfoMapData struct {
    /*
        补贴类型明细     */
    SubsidyTypeDetail  *string `json:"subsidy_type_detail,omitempty" `

    /*
        单笔补贴上限明细     */
    SubsidyUpperFeeDetail  *string `json:"subsidy_upper_fee_detail,omitempty" `

    /*
        补贴比率明细     */
    SubsidyRateDetail  *string `json:"subsidy_rate_detail,omitempty" `

    /*
        补贴金额明细     */
    SubsidyAmountDetail  *string `json:"subsidy_amount_detail,omitempty" `

}

func (s *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData) SetSubsidyTypeDetail(v string) *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData {
    s.SubsidyTypeDetail = &v
    return s
}
func (s *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData) SetSubsidyUpperFeeDetail(v string) *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData {
    s.SubsidyUpperFeeDetail = &v
    return s
}
func (s *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData) SetSubsidyRateDetail(v string) *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData {
    s.SubsidyRateDetail = &v
    return s
}
func (s *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData) SetSubsidyAmountDetail(v string) *TaobaoTbkDgMaterialRecommendSubsidyInfoMapData {
    s.SubsidyAmountDetail = &v
    return s
}
