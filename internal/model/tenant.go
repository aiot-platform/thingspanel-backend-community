package model

import "time"

// 门店绑定码与租户层级平台级数据表（public schema）：
//   - tenants：租户层级主表（经销商/宠物医院/门店），叠加在 users.tenant_id 之上；
//   - device_bind_codes：门店出码，家长凭码绑定设备。

const (
	TableNameTenant         = "tenants"
	TableNameDeviceBindCode = "device_bind_codes"
)

// 租户角色
const (
	TenantRoleDealer   = "dealer"   // 经销商
	TenantRoleHospital = "hospital" // 宠物医院
	TenantRoleStore    = "store"    // 门店
)

// Tenant 租户层级节点
type Tenant struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	ParentID  *string   `gorm:"column:parent_id" json:"parent_id"`
	Name      *string   `gorm:"column:name" json:"name"`
	Role      *string   `gorm:"column:role" json:"role"`
	Status    int16     `gorm:"column:status;not null" json:"status"` // 1-启用 0-停用
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark    *string   `gorm:"column:remark" json:"remark"`
}

func (*Tenant) TableName() string { return TableNameTenant }

// DeviceBindCode 设备绑定码
type DeviceBindCode struct {
	ID        string     `gorm:"column:id;primaryKey" json:"id"`
	TenantID  string     `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID  string     `gorm:"column:device_id;not null" json:"device_id"`
	Code      string     `gorm:"column:code;not null" json:"code"`
	ExpiresAt *time.Time `gorm:"column:expires_at" json:"expires_at"`
	MaxUses   int32      `gorm:"column:max_uses;not null" json:"max_uses"`
	UsedCount int32      `gorm:"column:used_count;not null" json:"used_count"`
	Status    int16      `gorm:"column:status;not null" json:"status"` // 1-有效 0-失效
	CreatedBy *string    `gorm:"column:created_by" json:"created_by"`
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (*DeviceBindCode) TableName() string { return TableNameDeviceBindCode }
