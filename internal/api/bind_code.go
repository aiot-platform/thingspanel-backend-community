package api

import (
	model "project/internal/model"
	service "project/internal/service"
	utils "project/pkg/utils"

	"github.com/gin-gonic/gin"
)

// BindCodeApi 门店绑定码接口（B 端后台，经后台 JWT + CasbinRBAC 鉴权）
type BindCodeApi struct{}

// Create 生成绑定码
// @Router   /api/v1/bind-codes [post]
func (*BindCodeApi) Create(c *gin.Context) {
	var req model.DeviceBindCodeCreateReq
	if !BindAndValidate(c, &req) {
		return
	}
	claims := c.MustGet("claims").(*utils.UserClaims)
	data, err := service.GroupApp.BindCode.Create(&req, claims.TenantID, claims.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// List 查询本租户绑定码
// @Router   /api/v1/bind-codes [get]
func (*BindCodeApi) List(c *gin.Context) {
	var req model.DeviceBindCodeQueryReq
	if !BindAndValidate(c, &req) {
		return
	}
	claims := c.MustGet("claims").(*utils.UserClaims)
	data, err := service.GroupApp.BindCode.List(claims.TenantID, req.DeviceID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// SetStatus 启用/失效绑定码
// @Router   /api/v1/bind-codes/status [put]
func (*BindCodeApi) SetStatus(c *gin.Context) {
	var req model.DeviceBindCodeStatusReq
	if !BindAndValidate(c, &req) {
		return
	}
	claims := c.MustGet("claims").(*utils.UserClaims)
	if err := service.GroupApp.BindCode.SetStatus(&req, claims.TenantID); err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}
