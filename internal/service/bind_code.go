package service

import (
	"crypto/rand"
	"math/big"
	"time"

	dal "project/internal/dal"
	model "project/internal/model"
	"project/pkg/errcode"

	"github.com/go-basic/uuid"
	"github.com/sirupsen/logrus"
)

// BindCode 门店绑定码业务服务（B 端门店出码 / 失效 / 查询已用次数）。
type BindCode struct{}

const (
	defaultBindCodeMaxUses = 1
	bindCodeAlphabet       = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去除易混淆字符
	bindCodeLength         = 8
)

// generateBindCode 生成随机绑定码
func generateBindCode() string {
	buf := make([]byte, bindCodeLength)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(bindCodeAlphabet))))
		if err != nil {
			// 退化到时间戳，保证不 panic（极低概率）
			buf[i] = bindCodeAlphabet[int(time.Now().UnixNano()+int64(i))%len(bindCodeAlphabet)]
			continue
		}
		buf[i] = bindCodeAlphabet[n.Int64()]
	}
	return string(buf)
}

func bindCodeResp(item *model.DeviceBindCode) *model.DeviceBindCodeResp {
	resp := &model.DeviceBindCodeResp{
		ID:        item.ID,
		DeviceID:  item.DeviceID,
		Code:      item.Code,
		MaxUses:   item.MaxUses,
		UsedCount: item.UsedCount,
		Status:    item.Status,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
	if item.ExpiresAt != nil {
		resp.ExpiresAt = item.ExpiresAt.Format(time.RFC3339)
	}
	if device, err := dal.GetDeviceByID(item.DeviceID); err == nil && device != nil {
		resp.DeviceName = stringValue(device.Name)
	}
	return resp
}

// Create 生成绑定码：校验设备归属本租户，按需设置有效期与最大使用次数
func (*BindCode) Create(req *model.DeviceBindCodeCreateReq, tenantID, createdBy string) (*model.DeviceBindCodeResp, error) {
	device, err := dal.GetDeviceByID(req.DeviceID)
	if err != nil || device == nil {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device not found",
		})
	}
	if device.TenantID != tenantID {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device does not belong to current tenant",
		})
	}

	maxUses := req.MaxUses
	if maxUses <= 0 {
		maxUses = defaultBindCodeMaxUses
	}
	var expiresAt *time.Time
	if req.ExpiresInMinutes > 0 {
		t := time.Now().UTC().Add(time.Duration(req.ExpiresInMinutes) * time.Minute)
		expiresAt = &t
	}

	item := &model.DeviceBindCode{
		ID:        uuid.New(),
		TenantID:  tenantID,
		DeviceID:  req.DeviceID,
		Code:      generateBindCode(),
		ExpiresAt: expiresAt,
		MaxUses:   maxUses,
		UsedCount: 0,
		Status:    1,
		CreatedBy: StringPtr(createdBy),
	}
	if err := dal.CreateDeviceBindCode(item); err != nil {
		logrus.Error("生成绑定码失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "create_device_bind_code",
			"error":     err.Error(),
		})
	}
	return bindCodeResp(item), nil
}

// List 查询本租户绑定码
func (*BindCode) List(tenantID, deviceID string) ([]*model.DeviceBindCodeResp, error) {
	list, err := dal.ListDeviceBindCodes(tenantID, deviceID)
	if err != nil {
		logrus.Error("查询绑定码失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_device_bind_codes",
			"error":     err.Error(),
		})
	}
	result := make([]*model.DeviceBindCodeResp, 0, len(list))
	for _, item := range list {
		result = append(result, bindCodeResp(item))
	}
	return result, nil
}

// SetStatus 启用/失效绑定码（仅限本租户）
func (*BindCode) SetStatus(req *model.DeviceBindCodeStatusReq, tenantID string) error {
	exist, err := dal.GetDeviceBindCodeByID(req.ID)
	if err != nil || exist == nil || exist.TenantID != tenantID {
		return errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"id": "bind code not found",
		})
	}
	if err := dal.UpdateDeviceBindCodeStatus(req.ID, tenantID, req.Status); err != nil {
		logrus.Error("更新绑定码状态失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "update_device_bind_code_status",
			"error":     err.Error(),
		})
	}
	return nil
}
