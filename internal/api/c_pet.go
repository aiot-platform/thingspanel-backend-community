package api

import (
	service "project/internal/service"
	utils "project/pkg/utils"

	"github.com/gin-gonic/gin"
)

// CPetApi C 端宠物看护接口。经 middleware.CAuth() 校验独立令牌，
// C 端用户 id 取自上下文键 "c_claims"，租户由绑定关系推导。
type CPetApi struct{}

func cUserID(c *gin.Context) string {
	claims := c.MustGet("c_claims").(*utils.CUserClaims)
	return claims.UserID
}

// ListPets C 端宠物列表
// @Router   /api/v1/c/pets [get]
func (*CPetApi) ListPets(c *gin.Context) {
	data, err := service.GroupApp.CPet.ListPets(cUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// GetProfile C 端宠物档案
// @Router   /api/v1/c/pets/{device_id}/profile [get]
func (*CPetApi) GetProfile(c *gin.Context) {
	data, err := service.GroupApp.CPet.GetProfile(cUserID(c), c.Param("device_id"))
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// GetDailyReport C 端宠物日报
// @Router   /api/v1/c/pets/{device_id}/daily-report [get]
func (*CPetApi) GetDailyReport(c *gin.Context) {
	data, err := service.GroupApp.CPet.GetDailyReport(cUserID(c), c.Param("device_id"), c.Query("date"))
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// ListStaffNotes C 端店员留言
// @Router   /api/v1/c/pets/{device_id}/staff-notes [get]
func (*CPetApi) ListStaffNotes(c *gin.Context) {
	data, err := service.GroupApp.CPet.ListStaffNotes(cUserID(c), c.Param("device_id"))
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// ListTimeline C 端宠物动态
// @Router   /api/v1/c/pets/{device_id}/timeline [get]
func (*CPetApi) ListTimeline(c *gin.Context) {
	limit := parseLimit(c.Query("limit"))
	data, err := service.GroupApp.CPet.ListTimeline(cUserID(c), c.Param("device_id"), limit)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// ListAlbum C 端宠物相册
// @Router   /api/v1/c/pets/{device_id}/album [get]
func (*CPetApi) ListAlbum(c *gin.Context) {
	limit := parseLimit(c.Query("limit"))
	data, err := service.GroupApp.CPet.ListAlbum(cUserID(c), c.Param("device_id"), limit)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// GenerateRtcToken C 端实时音视频入会凭证
// @Router   /api/v1/c/pets/{device_id}/rtc/token [post]
func (*CPetApi) GenerateRtcToken(c *gin.Context) {
	data, err := service.GroupApp.CPet.GenerateRtcToken(cUserID(c), c.Param("device_id"))
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// parseLimit 解析查询参数 limit，非法值返回 0（由服务层取默认值）
func parseLimit(raw string) int {
	if raw == "" {
		return 0
	}
	limit := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0
		}
		limit = limit*10 + int(ch-'0')
		if limit > 200 {
			return 200
		}
	}
	return limit
}
