package service

import (
	dal "project/internal/dal"
	model "project/internal/model"
	"project/pkg/errcode"

	"github.com/go-basic/uuid"
	"github.com/sirupsen/logrus"
)

// Tenant 租户层级业务服务（经销商 → 宠物医院 → 门店）。
// ThingsPanel 无原生 tenants 表，本服务基于 public.tenants 叠加层维护层级关系。
type Tenant struct{}

// Tree 返回租户层级树（按 parent_id 组装）
func (*Tenant) Tree() ([]*model.TenantNode, error) {
	all, err := dal.ListTenants()
	if err != nil {
		logrus.Error("查询租户层级失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "list_tenants",
			"error":     err.Error(),
		})
	}

	nodes := make(map[string]*model.TenantNode, len(all))
	for _, t := range all {
		nodes[t.ID] = &model.TenantNode{
			TenantID: t.ID,
			Name:     stringValue(t.Name),
			Role:     stringValue(t.Role),
			Children: make([]*model.TenantNode, 0),
		}
		if t.ParentID != nil {
			nodes[t.ID].ParentID = *t.ParentID
		}
	}

	roots := make([]*model.TenantNode, 0)
	for _, t := range all {
		node := nodes[t.ID]
		if t.ParentID != nil && *t.ParentID != "" {
			if parent, ok := nodes[*t.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots, nil
}

// Save 新增或修改租户层级节点
func (*Tenant) Save(req *model.TenantSaveReq) (*model.TenantNode, error) {
	id := req.ID
	if id == "" {
		id = uuid.New()
	}

	if req.ParentID != "" {
		if parent, err := dal.GetTenant(req.ParentID); err != nil || parent == nil {
			return nil, errcode.WithData(errcode.CodeParamError, map[string]interface{}{
				"parent_id": "parent tenant not found",
			})
		}
	}

	role := req.Role
	if role == "" {
		role = "tenant"
	}
	item := &model.Tenant{
		ID:       id,
		ParentID: StringPtr(req.ParentID),
		Name:     StringPtr(req.Name),
		Role:     StringPtr(role),
		Status:   1,
		Remark:   StringPtr(req.Remark),
	}
	if err := dal.UpsertTenant(item); err != nil {
		logrus.Error("保存租户层级失败:", err)
		return nil, errcode.WithData(errcode.CodeDBError, map[string]interface{}{
			"operation": "upsert_tenant",
			"error":     err.Error(),
		})
	}
	return &model.TenantNode{
		TenantID: item.ID,
		ParentID: req.ParentID,
		Name:     req.Name,
		Role:     role,
		Children: make([]*model.TenantNode, 0),
	}, nil
}
