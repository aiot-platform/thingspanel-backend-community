package model

import (
	"encoding/json"
	"time"
)

// GetDeviceLogReq 设备统一日志检索请求
type GetDeviceLogReq struct {
	PageReq
	DeviceID  string `json:"device_id" form:"device_id" validate:"required,max=36"`         // 设备ID
	MsgType   string `json:"msg_type" form:"msg_type" validate:"omitempty,max=32"`          // 消息类型，见 devicelog.Type*
	Direction string `json:"direction" form:"direction" validate:"omitempty,oneof=up down"` // 方向：up-上行 down-下行
	StartTime *int64 `json:"start_time" form:"start_time" validate:"omitempty"`             // 开始时间戳（秒）
	EndTime   *int64 `json:"end_time" form:"end_time" validate:"omitempty"`                 // 结束时间戳（秒）
	Keyword   string `json:"keyword" form:"keyword" validate:"omitempty,max=100"`           // 关键字：匹配报文/消息ID
	TenantID  string `json:"tenant_id" form:"tenant_id" validate:"omitempty,max=36"`        // 租户ID：仅系统管理员生效且必填，租户侧忽略
}

// GetDeviceLogDevicesReq 设备日志查询：按租户检索设备下拉（系统管理员）
type GetDeviceLogDevicesReq struct {
	TenantID string `json:"tenant_id" form:"tenant_id" validate:"required,max=36"` // 租户ID
	Keyword  string `json:"keyword" form:"keyword" validate:"omitempty,max=100"`   // 设备名称/编号关键字
}

// DeviceLogTenantOption 租户下拉项
type DeviceLogTenantOption struct {
	TenantID string  `json:"tenant_id" gorm:"column:tenant_id"`
	Name     *string `json:"name" gorm:"column:name"`
	Email    *string `json:"email" gorm:"column:email"`
}

// DeviceLogDeviceOption 设备下拉项
type DeviceLogDeviceOption struct {
	ID           string  `json:"id" gorm:"column:id"`
	Name         *string `json:"name" gorm:"column:name"`
	DeviceNumber string  `json:"device_number" gorm:"column:device_number"`
}

// DeviceLogItem 设备统一日志条目
type DeviceLogItem struct {
	TS          time.Time       `json:"ts" gorm:"column:ts"`
	MsgType     string          `json:"msg_type" gorm:"column:msg_type"`
	Direction   string          `json:"direction" gorm:"column:direction"`
	MessageID   *string         `json:"message_id" gorm:"column:message_id"`
	Status      *string         `json:"status" gorm:"column:status"`
	Error       *string         `json:"error" gorm:"column:error"`
	PayloadText *string         `json:"-" gorm:"column:payload"`
	Payload     json.RawMessage `json:"payload" gorm:"-"`
}
