package request


type TaobaoTbkDgVegasTljListReportRequest struct {
    /*
        推广位ID     */
    AdzoneId  *int64 `json:"adzone_id" required:"true" `
    /*
        礼金创建的起始时刻     */
    BeginDate  *int64 `json:"begin_date" required:"true" `
    /*
        礼金创建的最后时刻     */
    EndDate  *int64 `json:"end_date" required:"true" `
    /*
        页码     */
    PageNo  *int64 `json:"page_no" required:"true" `
    /*
        翻页不能超过10一页     */
    PageSize  *int64 `json:"page_size" required:"true" `
}

func (s *TaobaoTbkDgVegasTljListReportRequest) SetAdzoneId(v int64) *TaobaoTbkDgVegasTljListReportRequest {
    s.AdzoneId = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportRequest) SetBeginDate(v int64) *TaobaoTbkDgVegasTljListReportRequest {
    s.BeginDate = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportRequest) SetEndDate(v int64) *TaobaoTbkDgVegasTljListReportRequest {
    s.EndDate = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportRequest) SetPageNo(v int64) *TaobaoTbkDgVegasTljListReportRequest {
    s.PageNo = &v
    return s
}
func (s *TaobaoTbkDgVegasTljListReportRequest) SetPageSize(v int64) *TaobaoTbkDgVegasTljListReportRequest {
    s.PageSize = &v
    return s
}

func (req *TaobaoTbkDgVegasTljListReportRequest) ToMap() map[string]interface{} {
    paramMap := make(map[string]interface{})
    if(req.AdzoneId != nil) {
        paramMap["adzone_id"] = *req.AdzoneId
    }
    if(req.BeginDate != nil) {
        paramMap["begin_date"] = *req.BeginDate
    }
    if(req.EndDate != nil) {
        paramMap["end_date"] = *req.EndDate
    }
    if(req.PageNo != nil) {
        paramMap["page_no"] = *req.PageNo
    }
    if(req.PageSize != nil) {
        paramMap["page_size"] = *req.PageSize
    }
    return paramMap
}

func (req *TaobaoTbkDgVegasTljListReportRequest) ToFileMap() map[string]interface{} {
    fileMap := make(map[string]interface{})
    return fileMap
}