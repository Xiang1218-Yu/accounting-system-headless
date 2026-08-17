package service

import (
	"errors"
	"strings"

	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// AccountService 科目管理
type AccountService struct {
	st    *store.Store
	audit *AuditService
}

// NewAccountService 构造
func NewAccountService(st *store.Store, audit *AuditService) *AccountService {
	return &AccountService{st: st, audit: audit}
}

// List 全部科目（已按编码排序）
func (a *AccountService) List() []*domain.Account { return a.st.ListAccounts() }

// Get 取单个科目
func (a *AccountService) Get(code string) *domain.Account { return a.st.GetAccount(code) }

// Create 新增科目
func (a *AccountService) Create(acc *domain.Account, opUser string) error {
	if err := validateAccount(acc); err != nil {
		return err
	}
	if a.st.GetAccount(acc.Code) != nil {
		return errors.New("科目编码已存在: " + acc.Code)
	}
	if acc.ParentCode != "" && a.st.GetAccount(acc.ParentCode) == nil {
		return errors.New("上级科目不存在: " + acc.ParentCode)
	}
	acc.Category = domain.CategoryByCode(acc.Code)
	opID := NewOpID()
	return a.st.Mutate(func() {
		a.st.SetAccount(acc)
		if acc.ParentCode != "" {
			if p := a.st.AccountNL(acc.ParentCode); p != nil {
				p.IsLeaf = false
				a.st.SetAccount(p)
			}
		}
		a.audit.Log(opUser, "create", "account", acc.Code, opID, "", jsonOf(acc))
	})
}

// Update 修改科目（仅允许改名称/方向/辅助核算/外币/启用状态，编码不可改）
func (a *AccountService) Update(acc *domain.Account, opUser string) error {
	if err := validateAccount(acc); err != nil {
		return err
	}
	old := a.st.GetAccount(acc.Code)
	if old == nil {
		return errors.New("科目不存在: " + acc.Code)
	}
	opID := NewOpID()
	before := jsonOf(old)
	return a.st.Mutate(func() {
		merged := *old
		merged.Name = acc.Name
		merged.Direction = acc.Direction
		merged.AuxTypes = acc.AuxTypes
		merged.IsForeignCurrency = acc.IsForeignCurrency
		merged.Currency = acc.Currency
		merged.IsActive = acc.IsActive
		merged.Level = acc.Level
		merged.ParentCode = acc.ParentCode
		merged.IsLeaf = acc.IsLeaf
		a.st.SetAccount(&merged)
		a.audit.Log(opUser, "update", "account", acc.Code, opID, before, jsonOf(merged))
	})
}

// Delete 删除科目（需无子科目、无凭证引用）
func (a *AccountService) Delete(code, opUser string) error {
	acc := a.st.GetAccount(code)
	if acc == nil {
		return errors.New("科目不存在: " + code)
	}
	for _, c := range a.st.ListAccounts() {
		if c.ParentCode == code {
			return errors.New("存在下级科目，无法删除")
		}
	}
	for _, v := range a.st.ListVouchers() {
		for _, e := range v.Entries {
			if e.AccountCode == code {
				return errors.New("科目已被凭证引用，无法删除")
			}
		}
	}
	opID := NewOpID()
	before := jsonOf(acc)
	return a.st.Mutate(func() {
		a.st.DeleteAccount(code)
		a.audit.Log(opUser, "delete", "account", code, opID, before, "")
	})
}

func validateAccount(acc *domain.Account) error {
	if strings.TrimSpace(acc.Code) == "" {
		return errors.New("科目编码不能为空")
	}
	if strings.TrimSpace(acc.Name) == "" {
		return errors.New("科目名称不能为空")
	}
	if acc.Level <= 0 {
		acc.Level = 1
	}
	if acc.Direction != domain.DirDebit && acc.Direction != domain.DirCredit {
		return errors.New("余额方向无效")
	}
	return nil
}
