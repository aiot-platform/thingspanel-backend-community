package dal

import (
	"time"

	model "project/internal/model"
	global "project/pkg/global"
	solution "project/pkg/solution"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 行业方案数据访问层。平台级表（industry_solutions / tenant_solutions）位于 public schema，
// 使用 global.DB 访问；方案业务数据位于各自 schema，通过 solution.DB(schema) 访问。

// PetDB 返回宠物医疗方案 schema 的数据库连接。
func PetDB() (*gorm.DB, error) {
	return solution.DB(model.SchemaPetMedical)
}

// ---------------------------------------------------------------------------
// 方案字典
// ---------------------------------------------------------------------------

func ListIndustrySolutions() ([]*model.IndustrySolution, error) {
	var list []*model.IndustrySolution
	err := global.DB.Model(&model.IndustrySolution{}).
		Order("sort ASC, code ASC").
		Find(&list).Error
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return list, nil
}

func GetIndustrySolutionByID(id string) (*model.IndustrySolution, error) {
	item := &model.IndustrySolution{}
	err := global.DB.Model(&model.IndustrySolution{}).Where("id = ?", id).First(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func GetIndustrySolutionByCode(code string) (*model.IndustrySolution, error) {
	item := &model.IndustrySolution{}
	err := global.DB.Model(&model.IndustrySolution{}).Where("code = ?", code).First(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func CreateIndustrySolution(item *model.IndustrySolution) error {
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	return global.DB.Model(&model.IndustrySolution{}).Create(item).Error
}

func UpdateIndustrySolution(item *model.IndustrySolution) error {
	return global.DB.Model(&model.IndustrySolution{}).
		Where("id = ?", item.ID).
		Updates(map[string]interface{}{
			"name":        item.Name,
			"description": item.Description,
			"schema_name": item.SchemaName,
			"icon":        item.Icon,
			"status":      item.Status,
			"sort":        item.Sort,
			"remark":      item.Remark,
			"updated_at":  time.Now().UTC(),
		}).Error
}

func DeleteIndustrySolution(id string) error {
	return global.DB.Model(&model.IndustrySolution{}).Where("id = ?", id).Delete(&model.IndustrySolution{}).Error
}

// ---------------------------------------------------------------------------
// 租户列表（ThingsPanel 无 tenants 表，以 TENANT_ADMIN 用户代表租户）
// ---------------------------------------------------------------------------

func ListTenantAdmins() ([]*model.TenantInfo, error) {
	var list []*model.TenantInfo
	err := global.DB.Raw(`
		SELECT tenant_id, email, COALESCE(name, email) AS name
		FROM public.users
		WHERE authority = 'TENANT_ADMIN' AND tenant_id IS NOT NULL AND tenant_id <> ''
		ORDER BY created_at ASC`).Scan(&list).Error
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 方案↔租户 分配
// ---------------------------------------------------------------------------

func ListTenantSolutions(tenantID string) ([]*model.TenantSolution, error) {
	var list []*model.TenantSolution
	err := global.DB.Model(&model.TenantSolution{}).
		Where("tenant_id = ?", tenantID).
		Order("created_at ASC").
		Find(&list).Error
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return list, nil
}

// UpsertTenantSolution 按 (tenant_id, solution_code) 唯一键启用/停用分配关系。
func UpsertTenantSolution(item *model.TenantSolution) error {
	now := time.Now().UTC()
	exist := &model.TenantSolution{}
	err := global.DB.Model(&model.TenantSolution{}).
		Where("tenant_id = ? AND solution_code = ?", item.TenantID, item.SolutionCode).
		First(exist).Error
	if err != nil {
		item.CreatedAt = now
		item.UpdatedAt = now
		return global.DB.Model(&model.TenantSolution{}).Create(item).Error
	}
	return global.DB.Model(&model.TenantSolution{}).
		Where("id = ?", exist.ID).
		Updates(map[string]interface{}{
			"enabled":    item.Enabled,
			"config":     item.Config,
			"updated_at": now,
		}).Error
}

func DeleteTenantSolution(tenantID, solutionCode string) error {
	return global.DB.Model(&model.TenantSolution{}).
		Where("tenant_id = ? AND solution_code = ?", tenantID, solutionCode).
		Delete(&model.TenantSolution{}).Error
}
