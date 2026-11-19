package model

// 门店绑定码与租户层级接口请求/响应结构体

// DeviceBindCodeCreateReq 门店生成绑定码
type DeviceBindCodeCreateReq struct {
	DeviceID         string `json:"device_id" validate:"required,max=36"`
	MaxUses          int32  `json:"max_uses" validate:"omitempty,min=1,max=1000"`
	ExpiresInMinutes int32  `json:"expires_in_minutes" validate:"omitempty,min=0,max=525600"` // 0 或空表示不过期
}

// DeviceBindCodeResp 绑定码信息
type DeviceBindCodeResp struct {
	ID         string `json:"id"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Code       string `json:"code"`
	ExpiresAt  string `json:"expires_at"`
	MaxUses    int32  `json:"max_uses"`
	UsedCount  int32  `json:"used_count"`
	Status     int16  `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// DeviceBindCodeQueryReq 绑定码列表查询（device_id 为空时返回本租户全部）
type DeviceBindCodeQueryReq struct {
	DeviceID string `json:"device_id" form:"device_id" validate:"omitempty,max=36"`
}

// DeviceBindCodeStatusReq 启用/失效绑定码
type DeviceBindCodeStatusReq struct {
	ID     string `json:"id" validate:"required,max=36"`
	Status int16  `json:"status" validate:"oneof=0 1"`
}

// TenantSaveReq 新增/修改租户层级节点
type TenantSaveReq struct {
	ID       string `json:"id" validate:"omitempty,max=36"`
	ParentID string `json:"parent_id" validate:"omitempty,max=36"`
	Name     string `json:"name" validate:"required,max=255"`
	Role     string `json:"role" validate:"omitempty,oneof=dealer hospital store tenant"`
	Remark   string `json:"remark" validate:"omitempty,max=255"`
}

// TenantNode 租户层级树节点
type TenantNode struct {
	TenantID string        `json:"tenant_id"`
	ParentID string        `json:"parent_id"`
	Name     string        `json:"name"`
	Role     string        `json:"role"`
	Children []*TenantNode `json:"children"`
}
