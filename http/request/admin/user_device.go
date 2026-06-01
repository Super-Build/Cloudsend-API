package admin

type UserDeviceListQuery struct {
	PageQuery
	UserId   uint   `form:"user_id"`
	Username string `form:"username"`
}

type UserDeviceSetLimitForm struct {
	UserId     uint `json:"user_id" validate:"required,gt=0"`
	MaxDevices int  `json:"max_devices" validate:"required,gte=-1,lte=10000,ne=0"`
}

type UserDeviceUnbindForm struct {
	Id uint `json:"id" validate:"required,gt=0"`
}

type UserDeviceBatchUnbindForm struct {
	Ids []uint `json:"ids" validate:"required"`
}

type UserDeviceDeleteForm struct {
	Id uint `json:"id" validate:"required,gt=0"`
}

type UserDeviceBatchDeleteForm struct {
	Ids []uint `json:"ids" validate:"required"`
}

type UserDeviceClearUnboundForm struct {
	UserId uint `json:"user_id"`
}
