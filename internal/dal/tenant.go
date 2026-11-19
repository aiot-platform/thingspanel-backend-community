package dal

import (
	"time"

	model "project/internal/model"
	global "project/pkg/global"

	"github.com/sirupsen/logrus"
)

// 门店绑定码与租户层级数据访问层（public schema）。

// ---------------------------------------------------------------------------
// 租户层级
// ---------------------------------------------------------------------------

func GetTenant(id string) (*model.Tenant, error) {
	item := &model.Tenant{}
	err := global.DB.Model(&model.Tenant{}).Where("id = ?", id).First(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func ListTenants() ([]*model.Tenant, error) {
	var list []*model.Tenant
	err := global.DB.Model(&model.Tenant{}).Order("created_at ASC").Find(&list).Error
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return list, nil
}

// UpsertTenant 按 id 写入或更新租户层级节点
func UpsertTenant(item *model.Tenant) error {
	now := time.Now().UTC()
	exist, err := GetTenant(item.ID)
	if err != nil {
		item.CreatedAt = now
		item.UpdatedAt = now
		return global.DB.Model(&model.Tenant{}).Create(item).Error
	}
	item.CreatedAt = exist.CreatedAt
	item.UpdatedAt = now
	return global.DB.Model(&model.Tenant{}).
		Where("id = ?", exist.ID).
		Updates(map[string]interface{}{
			"parent_id":  item.ParentID,
			"name":       item.Name,
			"role":       item.Role,
			"status":     item.Status,
			"remark":     item.Remark,
			"updated_at": now,
		}).Error
}

// ListDescendantTenantIDs 返回包含自身在内的全部下级租户 id（递归 CTE）
func ListDescendantTenantIDs(tenantID string) ([]string, error) {
	var ids []string
	err := global.DB.Raw(`
		WITH RECURSIVE t AS (
			SELECT id FROM public.tenants WHERE id = ?
			UNION ALL
			SELECT c.id FROM public.tenants c JOIN t ON c.parent_id = t.id
		)
		SELECT id FROM t`, tenantID).Scan(&ids).Error
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return ids, nil
}

// ---------------------------------------------------------------------------
// 设备绑定码
// ---------------------------------------------------------------------------

func CreateDeviceBindCode(item *model.DeviceBindCode) error {
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	return global.DB.Model(&model.DeviceBindCode{}).Create(item).Error
}

func GetDeviceBindCodeByCode(code string) (*model.DeviceBindCode, error) {
	item := &model.DeviceBindCode{}
	err := global.DB.Model(&model.DeviceBindCode{}).Where("code = ?", code).First(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func GetDeviceBindCodeByID(id string) (*model.DeviceBindCode, error) {
	item := &model.DeviceBindCode{}
	err := global.DB.Model(&model.DeviceBindCode{}).Where("id = ?", id).First(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

// ListDeviceBindCodes 查询本租户的绑定码，deviceID 非空时按设备过滤
func ListDeviceBindCodes(tenantID, deviceID string) ([]*model.DeviceBindCode, error) {
	var list []*model.DeviceBindCode
	q := global.DB.Model(&model.DeviceBindCode{}).Where("tenant_id = ?", tenantID)
	if deviceID != "" {
		q = q.Where("device_id = ?", deviceID)
	}
	err := q.Order("created_at DESC").Find(&list).Error
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return list, nil
}

// IncrementBindCodeUsed 原子累加已用次数（并发下以 used_count < max_uses 为守卫）
func IncrementBindCodeUsed(id string) error {
	return global.DB.Exec(
		`UPDATE public.device_bind_codes SET used_count = used_count + 1, updated_at = ? WHERE id = ?`,
		time.Now().UTC(), id).Error
}

func UpdateDeviceBindCodeStatus(id, tenantID string, status int16) error {
	return global.DB.Model(&model.DeviceBindCode{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}).Error
}
