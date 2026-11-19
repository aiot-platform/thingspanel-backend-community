package service

import (
	"regexp"

	dal "project/internal/dal"
	model "project/internal/model"
	"project/pkg/errcode"

	"github.com/go-basic/uuid"
	"github.com/sirupsen/logrus"
)

// Solution 行业方案业务服务（方案字典 / 租户分配 / 方案运营概览）。
type Solution struct{}

var schemaNameRe = regexp.MustCompile(`^sol_[a-z0-9_]+$`)

// ListIndustrySolutions 方案字典列表
func (*Solution) ListIndustrySolutions() ([]*model.IndustrySolution, error) {
	return dal.ListIndustrySolutions()
}

// ListTenants 租户列表（用于超管分配方案）
func (*Solution) ListTenants() ([]*model.TenantInfo, error) {
	return dal.ListTenantAdmins()
}

// SaveIndustrySolution 新增或修改方案字典
func (*Solution) SaveIndustrySolution(req *model.IndustrySolutionSaveReq) (*model.IndustrySolution, error) {
	if !schemaNameRe.MatchString(req.SchemaName) {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"schema_name": "schema_name must match ^sol_[a-z0-9_]+$",
		})
	}

	if req.ID == "" {
		exist, err := dal.GetIndustrySolutionByCode(req.Code)
		if err == nil && exist != nil {
			return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
				"code": "solution code already exists",
			})
		}
		item := &model.IndustrySolution{
			ID:          uuid.New(),
			Code:        req.Code,
			Name:        req.Name,
			Description: StringPtr(req.Description),
			SchemaName:  req.SchemaName,
			Icon:        StringPtr(req.Icon),
			Status:      req.Status,
			Sort:        req.Sort,
			Remark:      StringPtr(req.Remark),
		}
		if err := dal.CreateIndustrySolution(item); err != nil {
			logrus.Error("新增行业方案失败:", err)
			return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
				"operation": "create_industry_solution",
				"error":     err.Error(),
			})
		}
		return item, nil
	}

	exist, err := dal.GetIndustrySolutionByID(req.ID)
	if err != nil || exist == nil {
		return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"id": "industry solution not found",
		})
	}
	exist.Name = req.Name
	exist.Description = StringPtr(req.Description)
	exist.SchemaName = req.SchemaName
	exist.Icon = StringPtr(req.Icon)
	exist.Status = req.Status
	exist.Sort = req.Sort
	exist.Remark = StringPtr(req.Remark)
	if err := dal.UpdateIndustrySolution(exist); err != nil {
		logrus.Error("修改行业方案失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "update_industry_solution",
			"error":     err.Error(),
		})
	}
	return exist, nil
}

// DeleteIndustrySolution 删除方案字典
func (*Solution) DeleteIndustrySolution(id string) error {
	if err := dal.DeleteIndustrySolution(id); err != nil {
		logrus.Error("删除行业方案失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "delete_industry_solution",
			"error":     err.Error(),
		})
	}
	return nil
}

// Assign 给租户分配方案（或调整启用状态）；Cascade 为真时沿租户层级下放给全部下级
func (*Solution) Assign(req *model.TenantSolutionAssignReq) error {
	sol, err := dal.GetIndustrySolutionByCode(req.SolutionCode)
	if err != nil || sol == nil {
		return errcode.WithData(errcode.CodeParamError, map[string]interface{}{
			"solution_code": "industry solution not found",
		})
	}

	targets := []string{req.TenantID}
	if req.Cascade {
		ids, err := dal.ListDescendantTenantIDs(req.TenantID)
		if err != nil {
			logrus.Error("查询下级租户失败:", err)
			return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
				"operation": "list_descendant_tenants",
				"error":     err.Error(),
			})
		}
		if len(ids) > 0 {
			targets = ids
		}
	}

	for _, tenantID := range targets {
		item := &model.TenantSolution{
			ID:           uuid.New(),
			TenantID:     tenantID,
			SolutionCode: req.SolutionCode,
			Enabled:      req.Enabled,
		}
		if err := dal.UpsertTenantSolution(item); err != nil {
			logrus.Error("分配方案给租户失败:", err)
			return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
				"operation": "assign_tenant_solution",
				"error":     err.Error(),
			})
		}
	}
	return nil
}

// Unassign 取消租户的方案分配
func (*Solution) Unassign(tenantID, solutionCode string) error {
	if err := dal.DeleteTenantSolution(tenantID, solutionCode); err != nil {
		logrus.Error("取消租户方案分配失败:", err)
		return errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "unassign_tenant_solution",
			"error":     err.Error(),
		})
	}
	return nil
}

// ListAssignments 查询某租户已分配方案（含方案元信息）
func (*Solution) ListAssignments(tenantID string) ([]*model.TenantSolutionResp, error) {
	assigned, err := dal.ListTenantSolutions(tenantID)
	if err != nil {
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_tenant_solutions",
			"error":     err.Error(),
		})
	}

	result := make([]*model.TenantSolutionResp, 0, len(assigned))
	for _, ts := range assigned {
		sol, err := dal.GetIndustrySolutionByCode(ts.SolutionCode)
		if err != nil || sol == nil {
			continue
		}
		result = append(result, &model.TenantSolutionResp{
			SolutionCode: ts.SolutionCode,
			Name:         sol.Name,
			Description:  stringValue(sol.Description),
			Icon:         stringValue(sol.Icon),
			SchemaName:   sol.SchemaName,
			Enabled:      ts.Enabled,
		})
	}
	return result, nil
}

// Mine 租户查看本租户已启用方案及其专属运营数据概览
func (s *Solution) Mine(tenantID string) ([]*model.TenantSolutionResp, error) {
	list, err := s.ListAssignments(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.TenantSolutionResp, 0, len(list))
	for _, item := range list {
		if item.Enabled != 1 {
			continue
		}
		item.DeviceCount = countSolutionData(tenantID, item.SolutionCode)
		result = append(result, item)
	}
	return result, nil
}

// countSolutionData 统计某方案在本租户下的业务数据量（当前仅宠物医疗有业务表）。
func countSolutionData(tenantID, solutionCode string) int64 {
	switch solutionCode {
	case model.SolutionCodePetMedical:
		count, err := dal.CountPetProfiles(tenantID)
		if err != nil {
			logrus.Warn("统计宠物方案运营数据失败:", err)
			return 0
		}
		return count
	default:
		return 0
	}
}
