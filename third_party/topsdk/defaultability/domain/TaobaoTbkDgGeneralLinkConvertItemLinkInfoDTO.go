package domain


type TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO struct {
    /*
        二合一长链接     */
    CouponLongUrl  *string `json:"coupon_long_url,omitempty" `

    /*
        1—单品 2—店铺 3—会场     */
    MaterialType  *int64 `json:"material_type,omitempty" `

    /*
        商品ID     */
    ItemId  *string `json:"item_id,omitempty" `

    /*
        CPS链接长链     */
    CpsLongUrl  *string `json:"cps_long_url,omitempty" `

    /*
        CPS链接的短口令     */
    CpsShortTpwd  *string `json:"cps_short_tpwd,omitempty" `

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

}

func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetMaterialType(v int64) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.MaterialType = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetItemId(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.ItemId = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponFullTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponFullTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsFullTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsFullTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetShareUnitUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.ShareUnitUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponChannelUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponChannelUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponChannelTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponChannelTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsChannelUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsChannelUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsChannelTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsChannelTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponSuperedLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponSuperedLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponSuperedShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponSuperedShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponSuperedLongTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponSuperedLongTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCouponSuperedShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CouponSuperedShortTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsSuperedLongUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsSuperedLongUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsSuperedShortUrl(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsSuperedShortUrl = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsSuperedLongTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsSuperedLongTpwd = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO) SetCpsSuperedShortTpwd(v string) *TaobaoTbkDgGeneralLinkConvertItemLinkInfoDTO {
    s.CpsSuperedShortTpwd = &v
    return s
}
