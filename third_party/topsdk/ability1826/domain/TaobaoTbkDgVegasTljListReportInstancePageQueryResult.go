package domain


type TaobaoTbkDgVegasTljListReportInstancePageQueryResult struct {
    /*
        PageResult     */
    Result  *TaobaoTbkDgVegasTljListReportPageResult `json:"result,omitempty" `

    /*
        position_index     */
    PositionIndex  *string `json:"position_index,omitempty" `

}

func (s *TaobaoTbkDgVegasTljListReportInstancePageQueryResult) SetResult(v TaobaoTbkDgVegasTljListReportPageResult) *TaobaoTbkDgVegasTljListReportInstancePageQueryResult {
    s.Result = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportInstancePageQueryResult) SetPositionIndex(v string) *TaobaoTbkDgVegasTljListReportInstancePageQueryResult {
    s.PositionIndex = &v
    return s
}
