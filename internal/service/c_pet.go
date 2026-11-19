package service

import (
	"time"

	dal "project/internal/dal"
	model "project/internal/model"
	"project/pkg/errcode"

	"github.com/sirupsen/logrus"
)

// CPet C 端宠物业务服务。
// C 端用户不持有后台 tenant_id，本服务先按绑定关系（c_end_user_devices）定位
// 设备所属租户，再复用 Pet 服务读取宠物业务数据，并强制校验「当前 C 端用户已绑定该设备」。
type CPet struct{}

// resolveBinding 校验当前 C 端用户已绑定目标设备，返回绑定关系（含 tenant_id / 方案）
func (*CPet) resolveBinding(cUserID, deviceID string) (*model.CEndUserDevice, error) {
	if deviceID == "" {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device_id is required",
		})
	}
	binding, err := dal.GetCEndUserDevice(cUserID, deviceID)
	if err != nil || binding == nil || binding.Status != 1 {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device not bound by current user",
		})
	}
	if binding.SolutionCode != model.SolutionCodePetMedical {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"solution_code": "device does not belong to pet solution",
		})
	}
	return binding, nil
}

// ListPets C 端宠物列表：按绑定关系聚合 宠物档案 + 设备 + 门店
func (*CPet) ListPets(cUserID string) ([]*model.CEndPetItem, error) {
	bindings, err := dal.ListCEndUserDevices(cUserID)
	if err != nil {
		logrus.Error("查询 C 端绑定设备失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_c_end_devices",
			"error":     err.Error(),
		})
	}

	result := make([]*model.CEndPetItem, 0, len(bindings))
	for _, b := range bindings {
		item := &model.CEndPetItem{
			DeviceID:     b.DeviceID,
			SolutionCode: b.SolutionCode,
			TenantID:     b.TenantID,
			BoundAt:      b.BoundAt.Format(time.RFC3339),
		}
		if device, err := dal.GetDeviceByID(b.DeviceID); err == nil && device != nil {
			item.DeviceNumber = device.DeviceNumber
			item.DeviceName = stringValue(device.Name)
		}
		if profile, err := dal.GetPetProfile(b.TenantID, b.DeviceID); err == nil && profile != nil {
			item.PetName = stringValue(profile.PetName)
			item.Breed = stringValue(profile.Breed)
			item.Age = stringValue(profile.Age)
			item.Avatar = stringValue(profile.Avatar)
			item.BoardingDays = int32Value(profile.BoardingDays)
		}
		if shop, err := dal.GetPetShop(b.TenantID, b.DeviceID); err == nil && shop != nil {
			item.ShopName = stringValue(shop.Name)
			item.OpenTime = stringValue(shop.OpenTime)
			item.CloseTime = stringValue(shop.CloseTime)
		}
		result = append(result, item)
	}
	return result, nil
}

// GetProfile C 端宠物档案 + 门店信息
func (s *CPet) GetProfile(cUserID, deviceID string) (*model.PetProfileResp, error) {
	binding, err := s.resolveBinding(cUserID, deviceID)
	if err != nil {
		return nil, err
	}
	return GroupApp.Pet.GetProfile(binding.TenantID, deviceID)
}

// GetDailyReport C 端宠物日报
func (s *CPet) GetDailyReport(cUserID, deviceID, date string) (*model.PetDailyReportResp, error) {
	binding, err := s.resolveBinding(cUserID, deviceID)
	if err != nil {
		return nil, err
	}
	return GroupApp.Pet.GetDailyReport(binding.TenantID, deviceID, date)
}

// ListStaffNotes C 端店员留言
func (s *CPet) ListStaffNotes(cUserID, deviceID string) ([]*model.PetStaffNoteResp, error) {
	binding, err := s.resolveBinding(cUserID, deviceID)
	if err != nil {
		return nil, err
	}
	return GroupApp.Pet.ListStaffNotes(binding.TenantID, deviceID)
}

// ListTimeline C 端宠物动态时间线
func (s *CPet) ListTimeline(cUserID, deviceID string, limit int) ([]*model.PetTimelineResp, error) {
	binding, err := s.resolveBinding(cUserID, deviceID)
	if err != nil {
		return nil, err
	}
	return GroupApp.Pet.ListTimeline(binding.TenantID, deviceID, limit)
}

// ListAlbum C 端宠物相册
func (s *CPet) ListAlbum(cUserID, deviceID string, limit int) ([]*model.PetAlbumResp, error) {
	binding, err := s.resolveBinding(cUserID, deviceID)
	if err != nil {
		return nil, err
	}
	return GroupApp.Pet.ListAlbum(binding.TenantID, deviceID, limit)
}

// GenerateRtcToken C 端实时音视频入会凭证（频道按设备维度固定，多位主人进入同一频道）
func (s *CPet) GenerateRtcToken(cUserID, deviceID string) (*model.PetRtcTokenResp, error) {
	binding, err := s.resolveBinding(cUserID, deviceID)
	if err != nil {
		return nil, err
	}
	userName := ""
	if user, err := dal.GetCEndUserByID(cUserID); err == nil && user != nil {
		userName = stringValue(user.Nickname)
		if userName == "" {
			userName = stringValue(user.Phone)
		}
	}
	return GroupApp.Pet.GenerateRtcTokenForDevice(binding.TenantID, deviceID, cUserID, userName)
}
