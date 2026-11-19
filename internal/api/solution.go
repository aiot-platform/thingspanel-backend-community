package api

import (
	model "project/internal/model"
	service "project/internal/service"
	utils "project/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SolutionApi 行业方案接口（平台超管维护方案字典与分配；租户查看本租户方案运营数据）
type SolutionApi struct{}

// ListIndustrySolutions 方案字典列表
// @Router   /api/v1/solutions [get]
func (*SolutionApi) ListIndustrySolutions(c *gin.Context) {
	data, err := service.GroupApp.Solution.ListIndustrySolutions()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// SaveIndustrySolution 新增/修改方案字典
// @Router   /api/v1/solutions [post]
func (*SolutionApi) SaveIndustrySolution(c *gin.Context) {
	var req model.IndustrySolutionSaveReq
	if !BindAndValidate(c, &req) {
		return
	}
	data, err := service.GroupApp.Solution.SaveIndustrySolution(&req)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// DeleteIndustrySolution 删除方案字典
// @Router   /api/v1/solutions/{id} [delete]
func (*SolutionApi) DeleteIndustrySolution(c *gin.Context) {
	id := c.Param("id")
	if err := service.GroupApp.Solution.DeleteIndustrySolution(id); err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}

// ListTenants 租户列表（用于分配方案）
// @Router   /api/v1/solutions/tenants [get]
func (*SolutionApi) ListTenants(c *gin.Context) {
	data, err := service.GroupApp.Solution.ListTenants()
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// ListAssignments 查询某租户已分配方案
// @Router   /api/v1/solutions/assignments [get]
func (*SolutionApi) ListAssignments(c *gin.Context) {
	var req model.TenantSolutionQueryReq
	if !BindAndValidate(c, &req) {
		return
	}
	data, err := service.GroupApp.Solution.ListAssignments(req.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}

// Assign 给租户分配方案
// @Router   /api/v1/solutions/assignments [post]
func (*SolutionApi) Assign(c *gin.Context) {
	var req model.TenantSolutionAssignReq
	if !BindAndValidate(c, &req) {
		return
	}
	if err := service.GroupApp.Solution.Assign(&req); err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}

// Unassign 取消租户的方案分配
// @Router   /api/v1/solutions/assignments/{tenant_id}/{code} [delete]
func (*SolutionApi) Unassign(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	code := c.Param("code")
	if err := service.GroupApp.Solution.Unassign(tenantID, code); err != nil {
		c.Error(err)
		return
	}
	c.Set("data", nil)
}

// Mine 租户查看本租户已启用方案与方案专属运营数据概览
// @Router   /api/v1/solutions/mine [get]
func (*SolutionApi) Mine(c *gin.Context) {
	userClaims := c.MustGet("claims").(*utils.UserClaims)
	data, err := service.GroupApp.Solution.Mine(userClaims.TenantID)
	if err != nil {
		c.Error(err)
		return
	}
	c.Set("data", data)
}
