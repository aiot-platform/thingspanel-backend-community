package model

// 行业方案接口请求/响应结构体

// IndustrySolutionSaveReq 新增/修改方案
type IndustrySolutionSaveReq struct {
	ID          string `json:"id" validate:"omitempty,max=36"`
	Code        string `json:"code" validate:"required,max=64"`
	Name        string `json:"name" validate:"required,max=99"`
	Description string `json:"description" validate:"omitempty,max=500"`
	SchemaName  string `json:"schema_name" validate:"required,max=63"`
	Icon        string `json:"icon" validate:"omitempty,max=99"`
	Status      int16  `json:"status" validate:"omitempty,oneof=0 1"`
	Sort        int32  `json:"sort" validate:"omitempty,min=0"`
	Remark      string `json:"remark" validate:"omitempty,max=255"`
}

// TenantSolutionAssignReq 给租户分配方案
type TenantSolutionAssignReq struct {
	TenantID     string `json:"tenant_id" validate:"required,max=36"`
	SolutionCode string `json:"solution_code" validate:"required,max=64"`
	Enabled      int16  `json:"enabled" validate:"omitempty,oneof=0 1"`
	Cascade      bool   `json:"cascade"` // 是否沿租户层级下放给全部下级租户
}

// TenantSolutionQueryReq 查询某租户已分配方案
type TenantSolutionQueryReq struct {
	TenantID string `json:"tenant_id" form:"tenant_id" validate:"required,max=36"`
}

// TenantInfo 租户信息（用于分配界面）
type TenantInfo struct {
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// TenantSolutionResp 租户已分配方案
type TenantSolutionResp struct {
	SolutionCode string `json:"solution_code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	SchemaName   string `json:"schema_name"`
	Enabled      int16  `json:"enabled"`
	// DeviceCount 该方案在本租户下的设备/业务数据量概览
	DeviceCount int64 `json:"device_count"`
}
