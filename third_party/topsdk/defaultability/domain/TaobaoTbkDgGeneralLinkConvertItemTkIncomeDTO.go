package domain


type TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO struct {
    /*
        商品佣金比例     */
    PureCommissionRate  *string `json:"pure_commission_rate,omitempty" `

    /*
        商品佣金金额     */
    PureCommissionAmount  *int64 `json:"pure_commission_amount,omitempty" `

    /*
        补贴比例     */
    SubsidyRate  *string `json:"subsidy_rate,omitempty" `

    /*
        补贴金额     */
    SubsidyAmount  *int64 `json:"subsidy_amount,omitempty" `

    /*
        补贴上限；仅在单笔订单命中补贴上限时返回结果否则出参为空     */
    SubsidyUpperFee  *string `json:"subsidy_upper_fee,omitempty" `

    /*
        补贴类型     */
    SubsidyType  *string `json:"subsidy_type,omitempty" `

}

func (s *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO) SetPureCommissionRate(v string) *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO {
    s.PureCommissionRate = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO) SetPureCommissionAmount(v int64) *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO {
    s.PureCommissionAmount = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO) SetSubsidyRate(v string) *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO {
    s.SubsidyRate = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO) SetSubsidyAmount(v int64) *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO {
    s.SubsidyAmount = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO) SetSubsidyUpperFee(v string) *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO {
    s.SubsidyUpperFee = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO) SetSubsidyType(v string) *TaobaoTbkDgGeneralLinkConvertItemTkIncomeDTO {
    s.SubsidyType = &v
    return s
}
