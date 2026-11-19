package model

import "time"

// 行业方案（Industry Solution）平台级数据表：
//   - industry_solutions：方案字典，由平台超级管理员维护；
//   - tenant_solutions：方案↔租户 分配关系（多对多）。
//
// 方案业务表存放在各自独立 schema（见 Schema* 常量），平台级表存放在 public。

// 方案编码（对外稳定标识）
const (
	SolutionCodePetMedical = "pet_medical"
	SolutionCodeYuezi      = "yuezi"
	SolutionCodeTuoyu      = "tuoyu"
	SolutionCodeXumu       = "xumu"
	SolutionCodeGongye     = "gongye"
)

// 方案 schema 名（须匹配 ^sol_[a-z0-9_]+$）
const (
	SchemaPetMedical = "sol_pet_medical"
	SchemaYuezi      = "sol_yuezi"
	SchemaTuoyu      = "sol_tuoyu"
	SchemaXumu       = "sol_xumu"
	SchemaGongye     = "sol_gongye"
)

const (
	TableNameIndustrySolution = "industry_solutions"
	TableNameTenantSolution   = "tenant_solutions"
)

// IndustrySolution 行业方案字典
type IndustrySolution struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	Code        string    `gorm:"column:code;not null" json:"code"`
	Name        string    `gorm:"column:name;not null" json:"name"`
	Description *string   `gorm:"column:description" json:"description"`
	SchemaName  string    `gorm:"column:schema_name;not null" json:"schema_name"`
	Icon        *string   `gorm:"column:icon" json:"icon"`
	Status      int16     `gorm:"column:status;not null" json:"status"` // 1-启用 0-停用
	Sort        int32     `gorm:"column:sort" json:"sort"`
	CreatedAt   time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark      *string   `gorm:"column:remark" json:"remark"`
}

func (*IndustrySolution) TableName() string { return TableNameIndustrySolution }

// TenantSolution 方案↔租户 分配关系
type TenantSolution struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	SolutionCode string    `gorm:"column:solution_code;not null" json:"solution_code"`
	Enabled      int16     `gorm:"column:enabled;not null" json:"enabled"` // 1-启用 0-停用
	Config       *string   `gorm:"column:config" json:"config"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (*TenantSolution) TableName() string { return TableNameTenantSolution }
