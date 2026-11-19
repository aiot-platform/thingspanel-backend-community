package model

import "time"

// C 端用户数据表（独立于后台 users 表）：
//   - c_end_users：C 端用户身份（平台级，可跨方案）；
//   - c_end_user_devices：C 端用户↔设备 绑定关系（多对多，一台设备可被多人共享）。

const (
	TableNameCEndUser       = "c_end_users"
	TableNameCEndUserDevice = "c_end_user_devices"
)

// CEndUser C 端用户
type CEndUser struct {
	ID                  string    `gorm:"column:id;primaryKey" json:"id"`
	Phone               *string   `gorm:"column:phone" json:"phone"`
	OpenID              *string   `gorm:"column:open_id" json:"open_id"`
	Nickname            *string   `gorm:"column:nickname" json:"nickname"`
	Avatar              *string   `gorm:"column:avatar" json:"avatar"`
	PasswordHash        *string   `gorm:"column:password_hash" json:"-"`
	DefaultSolutionCode *string   `gorm:"column:default_solution_code" json:"default_solution_code"`
	Status              int16     `gorm:"column:status;not null" json:"status"` // 1-正常 0-禁用
	CreatedAt           time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark              *string   `gorm:"column:remark" json:"remark"`
}

func (*CEndUser) TableName() string { return TableNameCEndUser }

// CEndUserDevice C 端用户↔设备 绑定关系
type CEndUserDevice struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	CUserID      string    `gorm:"column:c_user_id;not null" json:"c_user_id"`
	DeviceID     string    `gorm:"column:device_id;not null" json:"device_id"`
	TenantID     string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	SolutionCode string    `gorm:"column:solution_code;not null" json:"solution_code"`
	Status       int16     `gorm:"column:status;not null" json:"status"` // 1-有效 0-解绑
	BoundAt      time.Time `gorm:"column:bound_at;not null" json:"bound_at"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"created_at"`
}

func (*CEndUserDevice) TableName() string { return TableNameCEndUserDevice }
