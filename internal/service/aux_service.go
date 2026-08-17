package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// AuxService 辅助核算管理
type AuxService struct {
	st    *store.Store
	audit *AuditService
}

// NewAuxService 构造
func NewAuxService(st *store.Store, audit *AuditService) *AuxService {
	return &AuxService{st: st, audit: audit}
}

// List 辅助项列表（按类型过滤，空则全部）
func (s *AuxService) List(t domain.AuxType) []*domain.AuxiliaryItem { return s.st.ListAuxItems(t) }

// Get 取辅助项
func (s *AuxService) Get(id string) *domain.AuxiliaryItem { return s.st.GetAuxItem(id) }

// Create 新增辅助项
func (s *AuxService) Create(item *domain.AuxiliaryItem, opUser string) error {
	if err := validateAux(item); err != nil {
		return err
	}
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	item.IsActive = true
	opID := NewOpID()
	return s.st.Mutate(func() {
		s.st.SetAuxItem(item)
		s.audit.Log(opUser, "create", "auxitem", item.ID, opID, "", jsonOf(item))
	})
}

// Update 修改辅助项
func (s *AuxService) Update(item *domain.AuxiliaryItem, opUser string) error {
	if err := validateAux(item); err != nil {
		return err
	}
	old := s.st.GetAuxItem(item.ID)
	if old == nil {
		return errors.New("辅助项不存在")
	}
	opID := NewOpID()
	before := jsonOf(old)
	return s.st.Mutate(func() {
		merged := *old
		merged.Code = item.Code
		merged.Name = item.Name
		merged.ParentID = item.ParentID
		merged.IsActive = item.IsActive
		s.st.SetAuxItem(&merged)
		s.audit.Log(opUser, "update", "auxitem", item.ID, opID, before, jsonOf(merged))
	})
}

// Delete 删除辅助项（无凭证引用）
func (s *AuxService) Delete(id, opUser string) error {
	old := s.st.GetAuxItem(id)
	if old == nil {
		return errors.New("辅助项不存在")
	}
	for _, v := range s.st.ListVouchers() {
		for _, e := range v.Entries {
			for _, refID := range e.AuxRefs {
				if refID == id {
					return errors.New("辅助项已被凭证引用，无法删除")
				}
			}
		}
	}
	opID := NewOpID()
	before := jsonOf(old)
	return s.st.Mutate(func() {
		s.st.DeleteAuxItem(id)
		s.audit.Log(opUser, "delete", "auxitem", id, opID, before, "")
	})
}

func validateAux(item *domain.AuxiliaryItem) error {
	if item.Type == "" {
		return errors.New("辅助核算类型不能为空")
	}
	if strings.TrimSpace(item.Code) == "" {
		return errors.New("辅助项编码不能为空")
	}
	if strings.TrimSpace(item.Name) == "" {
		return errors.New("辅助项名称不能为空")
	}
	return nil
}

// AuxBalanceRow 辅助余额表行
type AuxBalanceRow struct {
	ItemID      string
	ItemCode    string
	ItemName    string
	AccountCode string
	AccountName string
	Debit       decimal.Decimal
	Credit      decimal.Decimal
}

// AuxBalanceReport 辅助余额表：按辅助类型 + 科目汇总某期间期末余额
func (s *AuxService) AuxBalanceReport(year, month int, t domain.AuxType, accountCode string) []AuxBalanceRow {
	rows := []AuxBalanceRow{}
	for _, b := range s.st.ListBalances() {
		if b.Year != year || b.Month != month {
			continue
		}
		itemID := extractAuxRef(b.AuxKey, t)
		if itemID == "" {
			continue
		}
		if accountCode != "" && b.AccountCode != accountCode {
			continue
		}
		acc := s.st.GetAccount(b.AccountCode)
		item := s.st.GetAuxItem(itemID)
		row := AuxBalanceRow{
			ItemID:      itemID,
			AccountCode: b.AccountCode,
			Debit:       b.ClosingDebit,
			Credit:      b.ClosingCredit,
		}
		if acc != nil {
			row.AccountName = acc.Name
		}
		if item != nil {
			row.ItemCode = item.Code
			row.ItemName = item.Name
		}
		rows = append(rows, row)
	}
	return rows
}

// extractAuxRef 从 AuxKey 中提取指定类型的辅助项 ID
func extractAuxRef(auxKey string, t domain.AuxType) string {
	if auxKey == "" {
		return ""
	}
	prefix := string(t) + ":"
	for _, seg := range strings.Split(auxKey, "|") {
		if strings.HasPrefix(seg, prefix) {
			return strings.TrimPrefix(seg, prefix)
		}
	}
	return ""
}
