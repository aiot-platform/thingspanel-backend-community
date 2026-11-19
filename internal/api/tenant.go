package api

import (
	model "project/internal/model"
	service "project/internal/service"

	"github.com/gin-gonic/gin"
)

// TenantApi 租户层级接口（B 端后台，经后台 JWT + CasbinRBAC 鉴权）
type TenantApi struct{}

// Tree 租户层级树
// @Router   /api/v1/tenants/tree [get]
func (*TenantApi) Tree(c *gin.Context) {
	data, err := service.GroupApp.Tenant.Tree()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// Save 新增/修改租户层级节点
// @Router   /api/v1/tenants [post]
func (*TenantApi) Save(c *gin.Context) {
	var req model.TenantSaveReq
	if !BindAndValidate(c, &req) {
		return
	}
	data, err := service.GroupApp.Tenant.Save(&req)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}
