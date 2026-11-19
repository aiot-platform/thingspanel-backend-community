package apps

import (
	"project/internal/api"

	"github.com/gin-gonic/gin"
)

type BindCode struct{}

// InitBindCode 门店绑定码路由（B 端后台，挂在后台 JWT 鉴权之后）
func (*BindCode) InitBindCode(Router *gin.RouterGroup) {
	url := Router.Group("bind-codes")
	{
		url.GET("", api.Controllers.BindCodeApi.List)
		url.POST("", api.Controllers.BindCodeApi.Create)
		url.PUT("status", api.Controllers.BindCodeApi.SetStatus)
	}
}

type Tenant struct{}

// InitTenant 租户层级路由（B 端后台，挂在后台 JWT 鉴权之后）
func (*Tenant) InitTenant(Router *gin.RouterGroup) {
	url := Router.Group("tenants")
	{
		url.GET("tree", api.Controllers.TenantApi.Tree)
		url.POST("", api.Controllers.TenantApi.Save)
	}
}
