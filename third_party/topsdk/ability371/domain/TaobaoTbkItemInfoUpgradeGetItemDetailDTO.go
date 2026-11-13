package domain


type TaobaoTbkItemInfoUpgradeGetItemDetailDTO struct {
    /*
        关联标识     */
    TraceCode  *string `json:"trace_code,omitempty" `

    /*
        自购补贴标识     */
    ZgBtCode  *string `json:"zg_bt_code,omitempty" `

    /*
        商品ID     */
    ItemId  *string `json:"item_id,omitempty" `

}

func (s *TaobaoTbkItemInfoUpgradeGetItemDetailDTO) SetTraceCode(v string) *TaobaoTbkItemInfoUpgradeGetItemDetailDTO {
    s.TraceCode = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetItemDetailDTO) SetZgBtCode(v string) *TaobaoTbkItemInfoUpgradeGetItemDetailDTO {
    s.ZgBtCode = &v
    return s
}
func (s *TaobaoTbkItemInfoUpgradeGetItemDetailDTO) SetItemId(v string) *TaobaoTbkItemInfoUpgradeGetItemDetailDTO {
    s.ItemId = &v
    return s
}
