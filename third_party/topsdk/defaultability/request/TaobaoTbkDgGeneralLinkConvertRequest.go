package request

import (
        "topsdk/defaultability/domain"
        "topsdk/util"
    )

type TaobaoTbkDgGeneralLinkConvertRequest struct {
    /*
        1-动态ID转链场景，2-消费者比价场景，4-     */
    BizSceneId  *string `json:"biz_scene_id,omitempty" required:"false" `
    /*
        1-自购省，2-推广赚（代理模式专属ID，代理模式必填，其它模式不用填写本字段）     */
    PromotionType  *string `json:"promotion_type,omitempty" required:"false" `
    /*
        物料列表，可以为url或淘口令,多个时使用英文逗号拼接传入（邀约制权限有申请门槛，权限分为①联盟推广链接转链；②天猫/淘宝复制链接转链；如有相关需求请在权限申请时同步说明）     */
    MaterialList  *[]string `json:"material_list,omitempty" required:"false" `
    /*
        渠道管理ID（如是主站选品推广场景，必须入参该字段，且bizSceneId字段需入参2-消费者比价场景，否则二次转链失败）     */
    RelationId  *string `json:"relation_id,omitempty" required:"false" `
    /*
        推广位id，mm_xx_xx_xx pid三段式中的第三段     */
    AdzoneId  *int64 `json:"adzone_id" required:"true" `
    /*
        卖家ID列表,多个时使用英文逗号拼接传入     */
    SellerIdList  *[]string `json:"seller_id_list,omitempty" required:"false" `
    /*
        商品ID列表,多个时使用英文逗号拼接传入     */
    ItemIdList  *[]string `json:"item_id_list,omitempty" required:"false" `
    /*
        会场ID列表,多个时使用英文逗号拼接传入     */
    PageIdList  *[]string `json:"page_id_list,omitempty" required:"false" `
    /*
        会场页面内定坑商品     */
    TargetItem  *domain.TaobaoTbkDgGeneralLinkConvertTargetItemDTO `json:"target_item,omitempty" required:"false" `
    /*
        商品转链     */
    ItemDto  *[]domain.TaobaoTbkDgGeneralLinkConvertLkItemDTO `json:"item_dto,omitempty" required:"false" `
    /*
        会场页面转链     */
    PageDto  *[]domain.TaobaoTbkDgGeneralLinkConvertLkPageDTO `json:"page_dto,omitempty" required:"false" `
    /*
        店铺转链     */
    ShopDto  *[]domain.TaobaoTbkDgGeneralLinkConvertLkShopDTO `json:"shop_dto,omitempty" required:"false" `
    /*
        链接/口令转链     */
    MaterialDto  *[]domain.TaobaoTbkDgGeneralLinkConvertLkMaterialDTO `json:"material_dto,omitempty" required:"false" `
    /*
        会员运营id     */
    SpecialId  *string `json:"special_id,omitempty" required:"false" `
    /*
        加密用户标识     */
    Uvid  *string `json:"uvid,omitempty" required:"false" `
    /*
        启明系统任务ID     */
    Qmtid  *string `json:"qmtid,omitempty" required:"false" `
    /*
        返利类UVID是否展示授权弹窗 0-弱提醒弹窗 1-不展示弹窗 2强提醒弹窗 3-强制授权引导(邀约制权限)  默认为0（引导授权成功后在订单API可透出UVID）     */
    UvidNotAuthorized  *int64 `json:"uvid_not_authorized,omitempty" required:"false" `
    /*
        用于指定link_info_dto字段出参。coupon_long_url, cps_long_url, cps_short_tpwd, coupon_short_tpwd, cps_short_url, coupon_short_url, coupon_full_tpwd, cps_full_tpwd, share_unit_url, tao_brick_url     */
    RequiredLinkType  *string `json:"required_link_type,omitempty" required:"false" `
    /*
        设备号加密类型：MD5；使用请先签署协议https://pub.alimama.com/fourth/protocol/common.htm?key=hangye_laxin     */
    DeviceEncrypt  *string `json:"device_encrypt,omitempty" required:"false" `
    /*
        设备号加密后的值（MD5加密需32位小写）；使用请先签署协议https://pub.alimama.com/fourth/protocol/common.htm?key=hangye_laxin     */
    DeviceValue  *string `json:"device_value,omitempty" required:"false" `
    /*
        设备号类型：OAID/IDFA；使用请先签署协议https://pub.alimama.com/fourth/protocol/common.htm?key=hangye_laxin     */
    DeviceType  *string `json:"device_type,omitempty" required:"false" `
}

