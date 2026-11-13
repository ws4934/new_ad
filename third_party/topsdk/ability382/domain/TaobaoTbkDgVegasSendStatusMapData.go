package domain


type TaobaoTbkDgVegasSendStatusMapData struct {
    /*
        若该用户当前无法领取该红包，则返回0，若当前可领取该红包，则返回1。（已核销红包不返回）。     */
    IsNewUser  *string `json:"is_new_user,omitempty" `

}

func (s *TaobaoTbkDgVegasSendStatusMapData) SetIsNewUser(v string) *TaobaoTbkDgVegasSendStatusMapData {
    s.IsNewUser = &v
    return s
}
