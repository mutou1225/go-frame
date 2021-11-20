package logicmodel

import (
	"go-frame/frame/protocol"
	"strconv"
)

type ProductAttributeReq struct {
	ProductId string `json:"productId" validate:"required,numeric"`		// 产品ID
}

type ProductAttribute struct {
	ProductId   string	`json:"productId"`
	ProductName string	`json:"productName"`
	ProductLogo string	`json:"productLogo"`
	BrandId     string	`json:"brandId"`
	BrandName   string	`json:"brandName"`
	ClassId     string	`json:"classId"`
	ClassName   string	`json:"className"`
	MarketTime  string	`json:"marketTime"`
	OsId        string	`json:"osId"`
	OsName      string	`json:"osName"`
	RecycleType string	`json:"recycleType"`
	RecycleName string	`json:"recycleName"`
	AttrList    []struct {
		AttrItemId   string	`json:"attrItemId"`
		AttrItemName string	`json:"attrItemName"`
		AttrTypeId   string	`json:"attrTypeId"`
		AttrTypeName string	`json:"attrTypeName"`
		Mandatory    string	`json:"mandatory"`
	}						`json:"attrList"`
	SkuList []struct {
		AnswerId     string	`json:"answerId"`
		AnswerName   string	`json:"answerName"`
		QuestionId   string	`json:"questionId"`
		QuestionName string	`json:"questionName"`
	}						`json:"skuList"`
}

// http://wiki.huishoubao.com/web/#/105?page_id=3671
// 获取产品属性：sku、属性等
func GetProductAttribute(productId int, callerServiceId string, signKey string) (*ProductAttribute, error) {
	params := ProductAttributeReq{strconv.Itoa(productId)}

	strUrl := "http://prdserver.huishoubao.com/rpc/new_product_lib"
	strInterface := "getProductAttribute"

	var response ProductAttribute
	if err := protocol.RequestCgiModel(protocol.ProtocolV1, strUrl, strInterface, callerServiceId, signKey, params, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
