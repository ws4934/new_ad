package domain


type TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO struct {
    /*
        当前媒体(不在任何媒体组时)或所在的媒体组是否可直塞     */
    CanIssue  *bool `json:"can_issue,omitempty" `

    /*
        直塞红包数量较低时的提示信息     */
    ExtraMsg  *string `json:"extra_msg,omitempty" `

    /*
        直塞红包数量较低时返回剩余直塞红包数量     */
    AutoPayoutCnt  *int64 `json:"auto_payout_cnt,omitempty" `

    /*
        手动领取红包数量较低时返回剩余剩余手抽红包数量     */
    ManualClaim2Go  *int64 `json:"manual_claim_2_go,omitempty" `

}

func (s *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO) SetCanIssue(v bool) *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO {
    s.CanIssue = &v
    return s
}
func (s *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO) SetExtraMsg(v string) *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO {
    s.ExtraMsg = &v
    return s
}
func (s *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO) SetAutoPayoutCnt(v int64) *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO {
    s.AutoPayoutCnt = &v
    return s
}
func (s *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO) SetManualClaim2Go(v int64) *TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO {
    s.ManualClaim2Go = &v
    return s
}
