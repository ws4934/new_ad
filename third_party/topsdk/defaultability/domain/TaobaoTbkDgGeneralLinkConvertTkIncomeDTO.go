package domain


type TaobaoTbkDgGeneralLinkConvertTkIncomeDTO struct {
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
        补贴上限     */
    SubsidyUpperFee  *string `json:"subsidy_upper_fee,omitempty" `

    /*
        补贴上限；仅在单笔订单命中补贴上限时返回结果否则出参为空     */
    SubsidyType  *string `json:"subsidy_type,omitempty" `

}

func (s *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO) SetPureCommissionRate(v string) *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO {
    s.PureCommissionRate = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO) SetPureCommissionAmount(v int64) *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO {
    s.PureCommissionAmount = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO) SetSubsidyRate(v string) *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO {
    s.SubsidyRate = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO) SetSubsidyAmount(v int64) *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO {
    s.SubsidyAmount = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO) SetSubsidyUpperFee(v string) *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO {
    s.SubsidyUpperFee = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO) SetSubsidyType(v string) *TaobaoTbkDgGeneralLinkConvertTkIncomeDTO {
    s.SubsidyType = &v
    return s
}
