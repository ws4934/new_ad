package domain


type TaobaoTbkDgVegasTljListReportPageResult struct {
    /*
        result     */
    Result  *[]TaobaoTbkDgVegasTljListReportInstanceDTO `json:"result,omitempty" `

    /*
        当前页码     */
    PageNo  *string `json:"page_no,omitempty" `

    /*
        每页分片     */
    PageSize  *string `json:"page_size,omitempty" `

    /*
        总共页数     */
    TotalPages  *string `json:"total_pages,omitempty" `

    /*
        总共数量     */
    TotalCount  *string `json:"total_count,omitempty" `

}

func (s *TaobaoTbkDgVegasTljListReportPageResult) SetResult(v []TaobaoTbkDgVegasTljListReportInstanceDTO) *TaobaoTbkDgVegasTljListReportPageResult {
    s.Result = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportPageResult) SetPageNo(v string) *TaobaoTbkDgVegasTljListReportPageResult {
    s.PageNo = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportPageResult) SetPageSize(v string) *TaobaoTbkDgVegasTljListReportPageResult {
    s.PageSize = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportPageResult) SetTotalPages(v string) *TaobaoTbkDgVegasTljListReportPageResult {
    s.TotalPages = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportPageResult) SetTotalCount(v string) *TaobaoTbkDgVegasTljListReportPageResult {
    s.TotalCount = &v
    return s
}
