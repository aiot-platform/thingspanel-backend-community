package service

import (
	"time"

	dal "project/internal/dal"
	model "project/internal/model"
	"project/pkg/errcode"

	"github.com/go-basic/uuid"
	"github.com/sirupsen/logrus"
)

// Pet 宠物看护业务服务（档案/店铺/店员留言/动态时间线/相册/日报）。
type Pet struct{}

// checkDevice 校验设备存在且属于当前租户
func (*Pet) checkDevice(deviceID, tenantID string) error {
	if deviceID == "" {
		return errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device_id is required",
		})
	}
	device, err := dal.GetDeviceByID(deviceID)
	if err != nil || device == nil {
		return errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device not found",
		})
	}
	if device.TenantID != tenantID {
		return errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"device_id": "device does not belong to current tenant",
		})
	}
	return nil
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int32Value(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// GetProfile 查询宠物档案与店铺信息（PRD 4.2 / 4.1 营业时段）
func (p *Pet) GetProfile(tenantID, deviceID string) (*model.PetProfileResp, error) {
	if err := p.checkDevice(deviceID, tenantID); err != nil {
		return nil, err
	}

	resp := &model.PetProfileResp{}

	profile, err := dal.GetPetProfile(tenantID, deviceID)
	if err != nil {
		logrus.Warn("查询宠物档案失败，返回空档案:", err)
	} else {
		resp.Pet = model.PetProfilePet{
			Name:         stringValue(profile.PetName),
			Breed:        stringValue(profile.Breed),
			Age:          stringValue(profile.Age),
			Avatar:       stringValue(profile.Avatar),
			BoardingDays: int32Value(profile.BoardingDays),
			AliasDays:    int32Value(profile.BoardingDays),
		}
	}

	shop, err := dal.GetPetShop(tenantID, deviceID)
	if err != nil {
		logrus.Warn("查询店铺信息失败，返回空店铺:", err)
	} else {
		resp.Shop = model.PetProfileShop{
			Name:          stringValue(shop.Name),
			OpenTime:      stringValue(shop.OpenTime),
			AliasOpenTime: stringValue(shop.OpenTime),
			CloseTime:     stringValue(shop.CloseTime),
			AliasClose:    stringValue(shop.CloseTime),
		}
	}

	return resp, nil
}

// SaveProfile 保存宠物档案与店铺信息
func (p *Pet) SaveProfile(req *model.PetSaveProfileReq, tenantID string) error {
	if err := p.checkDevice(req.DeviceID, tenantID); err != nil {
		return err
	}

	profile := &model.PetProfile{
		ID:           uuid.New(),
		TenantID:     tenantID,
		DeviceID:     req.DeviceID,
		PetName:      StringPtr(req.PetName),
		Breed:        StringPtr(req.Breed),
		Age:          StringPtr(req.Age),
		Avatar:       StringPtr(req.Avatar),
		BoardingDays: Int32Ptr(req.BoardingDays),
	}
	if err := dal.UpsertPetProfile(profile); err != nil {
		logrus.Error("保存宠物档案失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "save_pet_profile",
			"error":     err.Error(),
		})
	}

	shop := &model.PetShop{
		ID:        uuid.New(),
		TenantID:  tenantID,
		DeviceID:  req.DeviceID,
		Name:      StringPtr(req.ShopName),
		OpenTime:  StringPtr(req.OpenTime),
		CloseTime: StringPtr(req.CloseTime),
	}
	if err := dal.UpsertPetShop(shop); err != nil {
		logrus.Error("保存店铺信息失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "save_pet_shop",
			"error":     err.Error(),
		})
	}

	return nil
}