func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetBizSceneId(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.BizSceneId = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetPromotionType(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.PromotionType = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetMaterialList(v []string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.MaterialList = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetRelationId(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.RelationId = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetAdzoneId(v int64) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.AdzoneId = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetSellerIdList(v []string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.SellerIdList = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetItemIdList(v []string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.ItemIdList = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetPageIdList(v []string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.PageIdList = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetTargetItem(v domain.TaobaoTbkDgGeneralLinkConvertTargetItemDTO) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.TargetItem = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetItemDto(v []domain.TaobaoTbkDgGeneralLinkConvertLkItemDTO) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.ItemDto = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetPageDto(v []domain.TaobaoTbkDgGeneralLinkConvertLkPageDTO) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.PageDto = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetShopDto(v []domain.TaobaoTbkDgGeneralLinkConvertLkShopDTO) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.ShopDto = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetMaterialDto(v []domain.TaobaoTbkDgGeneralLinkConvertLkMaterialDTO) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.MaterialDto = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetSpecialId(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.SpecialId = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetUvid(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.Uvid = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetQmtid(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.Qmtid = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetUvidNotAuthorized(v int64) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.UvidNotAuthorized = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetRequiredLinkType(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.RequiredLinkType = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetDeviceEncrypt(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.DeviceEncrypt = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetDeviceValue(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.DeviceValue = &v
    return s
}
func (s *TaobaoTbkDgGeneralLinkConvertRequest) SetDeviceType(v string) *TaobaoTbkDgGeneralLinkConvertRequest {
    s.DeviceType = &v
    return s
}

func (req *TaobaoTbkDgGeneralLinkConvertRequest) ToMap() map[string]interface{} {
    paramMap := make(map[string]interface{})
    if(req.BizSceneId != nil) {
        paramMap["biz_scene_id"] = *req.BizSceneId
    }
    if(req.PromotionType != nil) {
        paramMap["promotion_type"] = *req.PromotionType
    }
    if(req.MaterialList != nil) {
        paramMap["material_list"] = util.ConvertBasicList(*req.MaterialList)
    }
    if(req.RelationId != nil) {
        paramMap["relation_id"] = *req.RelationId
    }
    if(req.AdzoneId != nil) {
        paramMap["adzone_id"] = *req.AdzoneId
    }
    if(req.SellerIdList != nil) {
        paramMap["seller_id_list"] = util.ConvertBasicList(*req.SellerIdList)
    }
    if(req.ItemIdList != nil) {
        paramMap["item_id_list"] = util.ConvertBasicList(*req.ItemIdList)
    }
    if(req.PageIdList != nil) {
        paramMap["page_id_list"] = util.ConvertBasicList(*req.PageIdList)
    }
    if(req.TargetItem != nil) {
        paramMap["target_item"] = util.ConvertStruct(*req.TargetItem)
    }
    if(req.ItemDto != nil) {
        paramMap["item_dto"] = util.ConvertStructList(*req.ItemDto)
    }
    if(req.PageDto != nil) {
        paramMap["page_dto"] = util.ConvertStructList(*req.PageDto)
    }
    if(req.ShopDto != nil) {
        paramMap["shop_dto"] = util.ConvertStructList(*req.ShopDto)
    }
    if(req.MaterialDto != nil) {
        paramMap["material_dto"] = util.ConvertStructList(*req.MaterialDto)
    }
    if(req.SpecialId != nil) {
        paramMap["special_id"] = *req.SpecialId
    }
    if(req.Uvid != nil) {
        paramMap["uvid"] = *req.Uvid
    }
    if(req.Qmtid != nil) {
        paramMap["qmtid"] = *req.Qmtid
    }
    if(req.UvidNotAuthorized != nil) {
        paramMap["uvid_not_authorized"] = *req.UvidNotAuthorized
    }
    if(req.RequiredLinkType != nil) {
        paramMap["required_link_type"] = *req.RequiredLinkType
    }
    if(req.DeviceEncrypt != nil) {
        paramMap["device_encrypt"] = *req.DeviceEncrypt
    }
    if(req.DeviceValue != nil) {
        paramMap["device_value"] = *req.DeviceValue
    }
    if(req.DeviceType != nil) {
        paramMap["device_type"] = *req.DeviceType
    }
    return paramMap
}

func (req *TaobaoTbkDgGeneralLinkConvertRequest) ToFileMap() map[string]interface{} {
    fileMap := make(map[string]interface{})
    return fileMap
}