package appinterface

import (
	"eva_services_go/frame/protocol"
)

// 获取调价方案列表请求参数
type Test struct {
	Id        string `json:"id" validate:"omitempty,numeric"`
	Status    string `json:"status" validate:"omitempty,oneof=0 1"`      // 方案状态 1有效；0无效
	Keyword   string `json:"keyword" validate:"omitempty,lte=64"`        // 搜索关键字 名称&ID
	PageIndex string `json:"pageIndex" validate:"required,numeric"`      // 分页：页码
	PageSize  string `json:"pageSize" validate:"required,numeric,lte=3"` // 分页：页数
}

// 获取调价方案列表响应参数
type RespTest struct {
	TestList []TestInfo           `json:"list"`
	Total    protocol.SubsysTotal `json:"pageInfo"`
}

type TestInfo struct {
	ClassId     string `json:"classId" validate:"required,numeric"`
	ClassName   string `json:"className"`
	ProductId   string `json:"productId" validate:"required,numeric"`
	ProductName string `json:"productName"  validate:"required,lte=64"` // 小于等于64个字符
	BrandId     string `json:"brandId" validate:"required,numeric"`
	BrandName   string `json:"brandName"`
	PicId       string `json:"picId" validate:"required"`
}

///////////////////////////
type TestSet struct {
	ItemId    string `json:"id" validate:"omitempty,numeric"`
	ItemName  string `json:"name" validate:"required,lte=100"`
	Pid       string `json:"pid" validate:"omitempty,numeric"`
	ClassId   string `json:"classId" validate:"required,numeric"`
	Attribute string `json:"attribute" validate:"required,oneof=1 2 3"`
	Type      string `json:"type" validate:"required,oneof=1 2 3"`
	Status    string `json:"status" validate:"required,oneof=0 1"`
	UserName  string `json:"userName" validate:"required,lte=100"`
}

/////
type TestEs struct {
	Id        string `json:"id" validate:"omitempty,numeric"`
	Status    string `json:"status" validate:"omitempty,oneof=0 1"`      // 方案状态 1有效；0无效
	Keyword   string `json:"keyword" validate:"omitempty,lte=64"`        // 搜索关键字 名称&ID
	PageIndex string `json:"pageIndex" validate:"required,numeric"`      // 分页：页码
	PageSize  string `json:"pageSize" validate:"required,numeric,lte=3"` // 分页：页数
}
