package domain

// AuxiliaryItem 辅助核算项（客户/供应商/部门/项目）
type AuxiliaryItem struct {
	ID       string  `json:"id"`
	Type     AuxType `json:"type"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	ParentID string  `json:"parent_id,omitempty"`
	IsActive bool    `json:"is_active"`
}