// GetDailyReport 查询日报；无记录时返回全 0，由调用方按需隐藏
func (p *Pet) GetDailyReport(tenantID, deviceID, date string) (*model.PetDailyReportResp, error) {
	if err := p.checkDevice(deviceID, tenantID); err != nil {
		return nil, err
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	report, err := dal.GetPetDailyReport(tenantID, deviceID, date)
	if err != nil {
		logrus.Warn("查询宠物日报失败，返回空日报:", err)
		return &model.PetDailyReportResp{}, nil
	}

	sleepMinutes := int32Value(report.SleepMinutes)
	return &model.PetDailyReportResp{
		EatCount:     int32Value(report.EatCount),
		SleepMinutes: sleepMinutes,
		AliasSleepH:  sleepHours(sleepMinutes),
		PlayMinutes:  int32Value(report.PlayMinutes),
		DrinkCount:   int32Value(report.DrinkCount),
	}, nil
}

// sleepHours 分钟转小时，保留一位小数（微信小程序端取 sleep_hours）
func sleepHours(minutes int32) float64 {
	return float64(minutes*10/6) / 10
}

// ListStaffNotes 店员留言列表
func (p *Pet) ListStaffNotes(tenantID, deviceID string) ([]*model.PetStaffNoteResp, error) {
	if err := p.checkDevice(deviceID, tenantID); err != nil {
		return nil, err
	}

	notes, err := dal.ListPetStaffNotes(tenantID, deviceID)
	if err != nil {
		logrus.Error("查询店员留言失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_pet_staff_notes",
			"error":     err.Error(),
		})
	}

	result := make([]*model.PetStaffNoteResp, 0, len(notes))
	for _, note := range notes {
		createdAt := note.CreatedAt.Format(time.RFC3339)
		result = append(result, &model.PetStaffNoteResp{
			ID:            note.ID,
			Content:       stringValue(note.Content),
			StaffName:     stringValue(note.StaffName),
			AliasName:     stringValue(note.StaffName),
			CreateAt:      createdAt,
			AliasCreateAt: createdAt,
		})
	}
	return result, nil
}

// SaveStaffNote 新增或修改店员留言
func (p *Pet) SaveStaffNote(req *model.PetSaveStaffNoteReq, tenantID string) (*model.PetStaffNoteResp, error) {
	if err := p.checkDevice(req.DeviceID, tenantID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	note := &model.PetStaffNote{
		ID:        uuid.New(),
		TenantID:  tenantID,
		DeviceID:  req.DeviceID,
		Content:   StringPtr(req.Content),
		StaffName: StringPtr(req.StaffName),
	}
	if req.ID != "" {
		exist, err := dal.GetPetStaffNote(req.ID, tenantID)
		if err != nil || exist == nil {
			return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
				"id": "pet staff note not found",
			})
		}
		note.ID = exist.ID
		note.CreatedAt = exist.CreatedAt
		if err := dal.UpdatePetStaffNote(note); err != nil {
			logrus.Error("修改店员留言失败:", err)
			return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
				"operation": "update_pet_staff_note",
				"error":     err.Error(),
			})
		}
	} else {
		if err := dal.CreatePetStaffNote(note); err != nil {
			logrus.Error("新增店员留言失败:", err)
			return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
				"operation": "create_pet_staff_note",
				"error":     err.Error(),
			})
		}
	}

	createdAt := note.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	return &model.PetStaffNoteResp{
		ID:            note.ID,
		Content:       stringValue(note.Content),
		StaffName:     stringValue(note.StaffName),
		AliasName:     stringValue(note.StaffName),
		CreateAt:      createdAt.Format(time.RFC3339),
		AliasCreateAt: createdAt.Format(time.RFC3339),
	}, nil
}

// DeleteStaffNote 删除店员留言
func (*Pet) DeleteStaffNote(id, tenantID string) error {
	if err := dal.DeletePetStaffNote(id, tenantID); err != nil {
		logrus.Error("删除店员留言失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "delete_pet_staff_note",
			"error":     err.Error(),
		})
	}
	return nil
}

// ListTimeline 动态时间线（倒序）
func (p *Pet) ListTimeline(tenantID, deviceID string, limit int) ([]*model.PetTimelineResp, error) {
	if err := p.checkDevice(deviceID, tenantID); err != nil {
		return nil, err
	}

	list, err := dal.ListPetTimeline(tenantID, deviceID, limit)
	if err != nil {
		logrus.Error("查询宠物动态失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_pet_timeline",
			"error":     err.Error(),
		})
	}

	result := make([]*model.PetTimelineResp, 0, len(list))
	for _, item := range list {
		createdAt := item.CreatedAt.Format(time.RFC3339)
		result = append(result, &model.PetTimelineResp{
			ID:            item.ID,
			Type:          stringValue(item.EventType),
			Title:         stringValue(item.Title),
			AliasTitle:    stringValue(item.Title),
			CreateAt:      createdAt,
			AliasCreateAt: createdAt,
		})
	}
	return result, nil
}

