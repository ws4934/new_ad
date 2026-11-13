package response

import (
    "topsdk/defaultability/domain"
)

type TaobaoTbkDgRtaSrIssueMatchResponse struct {

    /*
        System request id
    */
    RequestId string `json:"request_id,omitempty" `

    /*
        System body
    */
    Body string

    /*
        返回结果
    */
    Data  domain.TaobaoTbkDgRtaSrIssueMatchMediaIssueDTO `json:"data,omitempty" `
}
