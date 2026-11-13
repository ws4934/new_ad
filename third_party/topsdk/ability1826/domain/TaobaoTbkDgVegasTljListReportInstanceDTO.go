package domain


type TaobaoTbkDgVegasTljListReportInstanceDTO struct {
    /*
        礼金ID     */
    RightsId  *string `json:"rights_id,omitempty" `

    /*
        Extra     */
    Extra  *TaobaoTbkDgVegasTljListReportExtra `json:"extra,omitempty" `

}

func (s *TaobaoTbkDgVegasTljListReportInstanceDTO) SetRightsId(v string) *TaobaoTbkDgVegasTljListReportInstanceDTO {
    s.RightsId = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportInstanceDTO) SetExtra(v TaobaoTbkDgVegasTljListReportExtra) *TaobaoTbkDgVegasTljListReportInstanceDTO {
    s.Extra = &v
    return s
}
