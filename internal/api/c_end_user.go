package api

import (
	model "project/internal/model"
	service "project/internal/service"
	utils "project/pkg/utils"

	"github.com/gin-gonic/gin"
)

// CEndUserApi C 端用户接口（注册/登录/设备绑定）。注册与登录免鉴权，
// 其余接口经 middleware.CAuth() 校验独立令牌，claims 位于上下文键 "c_claims"。
type CEndUserApi struct{}

// Register C 端注册
// @Router   /api/v1/c/register [post]
func (*CEndUserApi) Register(c *gin.Context) {
	var req model.CEndRegisterReq
	if !BindAndValidate(c, &req) {
		return
	}
	data, err := service.GroupApp.CEndUser.Register(&req)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// CEndLogin C 端登录（方法名加 CEnd 前缀，避免与后台 UserApi.Login 在 Controller 嵌入后产生选择歧义）
// @Router   /api/v1/c/login [post]
func (*CEndUserApi) CEndLogin(c *gin.Context) {
	var req model.CEndLoginReq
	if !BindAndValidate(c, &req) {
		return
	}
	data, err := service.GroupApp.CEndUser.Login(&req)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// ListDevices C 端已绑定设备
// @Router   /api/v1/c/devices [get]
func (*CEndUserApi) ListDevices(c *gin.Context) {
	claims := c.MustGet("c_claims").(*utils.CUserClaims)
	data, err := service.GroupApp.CEndUser.ListDevices(claims.UserID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// BindDevice C 端凭绑定码绑定设备
// @Router   /api/v1/c/devices/bind [post]
func (*CEndUserApi) BindDevice(c *gin.Context) {
	var req model.CEndBindDeviceReq
	if !BindAndValidate(c, &req) {
		return
	}
	claims := c.MustGet("c_claims").(*utils.CUserClaims)
	data, err := service.GroupApp.CEndUser.BindDevice(claims.UserID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// UnbindDevice C 端解绑设备
// @Router   /api/v1/c/devices/{device_id} [delete]
func (*CEndUserApi) UnbindDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	claims := c.MustGet("c_claims").(*utils.CUserClaims)
	if err := service.GroupApp.CEndUser.UnbindDevice(claims.UserID, deviceID); err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}
