package api

import "strings"

type OidcAuthRequest struct {
	DeviceInfo DeviceInfoInLogin `json:"deviceInfo" label:"设备信息"`
	Id         string            `json:"id" label:"id"`
	MyId       string            `json:"my_id" label:"my_id"`
	DeviceId   string            `json:"device_id" label:"device_id"`
	Op         string            `json:"op" label:"op"`
	Uuid       string            `json:"uuid" label:"uuid"`
	DeviceUuid string            `json:"device_uuid" label:"device_uuid"`
}

func (f *OidcAuthRequest) NormalizeDeviceId() string {
	if f == nil {
		return ""
	}
	for _, v := range []string{f.Id, f.MyId, f.DeviceId} {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func (f *OidcAuthRequest) NormalizeUuid() string {
	if f == nil {
		return ""
	}
	for _, v := range []string{f.Uuid, f.DeviceUuid} {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

type OidcAuthQuery struct {
	Code       string `json:"code" form:"code" label:"code"`
	Id         string `json:"id" form:"id" label:"id"`
	MyId       string `json:"my_id" form:"my_id" label:"my_id"`
	DeviceId   string `json:"device_id" form:"device_id" label:"device_id"`
	Uuid       string `json:"uuid" form:"uuid" label:"uuid"`
	DeviceUuid string `json:"device_uuid" form:"device_uuid" label:"device_uuid"`
}
