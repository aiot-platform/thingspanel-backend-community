package dal

import (
	"fmt"
	"time"

	model "project/internal/model"
)

// 宠物看护（宠物医疗方案）数据访问层：所有读写都带 tenant_id，保证租户隔离。
// 数据表位于 sol_pet_medical schema，通过 PetDB() 获取该 schema 的数据库连接。

func GetPetProfile(tenantID, deviceID string) (*model.PetProfile, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	profile := &model.PetProfile{}
	err = db.Model(&model.PetProfile{}).
		Where("tenant_id = ? AND device_id = ?", tenantID, deviceID).
		First(profile).Error
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// UpsertPetProfile 按 device_id 唯一键写入或更新宠物档案
func UpsertPetProfile(profile *model.PetProfile) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	exist, err := GetPetProfile(profile.TenantID, profile.DeviceID)
	if err != nil {
		profile.CreatedAt = now
		profile.UpdatedAt = now
		return db.Model(&model.PetProfile{}).Create(profile).Error
	}
	profile.ID = exist.ID
	profile.CreatedAt = exist.CreatedAt
	profile.UpdatedAt = now
	return db.Model(&model.PetProfile{}).
		Where("id = ?", exist.ID).
		Updates(map[string]interface{}{
			"pet_name":      profile.PetName,
			"breed":         profile.Breed,
			"age":           profile.Age,
			"avatar":        profile.Avatar,
			"boarding_days": profile.BoardingDays,
			"updated_at":    now,
		}).Error
}

func GetPetShop(tenantID, deviceID string) (*model.PetShop, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	shop := &model.PetShop{}
	err = db.Model(&model.PetShop{}).
		Where("tenant_id = ? AND device_id = ?", tenantID, deviceID).
		First(shop).Error
	if err != nil {
		return nil, err
	}
	return shop, nil
}

// UpsertPetShop 按 device_id 唯一键写入或更新店铺信息
func UpsertPetShop(shop *model.PetShop) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	exist, err := GetPetShop(shop.TenantID, shop.DeviceID)
	if err != nil {
		shop.CreatedAt = now
		shop.UpdatedAt = now
		return db.Model(&model.PetShop{}).Create(shop).Error
	}
	shop.ID = exist.ID
	shop.CreatedAt = exist.CreatedAt
	shop.UpdatedAt = now
	return db.Model(&model.PetShop{}).
		Where("id = ?", exist.ID).
		Updates(map[string]interface{}{
			"name":       shop.Name,
			"open_time":  shop.OpenTime,
			"close_time": shop.CloseTime,
			"updated_at": now,
		}).Error
}

func ListPetStaffNotes(tenantID, deviceID string) ([]*model.PetStaffNote, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	var notes []*model.PetStaffNote
	err = db.Model(&model.PetStaffNote{}).
		Where("tenant_id = ? AND device_id = ?", tenantID, deviceID).
		Order("created_at DESC").
		Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func GetPetStaffNote(id, tenantID string) (*model.PetStaffNote, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	note := &model.PetStaffNote{}
	err = db.Model(&model.PetStaffNote{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(note).Error
	if err != nil {
		return nil, err
	}
	return note, nil
}

func CreatePetStaffNote(note *model.PetStaffNote) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	note.CreatedAt = now
	note.UpdatedAt = now
	return db.Model(&model.PetStaffNote{}).Create(note).Error
}

func UpdatePetStaffNote(note *model.PetStaffNote) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	return db.Model(&model.PetStaffNote{}).
		Where("id = ? AND tenant_id = ?", note.ID, note.TenantID).
		Updates(map[string]interface{}{
			"content":    note.Content,
			"staff_name": note.StaffName,
			"updated_at": time.Now().UTC(),
		}).Error
}

func DeletePetStaffNote(id, tenantID string) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	res := db.Model(&model.PetStaffNote{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&model.PetStaffNote{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("pet staff note not found: %s", id)
	}
	return nil
}

func ListPetTimeline(tenantID, deviceID string, limit int) ([]*model.PetTimeline, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var list []*model.PetTimeline
	err = db.Model(&model.PetTimeline{}).
		Where("tenant_id = ? AND device_id = ?", tenantID, deviceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func CreatePetTimeline(item *model.PetTimeline) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	return db.Model(&model.PetTimeline{}).Create(item).Error
}

func DeletePetTimeline(id, tenantID string) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	res := db.Model(&model.PetTimeline{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&model.PetTimeline{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("pet timeline not found: %s", id)
	}
	return nil
}

func ListPetAlbum(tenantID, deviceID string, limit int) ([]*model.PetAlbum, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 60
	}
	var list []*model.PetAlbum
	err = db.Model(&model.PetAlbum{}).
		Where("tenant_id = ? AND device_id = ?", tenantID, deviceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func CreatePetAlbum(item *model.PetAlbum) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	return db.Model(&model.PetAlbum{}).Create(item).Error
}

func DeletePetAlbum(id, tenantID string) error {
	db, err := PetDB()
	if err != nil {
		return err
	}
	res := db.Model(&model.PetAlbum{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&model.PetAlbum{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("pet album not found: %s", id)
	}
	return nil
}

// GetPetDailyReport 按设备与日期查询日报，日期为空时取当天
func GetPetDailyReport(tenantID, deviceID, date string) (*model.PetDailyReport, error) {
	db, err := PetDB()
	if err != nil {
		return nil, err
	}
	report := &model.PetDailyReport{}
	err = db.Model(&model.PetDailyReport{}).
		Where("tenant_id = ? AND device_id = ? AND report_date = ?", tenantID, deviceID, date).
		First(report).Error
	if err != nil {
		return nil, err
	}
	return report, nil
}

// CountPetProfiles 统计方案 schema 内本租户的宠物档案数量（方案运营数据概览用）。
func CountPetProfiles(tenantID string) (int64, error) {
	db, err := PetDB()
	if err != nil {
		return 0, err
	}
	var count int64
	err = db.Model(&model.PetProfile{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
