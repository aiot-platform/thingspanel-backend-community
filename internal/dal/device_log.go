package dal

import (
	"strings"
	"time"

	"project/internal/devicelog"
	"project/internal/model"
	global "project/pkg/global"

	"gorm.io/gorm"
)

// GetDeviceLogByPage 分页检索设备统一日志（按时间倒序），按租户隔离
func GetDeviceLogByPage(req *model.GetDeviceLogReq, tenantID string) (int64, []*model.DeviceLogItem, error) {
	db := global.DB.Table(devicelog.TableName).
		Where("tenant_id = ? AND device_id = ?", tenantID, req.DeviceID)

	if req.MsgType != "" {
		db = db.Where("msg_type = ?", req.MsgType)
	}
	if req.Direction != "" {
		db = db.Where("direction = ?", req.Direction)
	}
	if req.StartTime != nil {
		db = db.Where("ts >= ?", time.Unix(*req.StartTime, 0))
	}
	if req.EndTime != nil {
		db = db.Where("ts <= ?", time.Unix(*req.EndTime, 0))
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + escapeLike(kw) + "%"
		db = db.Where("(payload::text ILIKE ? OR message_id ILIKE ?)", like, like)
	}

	// Session 使链可安全复用于 Count 与 Find
	db = db.Session(&gorm.Session{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	list := make([]*model.DeviceLogItem, 0)
	err := db.Select("ts, msg_type, direction, message_id, status, error, payload::text AS payload").
		Order("ts DESC").
		Limit(req.PageSize).
		Offset((req.Page - 1) * req.PageSize).
		Scan(&list).Error
	if err != nil {
		return 0, nil, err
	}
	return total, list, nil
}

// GetDeviceLogTenants 租户下拉：所有租户管理员对应的租户
func GetDeviceLogTenants() ([]*model.DeviceLogTenantOption, error) {
	list := make([]*model.DeviceLogTenantOption, 0)
	err := global.DB.Table("users").
		Select("tenant_id, name, email").
		Where("authority = ? AND tenant_id IS NOT NULL AND tenant_id <> ''", TENANT_ADMIN).
		Order("created_at DESC").
		Scan(&list).Error
	return list, err
}

// GetDeviceLogDevices 设备下拉：指定租户下的设备，按名称/编号模糊匹配
func GetDeviceLogDevices(tenantID, keyword string, limit int) ([]*model.DeviceLogDeviceOption, error) {
	db := global.DB.Table("devices").
		Select("id, name, device_number").
		Where("tenant_id = ?", tenantID)
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + escapeLike(kw) + "%"
		db = db.Where("(name ILIKE ? OR device_number ILIKE ?)", like, like)
	}
	list := make([]*model.DeviceLogDeviceOption, 0)
	err := db.Order("created_at DESC NULLS LAST").Limit(limit).Scan(&list).Error
	return list, err
}

// escapeLike 转义 LIKE 通配符，关键字按字面匹配
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
