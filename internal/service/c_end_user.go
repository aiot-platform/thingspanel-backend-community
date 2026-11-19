package service

import (
	"context"
	"time"

	dal "project/internal/dal"
	model "project/internal/model"
	"project/pkg/errcode"
	global "project/pkg/global"
	utils "project/pkg/utils"

	"github.com/go-basic/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CEndUser C 端用户业务服务（注册/登录/设备绑定），鉴权与后台用户体系完全隔离。
type CEndUser struct{}

func cEndUserInfo(user *model.CEndUser) model.CEndUserInfo {
	return model.CEndUserInfo{
		ID:                  user.ID,
		Phone:               stringValue(user.Phone),
		Nickname:            stringValue(user.Nickname),
		Avatar:              stringValue(user.Avatar),
		DefaultSolutionCode: stringValue(user.DefaultSolutionCode),
	}
}

// Register C 端用户注册
func (*CEndUser) Register(req *model.CEndRegisterReq) (*model.CEndUserInfo, error) {
	if exist, err := dal.GetCEndUserByPhone(req.Phone); err == nil && exist != nil {
		return nil, errcode.WithData(errcode.CodePhoneDuplicated, map[string]interface{}{
			"phone": "phone already registered",
		})
	}

	now := time.Now().UTC()
	hash := utils.BcryptHash(req.Password)
	user := &model.CEndUser{
		ID:                  uuid.New(),
		Phone:               StringPtr(req.Phone),
		Nickname:            StringPtr(req.Nickname),
		PasswordHash:        &hash,
		DefaultSolutionCode: StringPtr(req.SolutionCode),
		Status:              1,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := dal.CreateCEndUser(user); err != nil {
		logrus.Error("C 端用户注册失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "create_c_end_user",
			"error":     err.Error(),
		})
	}
	info := cEndUserInfo(user)
	return &info, nil
}

// Login C 端用户登录，签发独立令牌
func (*CEndUser) Login(req *model.CEndLoginReq) (*model.CEndLoginResp, error) {
	user, err := dal.GetCEndUserByPhone(req.Phone)
	if err != nil || user == nil {
		return nil, errcode.New(errcode.CodeInvalidAuth)
	}
	if user.Status != 1 {
		return nil, errcode.New(errcode.CodeUserDisabled)
	}
	if user.PasswordHash == nil || !utils.BcryptCheck(req.Password, *user.PasswordHash) {
		return nil, errcode.New(errcode.CodeInvalidAuth)
	}

	key := viper.GetString("jwt.c_secret")
	if key == "" {
		key = viper.GetString("jwt.key")
	}
	if key == "" {
		return nil, errcode.WithData(errcode.CodeTokenGenerateError, map[string]interface{}{
			"error": "jwt.c_secret is empty",
		})
	}

	j := utils.NewJWT([]byte(key))
	claims := utils.CUserClaims{
		UserID:       user.ID,
		Phone:        stringValue(user.Phone),
		SolutionCode: stringValue(user.DefaultSolutionCode),
	}
	token, err := j.GenerateCToken(claims)
	if err != nil {
		return nil, errcode.WithData(errcode.CodeTokenGenerateError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	timeout := viper.GetInt("session.timeout")
	if timeout <= 0 {
		timeout = 60
	}
	if global.REDIS == nil {
		return nil, errcode.WithData(errcode.CodeTokenSaveError, map[string]interface{}{
			"error": "redis client is not initialized",
		})
	}
	if err := global.REDIS.Set(context.Background(), utils.CTokenKeyPrefix+token, "1", time.Duration(timeout)*time.Minute).Err(); err != nil {
		return nil, errcode.WithData(errcode.CodeTokenSaveError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return &model.CEndLoginResp{
		Token:     token,
		ExpiresIn: int64(timeout * 60),
		User:      cEndUserInfo(user),
	}, nil
}

// ListDevices C 端用户已绑定设备列表
func (*CEndUser) ListDevices(cUserID string) ([]*model.CEndDeviceItem, error) {
	bindings, err := dal.ListCEndUserDevices(cUserID)
	if err != nil {
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_c_end_devices",
			"error":     err.Error(),
		})
	}

	result := make([]*model.CEndDeviceItem, 0, len(bindings))
	for _, b := range bindings {
		item := &model.CEndDeviceItem{
			DeviceID:     b.DeviceID,
			SolutionCode: b.SolutionCode,
			TenantID:     b.TenantID,
			BoundAt:      b.BoundAt.Format(time.RFC3339),
		}
		if device, err := dal.GetDeviceByID(b.DeviceID); err == nil && device != nil {
			item.DeviceNumber = device.DeviceNumber
			item.DeviceName = stringValue(device.Name)
		}
		result = append(result, item)
	}
	return result, nil
}

// BindDevice C 端用户凭门店绑定码绑定设备（多对多，一台设备可被多人共享）。
// 校验绑定码有效性（状态 / 有效期 / 剩余次数）与门店归属，成功后累加已用次数。
func (*CEndUser) BindDevice(cUserID string, req *model.CEndBindDeviceReq) (*model.CEndDeviceItem, error) {
	code, err := dal.GetDeviceBindCodeByCode(req.Code)
	if err != nil || code == nil {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"code": "bind code not found",
		})
	}
	if code.Status != 1 {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"code": "bind code is disabled",
		})
	}
	if code.ExpiresAt != nil && code.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"code": "bind code has expired",
		})
	}
	if code.UsedCount >= code.MaxUses {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"code": "bind code has reached its usage limit",
		})
	}

	device, err := dal.GetDeviceByID(code.DeviceID)
	if err != nil || device == nil {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device not found",
		})
	}
	if device.TenantID != code.TenantID {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device no longer belongs to the issuing tenant",
		})
	}

	if exist, err := dal.GetCEndUserDevice(cUserID, code.DeviceID); err == nil && exist != nil {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device already bound by current user",
		})
	}

	item := &model.CEndUserDevice{
		ID:           uuid.New(),
		CUserID:      cUserID,
		DeviceID:     code.DeviceID,
		TenantID:     code.TenantID,
		SolutionCode: resolveTenantSolution(code.TenantID),
		Status:       1,
	}
	if err := dal.CreateCEndUserDevice(item); err != nil {
		logrus.Error("C 端绑定设备失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "bind_c_end_device",
			"error":     err.Error(),
		})
	}
	if err := dal.IncrementBindCodeUsed(code.ID); err != nil {
		// 绑定已成功，累加失败仅记录日志，不阻断
		logrus.Error("累加绑定码已用次数失败:", err)
	}

	return &model.CEndDeviceItem{
		DeviceID:     code.DeviceID,
		DeviceNumber: device.DeviceNumber,
		DeviceName:   stringValue(device.Name),
		SolutionCode: item.SolutionCode,
		TenantID:     code.TenantID,
		BoundAt:      item.BoundAt.Format(time.RFC3339),
	}, nil
}

// resolveTenantSolution 推导租户用于 C 端绑定的方案编码：优先宠物医疗，其次任一启用方案。
func resolveTenantSolution(tenantID string) string {
	list, err := dal.ListTenantSolutions(tenantID)
	if err == nil {
		for _, ts := range list {
			if ts.Enabled == 1 && ts.SolutionCode == model.SolutionCodePetMedical {
				return ts.SolutionCode
			}
		}
		for _, ts := range list {
			if ts.Enabled == 1 {
				return ts.SolutionCode
			}
		}
	}
	return model.SolutionCodePetMedical
}

// UnbindDevice C 端用户解绑设备
func (*CEndUser) UnbindDevice(cUserID, deviceID string) error {
	if err := dal.DeleteCEndUserDevice(cUserID, deviceID); err != nil {
		logrus.Error("C 端解绑设备失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "unbind_c_end_device",
			"error":     err.Error(),
		})
	}
	return nil
}
