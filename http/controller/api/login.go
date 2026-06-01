package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	apiResp "github.com/lejianwen/rustdesk-api/v2/http/response/api"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

type Login struct {}

func (l *Login) Login(c *gin.Context) {
	if global.Config.App.DisablePwdLogin {
		response.Error(c, response.TranslateMsg(c, "PwdLoginDisabled"))
		return
	}

	loginLimiter := global.LoginLimiter
	clientIp := c.ClientIP()

	f := &api.LoginForm{}
	err := c.ShouldBindJSON(f)
	if err != nil {
		loginLimiter.RecordFailedAttempt(clientIp)
		global.Logger.Warn(fmt.Sprintf("Login Fail: %s %s %s", "ParamsError", c.RemoteIP(), c.ClientIP()))
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}

	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		loginLimiter.RecordFailedAttempt(clientIp)
		global.Logger.Warn(fmt.Sprintf("Login Fail: %s %s %s", "ParamsError", c.RemoteIP(), c.ClientIP()))
		response.Error(c, errList[0])
		return
	}

	u := service.AllService.UserService.InfoByUsernamePassword(f.Username, f.Password)
	if u.Id == 0 {
		loginLimiter.RecordFailedAttempt(clientIp)
		global.Logger.Warn(fmt.Sprintf("Login Fail: %s %s %s", "UsernameOrPasswordError", c.RemoteIP(), c.ClientIP()))
		response.Error(c, response.TranslateMsg(c, "UsernameOrPasswordError"))
		return
	}

	if !service.AllService.UserService.CheckUserEnable(u) {
		response.Error(c, response.TranslateMsg(c, "UserDisabled"))
		return
	}

	ref := c.GetHeader("referer")
	clientType := strings.TrimSpace(f.DeviceInfo.Type)
	if ref != "" {
		clientType = model.LoginLogClientWeb
	} else if clientType == "" {
		clientType = model.LoginLogClientApp
	}

	ut, loginErr := service.AllService.UserService.Login(u, &model.LoginLog{
		UserId:   u.Id,
		Client:   clientType,
		DeviceId: f.NormalizeDeviceId(),
		DeviceName: strings.TrimSpace(f.DeviceInfo.Name),
		Uuid:     f.NormalizeUuid(),
		Ip:       c.ClientIP(),
		Type:     model.LoginLogTypeAccount,
		Platform: f.DeviceInfo.Os,
	})
	if loginErr != nil {
		msg := response.TranslateMsg(c, loginErr.Error())
		if loginErr.Error() == "DeviceLimitExceeded" || loginErr.Error() == "DeviceIdentifierMissing" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   loginErr.Error(),
				"message": msg,
			})
			return
		}
		response.Error(c, msg)
		return
	}

	c.JSON(http.StatusOK, apiResp.LoginRes{
		AccessToken: ut.Token,
		Type:        "access_token",
		User:        *(&apiResp.UserPayload{}).FromUser(u),
	})
}

func (l *Login) LoginOptions(c *gin.Context) {
	ops := service.AllService.OauthService.GetOauthProviders()
	if global.Config.App.WebSso {
		ops = append(ops, model.OauthTypeWebauth)
	}
	var oidcItems []map[string]string
	for _, v := range ops {
		oidcItems = append(oidcItems, map[string]string{"name": v})
	}
	common, err := json.Marshal(oidcItems)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "SystemError")+err.Error())
		return
	}
	var res []string
	res = append(res, "common-oidc/"+string(common))
	for _, v := range ops {
		res = append(res, "oidc/"+v)
	}
	c.JSON(http.StatusOK, res)
}

func (l *Login) Logout(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	token, ok := c.Get("token")
	if ok {
		service.AllService.UserService.Logout(u, token.(string))
	}
	c.JSON(http.StatusOK, nil)
}
