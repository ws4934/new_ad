package response

import (
    "topsdk/ability1826/domain"
)

type TaobaoTbkDgVegasTljListReportResponse struct {

    /*
        System request id
    */
    RequestId string `json:"request_id,omitempty" `

    /*
        System body
    */
    Body string

    /*
        instancePageQueryResult
    */
    Model  domain.TaobaoTbkDgVegasTljListReportInstancePageQueryResult `json:"model,omitempty" `
    /*
        msg_info
    */
    MsgInfo  string `json:"msg_info,omitempty" `
    /*
        msgCode
    */
    MsgCode  string `json:"msg_code,omitempty" `
    /*
        result_success
    */
    ResultSuccess  bool `json:"result_success,omitempty" `
}
