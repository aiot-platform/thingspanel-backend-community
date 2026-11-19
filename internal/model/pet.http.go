package model

// 宠物看护接口请求/响应结构体
//
// 字段命名约定：主字段使用 snake_case（与 app/README.md 的云端交互契约一致），
// 对 uni-app 与微信小程序取值不一致的字段额外返回 camelCase 别名，保证两端
// 页面无需改动即可消费（例如 sleep_minutes 与 sleep_hours、title 与 text）。

// PetRtcTokenReq 签发实时音视频入会凭证
type PetRtcTokenReq struct {
	DeviceID string `json:"device_id" validate:"required,max=36"`
}

// PetRtcTokenResp 阿里云 ARTC 单参数入会凭证
type PetRtcTokenResp struct {
	AppID     string `json:"app_id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// PetDeviceReq 只按设备查询的公共请求
type PetDeviceReq struct {
	DeviceID string `json:"device_id" form:"device_id" validate:"required,max=36"`
}

// PetDailyReportReq 日报查询
type PetDailyReportReq struct {
	DeviceID string `json:"device_id" form:"device_id" validate:"required,max=36"`
	Date     string `json:"date" form:"date" validate:"omitempty,max=20"`
}

// PetTimelineListReq 时间线查询
type PetTimelineListReq struct {
	DeviceID string `json:"device_id" form:"device_id" validate:"required,max=36"`
	Limit    int    `json:"limit" form:"limit" validate:"omitempty,min=1,max=200"`
}

// PetAlbumListReq 相册查询
type PetAlbumListReq struct {
	DeviceID string `json:"device_id" form:"device_id" validate:"required,max=36"`
	Limit    int    `json:"limit" form:"limit" validate:"omitempty,min=1,max=200"`
}

// PetProfilePet 宠物档案
type PetProfilePet struct {
	Name         string `json:"name"`
	Breed        string `json:"breed"`
	Age          string `json:"age"`
	Avatar       string `json:"avatar"`
	BoardingDays int32  `json:"boarding_days"`
	AliasDays    int32  `json:"boardingDays"` // 兼容 uni-app 取值
}

// PetProfileShop 店铺信息与营业时段
type PetProfileShop struct {
	Name          string `json:"name"`
	OpenTime      string `json:"open_time"`
	AliasOpenTime string `json:"openTime"` // 兼容 uni-app 取值
	CloseTime     string `json:"close_time"`
	AliasClose    string `json:"closeTime"` // 兼容 uni-app 取值
}

// PetProfileResp 宠物档案 + 店铺信息
type PetProfileResp struct {
	Pet  PetProfilePet  `json:"pet"`
	Shop PetProfileShop `json:"shop"`
}

// PetSaveProfileReq 保存宠物档案与店铺信息
type PetSaveProfileReq struct {
	DeviceID     string `json:"device_id" validate:"required,max=36"`
	PetName      string `json:"pet_name" validate:"omitempty,max=99"`
	Breed        string `json:"breed" validate:"omitempty,max=99"`
	Age          string `json:"age" validate:"omitempty,max=50"`
	Avatar       string `json:"avatar" validate:"omitempty,max=999"`
	BoardingDays int32  `json:"boarding_days" validate:"omitempty,min=0"`
	ShopName     string `json:"shop_name" validate:"omitempty,max=255"`
	OpenTime     string `json:"open_time" validate:"omitempty,max=10"`
	CloseTime    string `json:"close_time" validate:"omitempty,max=10"`
}

// PetDailyReportResp 今日日报
type PetDailyReportResp struct {
	EatCount     int32   `json:"eat_count"`
	SleepMinutes int32   `json:"sleep_minutes"`
	AliasSleepH  float64 `json:"sleep_hours"` // 兼容微信小程序取值（分钟换算为小时）
	PlayMinutes  int32   `json:"play_minutes"`
	DrinkCount   int32   `json:"drink_count"`
}

// PetStaffNoteResp 店员留言
type PetStaffNoteResp struct {
	ID            string `json:"id"`
	Content       string `json:"content"`
	StaffName     string `json:"staff_name"`
	AliasName     string `json:"staffName"` // 兼容 uni-app / 小程序取值
	CreateAt      string `json:"create_at"`
	AliasCreateAt string `json:"time"` // 兼容 uni-app / 小程序取值
}

// PetSaveStaffNoteReq 新增/修改店员留言
type PetSaveStaffNoteReq struct {
	ID        string `json:"id" validate:"omitempty,max=36"`
	DeviceID  string `json:"device_id" validate:"required,max=36"`
	Content   string `json:"content" validate:"required,max=2000"`
	StaffName string `json:"staff_name" validate:"omitempty,max=99"`
}

// PetTimelineResp 动态时间线
type PetTimelineResp struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	AliasTitle    string `json:"text"` // 兼容微信小程序取值
	CreateAt      string `json:"create_at"`
	AliasCreateAt string `json:"time"` // 兼容 uni-app / 小程序取值
}

// PetSaveTimelineReq 新增动态
type PetSaveTimelineReq struct {
	DeviceID string `json:"device_id" validate:"required,max=36"`
	Type     string `json:"type" validate:"omitempty,max=20"`
	Title    string `json:"title" validate:"required,max=255"`
}

// PetAlbumResp 相册记录
type PetAlbumResp struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	AliasURL      string `json:"src"` // 兼容 uni-app / 小程序取值
	CreateAt      string `json:"create_at"`
	AliasCreateAt string `json:"time"` // 兼容 uni-app / 小程序取值
}

// PetSaveAlbumReq 新增相册记录
type PetSaveAlbumReq struct {
	DeviceID string `json:"device_id" validate:"required,max=36"`
	URL      string `json:"url" validate:"required,max=999"`
}

// PetDeleteReq 按 id 删除
type PetDeleteReq struct {
	ID string `json:"id" validate:"required,max=36"`
}
