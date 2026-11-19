package api

import (
	model "project/internal/model"
	service "project/internal/service"
	utils "project/pkg/utils"

	"github.com/gin-gonic/gin"
)

// PetApi 宠物看护接口（供 app 宠物面板与平台管理页使用）
type PetApi struct{}

// GenerateRtcToken 签发实时音视频入会凭证
// @Router   /api/v1/pet/rtc/token [post]
func (*PetApi) GenerateRtcToken(c *gin.Context) {
	var req model.PetRtcTokenReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.GenerateRtcToken(&req, userClaims)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// GetProfile 查询宠物档案与店铺信息
// @Router   /api/v1/pet/profile [get]
func (*PetApi) GetProfile(c *gin.Context) {
	var req model.PetDeviceReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.GetProfile(userClaims.TenantID, req.DeviceID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// SaveProfile 保存宠物档案与店铺信息
// @Router   /api/v1/pet/profile [put]
func (*PetApi) SaveProfile(c *gin.Context) {
	var req model.PetSaveProfileReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	err := service.GroupApp.Pet.SaveProfile(&req, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}

// GetDailyReport 查询宠物日报
// @Router   /api/v1/pet/daily-report [get]
func (*PetApi) GetDailyReport(c *gin.Context) {
	var req model.PetDailyReportReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.GetDailyReport(userClaims.TenantID, req.DeviceID, req.Date)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// ListStaffNotes 店员留言列表
// @Router   /api/v1/pet/staff-notes [get]
func (*PetApi) ListStaffNotes(c *gin.Context) {
	var req model.PetDeviceReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.ListStaffNotes(userClaims.TenantID, req.DeviceID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// SaveStaffNote 新增/修改店员留言
// @Router   /api/v1/pet/staff-note [post]
func (*PetApi) SaveStaffNote(c *gin.Context) {
	var req model.PetSaveStaffNoteReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.SaveStaffNote(&req, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// DeleteStaffNote 删除店员留言
// @Router   /api/v1/pet/staff-note/{id} [delete]
func (*PetApi) DeleteStaffNote(c *gin.Context) {
	id := c.Param("id")
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	err := service.GroupApp.Pet.DeleteStaffNote(id, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}

// ListTimeline 动态时间线
// @Router   /api/v1/pet/timeline [get]
func (*PetApi) ListTimeline(c *gin.Context) {
	var req model.PetTimelineListReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.ListTimeline(userClaims.TenantID, req.DeviceID, req.Limit)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// CreateTimeline 新增动态
// @Router   /api/v1/pet/timeline [post]
func (*PetApi) CreateTimeline(c *gin.Context) {
	var req model.PetSaveTimelineReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.CreateTimeline(&req, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// DeleteTimeline 删除动态
// @Router   /api/v1/pet/timeline/{id} [delete]
func (*PetApi) DeleteTimeline(c *gin.Context) {
	id := c.Param("id")
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	err := service.GroupApp.Pet.DeleteTimeline(id, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}

// ListAlbum 相册列表
// @Router   /api/v1/pet/album [get]
func (*PetApi) ListAlbum(c *gin.Context) {
	var req model.PetAlbumListReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.ListAlbum(userClaims.TenantID, req.DeviceID, req.Limit)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// CreateAlbum 新增相册记录
// @Router   /api/v1/pet/album [post]
func (*PetApi) CreateAlbum(c *gin.Context) {
	var req model.PetSaveAlbumReq
	if !BindAndValidate(c, &req) {
		return
	}
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	data, err := service.GroupApp.Pet.CreateAlbum(&req, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// DeleteAlbum 删除相册记录
// @Router   /api/v1/pet/album/{id} [delete]
func (*PetApi) DeleteAlbum(c *gin.Context) {
	id := c.Param("id")
	userClaims := c.MustGet("claims").(*utils.UserClaims)

	err := service.GroupApp.Pet.DeleteAlbum(id, userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}
