package apps

import (
	"project/internal/api"

	"github.com/gin-gonic/gin"
)

type Pet struct{}

// InitPet 宠物看护路由
func (*Pet) InitPet(Router *gin.RouterGroup) {
	url := Router.Group("pet")
	{
		// 实时音视频入会凭证
		url.POST("rtc/token", api.Controllers.PetApi.GenerateRtcToken)

		// 宠物档案与店铺信息
		url.GET("profile", api.Controllers.PetApi.GetProfile)
		url.PUT("profile", api.Controllers.PetApi.SaveProfile)

		// 日报
		url.GET("daily-report", api.Controllers.PetApi.GetDailyReport)

		// 店员留言
		url.GET("staff-notes", api.Controllers.PetApi.ListStaffNotes)
		url.POST("staff-note", api.Controllers.PetApi.SaveStaffNote)
		url.DELETE("staff-note/:id", api.Controllers.PetApi.DeleteStaffNote)

		// 动态时间线
		url.GET("timeline", api.Controllers.PetApi.ListTimeline)
		url.POST("timeline", api.Controllers.PetApi.CreateTimeline)
		url.DELETE("timeline/:id", api.Controllers.PetApi.DeleteTimeline)

		// 相册
		url.GET("album", api.Controllers.PetApi.ListAlbum)
		url.POST("album", api.Controllers.PetApi.CreateAlbum)
		url.DELETE("album/:id", api.Controllers.PetApi.DeleteAlbum)
	}
}
