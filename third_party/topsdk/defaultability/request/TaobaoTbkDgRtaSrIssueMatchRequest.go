package request


type TaobaoTbkDgRtaSrIssueMatchRequest struct {
    /*
        活动序号：1（超红）、2（福利购）、4（福利直降版本A+版本B）、不传时默认查询超级红包数据 defalutValue��1    */
    ActivityCategory  *int64 `json:"activity_category,omitempty" required:"false" `
}

func (s *TaobaoTbkDgRtaSrIssueMatchRequest) SetActivityCategory(v int64) *TaobaoTbkDgRtaSrIssueMatchRequest {
    s.ActivityCategory = &v
    return s
}

func (req *TaobaoTbkDgRtaSrIssueMatchRequest) ToMap() map[string]interface{} {
    paramMap := make(map[string]interface{})
    if(req.ActivityCategory != nil) {
        paramMap["activity_category"] = *req.ActivityCategory
    }
    return paramMap
}

func (req *TaobaoTbkDgRtaSrIssueMatchRequest) ToFileMap() map[string]interface{} {
    fileMap := make(map[string]interface{})
    return fileMap
}