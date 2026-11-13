package domain


type TaobaoTbkDgMaterialOptionalUpgradeMapData struct {
    /*
        商品信息-淘宝客新商品id；使用说明参考《淘宝客新商品ID升级》白皮书：https://www.yuque.com/taobaolianmengguanfangxiaoer/zmig94/tfyt0pahmlpzu2ud     */
    ItemId  *string `json:"item_id,omitempty" `

    /*
        淘客推广信息     */
    PublishInfo  *TaobaoTbkDgMaterialOptionalUpgradePublishInfo `json:"publish_info,omitempty" `

    /*
        价格促销信息     */
    PricePromotionInfo  *TaobaoTbkDgMaterialOptionalUpgradePromotionInfoMapData `json:"price_promotion_info,omitempty" `

    /*
        商品基础信息     */
    ItemBasicInfo  *TaobaoTbkDgMaterialOptionalUpgradeBasicMapData `json:"item_basic_info,omitempty" `

    /*
        天猫榜单信息     */
    TmallRankInfo  *TaobaoTbkDgMaterialOptionalUpgradeTmallRankInfo `json:"tmall_rank_info,omitempty" `

    /*
        预售信息     */
    PresaleInfo  *TaobaoTbkDgMaterialOptionalUpgradePresaleInfo `json:"presale_info,omitempty" `

    /*
        商品库范围信息     */
    ScopeInfo  *TaobaoTbkDgMaterialOptionalUpgradeScopeInfo `json:"scope_info,omitempty" `

    /*
        线报内容     */
    MgcInfo  *TaobaoTbkDgMaterialOptionalUpgradeMgcInfo `json:"mgc_info,omitempty" `

    /*
        物料评估-匹配分     */
    MatchScore  *string `json:"match_score,omitempty" `

    /*
        物料评估-收益分     */
    CommiScore  *string `json:"commi_score,omitempty" `

    /*
        若为同款商品，枚举返回1 否则不返回该字段     */
    SimilarItem  *string `json:"similar_item,omitempty" `

    /*
        自购补贴标识     */
    ZgBtCode  *string `json:"zg_bt_code,omitempty" `

    /*
        链路关联标识     */
    TraceCode  *string `json:"trace_code,omitempty" `

    /*
        自购补贴专属优惠     */
    ZgBtPromotionInfo  *TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo `json:"zg_bt_promotion_info,omitempty" `

}

func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetItemId(v string) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.ItemId = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetPublishInfo(v TaobaoTbkDgMaterialOptionalUpgradePublishInfo) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.PublishInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetPricePromotionInfo(v TaobaoTbkDgMaterialOptionalUpgradePromotionInfoMapData) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.PricePromotionInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetItemBasicInfo(v TaobaoTbkDgMaterialOptionalUpgradeBasicMapData) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.ItemBasicInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetTmallRankInfo(v TaobaoTbkDgMaterialOptionalUpgradeTmallRankInfo) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.TmallRankInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetPresaleInfo(v TaobaoTbkDgMaterialOptionalUpgradePresaleInfo) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.PresaleInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetScopeInfo(v TaobaoTbkDgMaterialOptionalUpgradeScopeInfo) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.ScopeInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetMgcInfo(v TaobaoTbkDgMaterialOptionalUpgradeMgcInfo) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.MgcInfo = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetMatchScore(v string) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.MatchScore = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetCommiScore(v string) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.CommiScore = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetSimilarItem(v string) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.SimilarItem = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetZgBtCode(v string) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.ZgBtCode = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetTraceCode(v string) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.TraceCode = &v
    return s
}
func (s *TaobaoTbkDgMaterialOptionalUpgradeMapData) SetZgBtPromotionInfo(v TaobaoTbkDgMaterialOptionalUpgradeZgBtPromotionInfo) *TaobaoTbkDgMaterialOptionalUpgradeMapData {
    s.ZgBtPromotionInfo = &v
    return s
}