// CreateTimeline 新增动态（例如主人投喂零食）
func (p *Pet) CreateTimeline(req *model.PetSaveTimelineReq, tenantID string) (*model.PetTimelineResp, error) {
	if err := p.checkDevice(req.DeviceID, tenantID); err != nil {
		return nil, err
	}

	item := &model.PetTimeline{
		ID:        uuid.New(),
		TenantID:  tenantID,
		DeviceID:  req.DeviceID,
		EventType: StringPtr(req.Type),
		Title:     StringPtr(req.Title),
	}
	if err := dal.CreatePetTimeline(item); err != nil {
		logrus.Error("新增宠物动态失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "create_pet_timeline",
			"error":     err.Error(),
		})
	}

	createdAt := item.CreatedAt.Format(time.RFC3339)
	return &model.PetTimelineResp{
		ID:            item.ID,
		Type:          stringValue(item.EventType),
		Title:         stringValue(item.Title),
		AliasTitle:    stringValue(item.Title),
		CreateAt:      createdAt,
		AliasCreateAt: createdAt,
	}, nil
}

// DeleteTimeline 删除动态
func (*Pet) DeleteTimeline(id, tenantID string) error {
	if err := dal.DeletePetTimeline(id, tenantID); err != nil {
		logrus.Error("删除宠物动态失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "delete_pet_timeline",
			"error":     err.Error(),
		})
	}
	return nil
}

// ListAlbum 相册列表（倒序）
func (p *Pet) ListAlbum(tenantID, deviceID string, limit int) ([]*model.PetAlbumResp, error) {
	if err := p.checkDevice(deviceID, tenantID); err != nil {
		return nil, err
	}

	list, err := dal.ListPetAlbum(tenantID, deviceID, limit)
	if err != nil {
		logrus.Error("查询宠物相册失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_pet_album",
			"error":     err.Error(),
		})
	}

	result := make([]*model.PetAlbumResp, 0, len(list))
	for _, item := range list {
		createdAt := item.CreatedAt.Format(time.RFC3339)
		result = append(result, &model.PetAlbumResp{
			ID:            item.ID,
			URL:           stringValue(item.URL),
			AliasURL:      stringValue(item.URL),
			CreateAt:      createdAt,
			AliasCreateAt: createdAt,
		})
	}
	return result, nil
}

// CreateAlbum 新增相册记录（抓拍成功后由客户端上报）
func (p *Pet) CreateAlbum(req *model.PetSaveAlbumReq, tenantID string) (*model.PetAlbumResp, error) {
	if err := p.checkDevice(req.DeviceID, tenantID); err != nil {
		return nil, err
	}

	item := &model.PetAlbum{
		ID:       uuid.New(),
		TenantID: tenantID,
		DeviceID: req.DeviceID,
		URL:      StringPtr(req.URL),
	}
	if err := dal.CreatePetAlbum(item); err != nil {
		logrus.Error("新增相册记录失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "create_pet_album",
			"error":     err.Error(),
		})
	}

	createdAt := item.CreatedAt.Format(time.RFC3339)
	return &model.PetAlbumResp{
		ID:            item.ID,
		URL:           stringValue(item.URL),
		AliasURL:      stringValue(item.URL),
		CreateAt:      createdAt,
		AliasCreateAt: createdAt,
	}, nil
}

// DeleteAlbum 删除相册记录
func (*Pet) DeleteAlbum(id, tenantID string) error {
	if err := dal.DeletePetAlbum(id, tenantID); err != nil {
		logrus.Error("删除相册记录失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "delete_pet_album",
			"error":     err.Error(),
		})
	}
	return nil
}

// Int32Ptr 返回 int32 的指针
func Int32Ptr(v int32) *int32 {
	return &v
}
