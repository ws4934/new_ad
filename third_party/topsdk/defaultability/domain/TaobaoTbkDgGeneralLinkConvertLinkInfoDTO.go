package domain


type TaobaoTbkDgGeneralLinkConvertLinkInfoDTO struct {
    /*
        二合一长链接     */
    CouponLongUrl  *string `json:"coupon_long_url,omitempty" `

    /*
        1—单品 2—店铺 3—会场 4-承接开放 5-优惠券 6-直播间 7-淘积木页     */
    MaterialType  *int64 `json:"material_type,omitempty" `

    /*
        CPS链接长链     */
    CpsLongUrl  *string `json:"cps_long_url,omitempty" `

    /*
        CPS链接的短口令     */
    CpsShortTpwd  *string `json:"cps_short_tpwd,omitempty" `

    /*
        1-联盟口令，2-主站口令。入参物料不为淘口令时，此字段返回null     */
    TkBizType  *int64 `json:"tk_biz_type,omitempty" `

    /*
        二合一链接的短口令     */
    CouponShortTpwd  *string `json:"coupon_short_tpwd,omitempty" `

    /*
        CPS链接短链     */
    CpsShortUrl  *string `json:"cps_short_url,omitempty" `

    /*
        二合一链接长链     */
    CouponShortUrl  *string `json:"coupon_short_url,omitempty" `

    /*
        二合一链接的长口令     */
    CouponFullTpwd  *string `json:"coupon_full_tpwd,omitempty" `

    /*
        CPS链接的长口令     */
    CpsFullTpwd  *string `json:"cps_full_tpwd,omitempty" `

    /*
        种草页链接，定向开放     */
    ShareUnitUrl  *string `json:"share_unit_url,omitempty" `

    /*
        淘积木页面链接，定向开放     */
    TaoBrickUrl  *string `json:"tao_brick_url,omitempty" `

    /*
        频道页二合一链接，必须在required_link_type字段入参才可出参 目前支持淘金币     */
    CouponChannelUrl  *string `json:"coupon_channel_url,omitempty" `

    /*
        频道页二合一口令，必须在required_link_type字段入参才可出参 目前支持淘金币     */
    CouponChannelTpwd  *string `json:"coupon_channel_tpwd,omitempty" `

    /*
        频道页CPS链接，必须在required_link_type字段入参才可出参 目前支持淘金币     */
    CpsChannelUrl  *string `json:"cps_channel_url,omitempty" `

    /*
        频道页CPS口令，必须在required_link_type字段入参才可出参 目前支持淘金币     */
    CpsChannelTpwd  *string `json:"cps_channel_tpwd,omitempty" `

    /*
        超红二合一长链接，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CouponSuperedLongUrl  *string `json:"coupon_supered_long_url,omitempty" `

    /*
        超红二合一短链接，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CouponSuperedShortUrl  *string `json:"coupon_supered_short_url,omitempty" `

    /*
        超红二合一长口令，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CouponSuperedLongTpwd  *string `json:"coupon_supered_long_tpwd,omitempty" `

    /*
        超红二合一短口令，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CouponSuperedShortTpwd  *string `json:"coupon_supered_short_tpwd,omitempty" `

    /*
        超红cps长链接，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CpsSuperedLongUrl  *string `json:"cps_supered_long_url,omitempty" `

    /*
        超红cps短链接，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CpsSuperedShortUrl  *string `json:"cps_supered_short_url,omitempty" `

    /*
        超红cps长口令，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CpsSuperedLongTpwd  *string `json:"cps_supered_long_tpwd,omitempty" `

    /*
        超红cps短口令，必须在required_link_type字段入参才可出参（非超红活动期间，请求该字段返回普通推广链接）     */
    CpsSuperedShortTpwd  *string `json:"cps_supered_short_tpwd,omitempty" `

    /*
        卖家ID     */
    SellerId  *string `json:"seller_id,omitempty" `

    /*
        会场ID     */
    PageId  *string `json:"page_id,omitempty" `

}

func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetMaterialType(v int64) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.MaterialType = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetTkBizType(v int64) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.TkBizType = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponFullTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponFullTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsFullTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsFullTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetShareUnitUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.ShareUnitUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetTaoBrickUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.TaoBrickUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponChannelUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponChannelUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponChannelTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponChannelTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsChannelUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsChannelUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsChannelTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsChannelTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponSuperedLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponSuperedLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponSuperedShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponSuperedShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponSuperedLongTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponSuperedLongTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCouponSuperedShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CouponSuperedShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsSuperedLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsSuperedLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsSuperedShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsSuperedShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsSuperedLongTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsSuperedLongTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetCpsSuperedShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.CpsSuperedShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetSellerId(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.SellerId = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO) SetPageId(v string) *TaobaoTbkDgGeneralLinkConvertLinkInfoDTO {
    s.PageId = &v
    return s
}
