package domain


type TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData struct {
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

func (s *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData) SetSubsidyTypeDetail(v string) *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData {
    s.SubsidyTypeDetail = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData) SetSubsidyUpperFeeDetail(v string) *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData {
    s.SubsidyUpperFeeDetail = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData) SetSubsidyRateDetail(v string) *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData {
    s.SubsidyRateDetail = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData) SetSubsidyAmountDetail(v string) *TaobaoTbkDgMaterialOptionalUpgradeSubsidyInfoMapData {
    s.SubsidyAmountDetail = &v
    return s
}
