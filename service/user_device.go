package service

import (
	"errors"
	"strings"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/model/custom_types"
	"gorm.io/gorm"
)

func (us *UserService) normalizeMaxDevices(u *model.User, maxDevices int) int {
	if maxDevices == 0 {
		if u != nil && us.IsAdmin(u) {
			return -1
		}
		return 1
	}
	if maxDevices < -1 {
		return 1
	}
	if maxDevices > 10000 {
		return 10000
	}
	return maxDevices
}

func (us *UserService) NormalizeMaxDevicesForAdminApi(u *model.User, maxDevices int) int {
	return us.normalizeMaxDevices(u, maxDevices)
}

func (us *UserService) normalizeLoginUuid(log *model.LoginLog) string {
	if log == nil {
		return ""
	}
	if log.Uuid != "" {
		return strings.TrimSpace(log.Uuid)
	}
	if log.DeviceId != "" {
		return "device:" + strings.TrimSpace(log.DeviceId)
	}
	return ""
}

func (us *UserService) shouldBindUserDevice(log *model.LoginLog) bool {
	if log == nil {
		return false
	}
	client := strings.TrimSpace(strings.ToLower(log.Client))
	// Web admin login should not consume device quota. Other client types
	// (including empty/legacy app payloads) are treated as real devices.
	return client != model.LoginLogClientWebAdmin
}

func (us *UserService) ensureUserDeviceBinding(u *model.User, log *model.LoginLog) error {
	if !us.shouldBindUserDevice(log) {
		return nil
	}
	uuid := us.normalizeLoginUuid(log)
	if uuid == "" {
		return errors.New("DeviceIdentifierMissing")
	}

	now := time.Now()
	device := &model.UserDevice{}
	err := DB.Where("user_id = ? and uuid = ?", u.Id, uuid).First(device).Error
	if err == nil {
		return DB.Model(device).Updates(map[string]interface{}{
			"status":        model.UserDeviceStatusBound,
			"device_id":     log.DeviceId,
			"hostname":      log.DeviceName,
			"platform":      log.Platform,
			"ip":            log.Ip,
			"last_login_at": custom_types.AutoTime(now),
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	maxDevices := us.normalizeMaxDevices(u, u.MaxDevices)
	if maxDevices != -1 {
		var cnt int64
		if err = DB.Model(&model.UserDevice{}).
			Where("user_id = ? and status = ?", u.Id, model.UserDeviceStatusBound).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt >= int64(maxDevices) {
			return errors.New("DeviceLimitExceeded")
		}
	}

	return DB.Create(&model.UserDevice{
		UserId:       u.Id,
		Uuid:         uuid,
		DeviceId:     log.DeviceId,
		Hostname:     log.DeviceName,
		Platform:     log.Platform,
		Ip:           log.Ip,
		Status:       model.UserDeviceStatusBound,
		FirstLoginAt: custom_types.AutoTime(now),
		LastLoginAt:  custom_types.AutoTime(now),
	}).Error
}

func (us *UserService) UserDeviceList(page, size uint, userId uint, username string) (list []*model.UserDevice, total int64) {
	tx := DB.Model(&model.UserDevice{})
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if username != "" {
		uids := make([]uint, 0)
		DB.Model(&model.User{}).Where("username like ?", "%"+username+"%").Pluck("id", &uids)
		if len(uids) == 0 {
			return []*model.UserDevice{}, 0
		}
		tx = tx.Where("user_id in ?", uids)
	}
	tx.Count(&total)
	tx.Order("id desc").Scopes(Paginate(page, size)).Find(&list)
	return
}

func (us *UserService) UserDeviceInfoById(id uint) *model.UserDevice {
	item := &model.UserDevice{}
	DB.Where("id = ?", id).First(item)
	return item
}

func (us *UserService) UnbindUserDevice(item *model.UserDevice) error {
	if item == nil || item.Id == 0 {
		return errors.New("ItemNotFound")
	}
	err := DB.Model(item).Update("status", model.UserDeviceStatusUnbound).Error
	if err != nil {
		return err
	}
	return DB.Where("user_id = ? and (device_uuid = ? or device_id = ?)", item.UserId, item.Uuid, strings.TrimPrefix(item.Uuid, "device:")).Delete(&model.UserToken{}).Error
}

func (us *UserService) BatchUnbindUserDevices(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	items := make([]model.UserDevice, 0)
	if err := DB.Where("id in ?", ids).Find(&items).Error; err != nil {
		return err
	}
	if err := DB.Model(&model.UserDevice{}).Where("id in ?", ids).Update("status", model.UserDeviceStatusUnbound).Error; err != nil {
		return err
	}
	for _, it := range items {
		_ = DB.Where("user_id = ? and (device_uuid = ? or device_id = ?)", it.UserId, it.Uuid, strings.TrimPrefix(it.Uuid, "device:")).Delete(&model.UserToken{}).Error
	}
	return nil
}

func (us *UserService) DeleteUserDevice(item *model.UserDevice) error {
	if item == nil || item.Id == 0 {
		return errors.New("ItemNotFound")
	}
	if item.Status == model.UserDeviceStatusBound {
		if err := DB.Where("user_id = ? and (device_uuid = ? or device_id = ?)", item.UserId, item.Uuid, strings.TrimPrefix(item.Uuid, "device:")).Delete(&model.UserToken{}).Error; err != nil {
			return err
		}
	}
	return DB.Delete(&model.UserDevice{}, item.Id).Error
}

func (us *UserService) BatchDeleteUserDevices(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	items := make([]model.UserDevice, 0)
	if err := DB.Where("id in ?", ids).Find(&items).Error; err != nil {
		return err
	}
	for _, it := range items {
		if it.Status == model.UserDeviceStatusBound {
			if err := DB.Where("user_id = ? and (device_uuid = ? or device_id = ?)", it.UserId, it.Uuid, strings.TrimPrefix(it.Uuid, "device:")).Delete(&model.UserToken{}).Error; err != nil {
				return err
			}
		}
	}
	return DB.Where("id in ?", ids).Delete(&model.UserDevice{}).Error
}

func (us *UserService) ClearUnboundUserDevices(userId uint) error {
	tx := DB.Model(&model.UserDevice{}).Where("status = ?", model.UserDeviceStatusUnbound)
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	return tx.Delete(&model.UserDevice{}).Error
}
