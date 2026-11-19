package apps

import (
	"project/internal/api"
	middleware "project/internal/middleware"

	"github.com/gin-gonic/gin"
)

type CEndUser struct{}

// InitCEndUser C 端用户路由。该分组必须在 v1 挂载后台 JWT 鉴权之前创建，
// 以便使用独立的 C 端令牌鉴权（middleware.CAuth）。
func (*CEndUser) InitCEndUser(Router *gin.RouterGroup) {
	url := Router.Group("c")
	{
		// 免鉴权：注册 / 登录
		url.POST("register", api.Controllers.CEndUserApi.Register)
		url.POST("login", api.Controllers.CEndUserApi.CEndLogin)

		// 需 C 端独立鉴权
		auth := url.Group("")
		auth.Use(middleware.CAuth())
		{
			auth.GET("devices", api.Controllers.CEndUserApi.ListDevices)
			auth.POST("devices/bind", api.Controllers.CEndUserApi.BindDevice)
			auth.DELETE("devices/:device_id", api.Controllers.CEndUserApi.UnbindDevice)
		}
	}
}
