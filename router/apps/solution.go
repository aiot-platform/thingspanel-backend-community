package apps

import (
	"project/internal/api"

	"github.com/gin-gonic/gin"
)

type Solution struct{}

// InitSolution 行业方案路由（平台超管维护方案字典与分配；租户查看本租户方案运营数据）
func (*Solution) InitSolution(Router *gin.RouterGroup) {
	url := Router.Group("solutions")
	{
		url.GET("", api.Controllers.SolutionApi.ListIndustrySolutions)
		url.POST("", api.Controllers.SolutionApi.SaveIndustrySolution)
		url.DELETE(":id", api.Controllers.SolutionApi.DeleteIndustrySolution)

		url.GET("tenants", api.Controllers.SolutionApi.ListTenants)
		url.GET("assignments", api.Controllers.SolutionApi.ListAssignments)
		url.POST("assignments", api.Controllers.SolutionApi.Assign)
		url.DELETE("assignments/:tenant_id/:code", api.Controllers.SolutionApi.Unassign)

		url.GET("mine", api.Controllers.SolutionApi.Mine)
	}
}
