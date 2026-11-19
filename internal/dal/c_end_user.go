package dal

import (
	"time"

	model "project/internal/model"
	global "project/pkg/global"
)

// C 端用户与设备绑定关系数据访问层（public schema）。

func GetCEndUserByPhone(phone string) (*model.CEndUser, error) {
	user := &model.CEndUser{}
	err := global.DB.Model(&model.CEndUser{}).Where("phone = ?", phone).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetCEndUserByID(id string) (*model.CEndUser, error) {
	user := &model.CEndUser{}
	err := global.DB.Model(&model.CEndUser{}).Where("id = ?", id).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateCEndUser(user *model.CEndUser) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	return global.DB.Model(&model.CEndUser{}).Create(user).Error
}

func ListCEndUserDevices(cUserID string) ([]*model.CEndUserDevice, error) {
	var list []*model.CEndUserDevice
	err := global.DB.Model(&model.CEndUserDevice{}).
		Where("c_user_id = ? AND status = 1", cUserID).
		Order("bound_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetCEndUserDevice(cUserID, deviceID string) (*model.CEndUserDevice, error) {
	item := &model.CEndUserDevice{}
	err := global.DB.Model(&model.CEndUserDevice{}).
		Where("c_user_id = ? AND device_id = ?", cUserID, deviceID).
		First(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func CreateCEndUserDevice(item *model.CEndUserDevice) error {
	now := time.Now().UTC()
	item.CreatedAt = now
	item.BoundAt = now
	return global.DB.Model(&model.CEndUserDevice{}).Create(item).Error
}

func DeleteCEndUserDevice(cUserID, deviceID string) error {
	return global.DB.Model(&model.CEndUserDevice{}).
		Where("c_user_id = ? AND device_id = ?", cUserID, deviceID).
		Delete(&model.CEndUserDevice{}).Error
}
