package model

import "time"

// 宠物看护相关数据表（宠物档案、店铺、店员留言、动态时间线、相册、日报）

const (
	TableNamePetProfile     = "pet_profiles"
	TableNamePetShop        = "pet_shops"
	TableNamePetStaffNote   = "pet_staff_notes"
	TableNamePetTimeline    = "pet_timelines"
	TableNamePetAlbum       = "pet_albums"
	TableNamePetDailyReport = "pet_daily_reports"
)

// PetProfile 宠物档案
type PetProfile struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID     string    `gorm:"column:device_id;not null" json:"device_id"`
	PetName      *string   `gorm:"column:pet_name" json:"pet_name"`
	Breed        *string   `gorm:"column:breed" json:"breed"`
	Age          *string   `gorm:"column:age" json:"age"`
	Avatar       *string   `gorm:"column:avatar" json:"avatar"`
	BoardingDays *int32    `gorm:"column:boarding_days" json:"boarding_days"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark       *string   `gorm:"column:remark" json:"remark"`
}

func (*PetProfile) TableName() string { return TableNamePetProfile }

// PetShop 宠物店铺信息与营业时段
type PetShop struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID  string    `gorm:"column:device_id;not null" json:"device_id"`
	Name      *string   `gorm:"column:name" json:"name"`
	OpenTime  *string   `gorm:"column:open_time" json:"open_time"`
	CloseTime *string   `gorm:"column:close_time" json:"close_time"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark    *string   `gorm:"column:remark" json:"remark"`
}

func (*PetShop) TableName() string { return TableNamePetShop }

// PetStaffNote 店员留言
type PetStaffNote struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID  string    `gorm:"column:device_id;not null" json:"device_id"`
	Content   *string   `gorm:"column:content" json:"content"`
	StaffName *string   `gorm:"column:staff_name" json:"staff_name"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark    *string   `gorm:"column:remark" json:"remark"`
}

func (*PetStaffNote) TableName() string { return TableNamePetStaffNote }

// PetTimeline 宠物动态时间线
type PetTimeline struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID  string    `gorm:"column:device_id;not null" json:"device_id"`
	EventType *string   `gorm:"column:event_type" json:"event_type"`
	Title     *string   `gorm:"column:title" json:"title"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark    *string   `gorm:"column:remark" json:"remark"`
}

func (*PetTimeline) TableName() string { return TableNamePetTimeline }

// PetAlbum 宠物相册（抓拍）
type PetAlbum struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID  string    `gorm:"column:device_id;not null" json:"device_id"`
	URL       *string   `gorm:"column:url" json:"url"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark    *string   `gorm:"column:remark" json:"remark"`
}

func (*PetAlbum) TableName() string { return TableNamePetAlbum }

// PetDailyReport 宠物日报
type PetDailyReport struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DeviceID     string    `gorm:"column:device_id;not null" json:"device_id"`
	ReportDate   time.Time `gorm:"column:report_date;not null" json:"report_date"`
	EatCount     *int32    `gorm:"column:eat_count" json:"eat_count"`
	SleepMinutes *int32    `gorm:"column:sleep_minutes" json:"sleep_minutes"`
	PlayMinutes  *int32    `gorm:"column:play_minutes" json:"play_minutes"`
	DrinkCount   *int32    `gorm:"column:drink_count" json:"drink_count"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Remark       *string   `gorm:"column:remark" json:"remark"`
}

func (*PetDailyReport) TableName() string { return TableNamePetDailyReport }
