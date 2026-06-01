package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

type UserDevice struct{}

func (ct *UserDevice) List(c *gin.Context) {
	query := &admin.UserDeviceListQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}

	items, total := service.AllService.UserService.UserDeviceList(query.Page, query.PageSize, query.UserId, query.Username)
	userIDs := make([]uint, 0)
	seen := make(map[uint]struct{})
	for _, item := range items {
		if _, ok := seen[item.UserId]; ok {
			continue
		}
		seen[item.UserId] = struct{}{}
		userIDs = append(userIDs, item.UserId)
	}
	users := service.AllService.UserService.ListByIds(userIDs)
	userMap := make(map[uint]*model.User)
	for _, u := range users {
		userMap[u.Id] = u
	}

	res := make([]gin.H, 0, len(items))
	for _, item := range items {
		u := userMap[item.UserId]
		maxDevices := 1
		username := ""
		hostname := item.Hostname
		if hostname == "" {
			hostname = item.DeviceId
		}
		if u != nil {
			maxDevices = u.MaxDevices
			username = u.Username
		}
		res = append(res, gin.H{
			"id":             item.Id,
			"user_id":        item.UserId,
			"username":       username,
			"max_devices":    maxDevices,
			"uuid":           item.Uuid,
			"device_id":      item.DeviceId,
			"hostname":       hostname,
			"platform":       item.Platform,
			"ip":             item.Ip,
			"status":         item.Status,
			"first_login_at": item.FirstLoginAt,
			"last_login_at":  item.LastLoginAt,
			"created_at":     item.CreatedAt,
			"updated_at":     item.UpdatedAt,
		})
	}

	response.Success(c, gin.H{
		"list":      res,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

func (ct *UserDevice) SetLimit(c *gin.Context) {
	f := &admin.UserDeviceSetLimitForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}

	u := service.AllService.UserService.InfoById(f.UserId)
	if u.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}

	normalized := service.AllService.UserService.NormalizeMaxDevicesForAdminApi(u, f.MaxDevices)
	u.MaxDevices = normalized
	if err := global.DB.Model(u).Update("max_devices", normalized).Error; err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}

	response.Success(c, gin.H{
		"user_id":     u.Id,
		"max_devices": u.MaxDevices,
	})
}

func (ct *UserDevice) Unbind(c *gin.Context) {
	f := &admin.UserDeviceUnbindForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}

	item := service.AllService.UserService.UserDeviceInfoById(f.Id)
	if item.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}
	if err := service.AllService.UserService.UnbindUserDevice(item); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

func (ct *UserDevice) BatchUnbind(c *gin.Context) {
	f := &admin.UserDeviceBatchUnbindForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if len(f.Ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	if err := service.AllService.UserService.BatchUnbindUserDevices(f.Ids); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

func (ct *UserDevice) Delete(c *gin.Context) {
	f := &admin.UserDeviceDeleteForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	item := service.AllService.UserService.UserDeviceInfoById(f.Id)
	if item.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}
	if err := service.AllService.UserService.DeleteUserDevice(item); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

func (ct *UserDevice) BatchDelete(c *gin.Context) {
	f := &admin.UserDeviceBatchDeleteForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if len(f.Ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	if err := service.AllService.UserService.BatchDeleteUserDevices(f.Ids); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

func (ct *UserDevice) ClearUnbound(c *gin.Context) {
	f := &admin.UserDeviceClearUnboundForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if err := service.AllService.UserService.ClearUnboundUserDevices(f.UserId); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}
