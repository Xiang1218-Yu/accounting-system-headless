package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// VoucherService 凭证录入与管理
type VoucherService struct {
	st     *store.Store
	audit  *AuditService
	engine *PostingEngine
}

// NewVoucherService 构造
func NewVoucherService(st *store.Store, audit *AuditService, engine *PostingEngine) *VoucherService {
	return &VoucherService{st: st, audit: audit, engine: engine}
}

// List 凭证列表
func (s *VoucherService) List() []*domain.Voucher { return s.st.ListVouchers() }

// Get 取凭证
func (s *VoucherService) Get(id string) *domain.Voucher { return s.st.GetVoucher(id) }

// Create 录入凭证（草稿，允许暂不平衡）
func (s *VoucherService) Create(v *domain.Voucher, opUser string) error {
	if err := s.validateEntries(v, false); err != nil {
		return err
	}
	if v.ID == "" {
		v.ID = uuid.NewString()
	}
	v.Status = domain.StatusDraft
	if v.Type == "" {
		v.Type = domain.VoucherNormal
	}
	if v.VoucherDate.IsZero() {
		v.VoucherDate = time.Now()
	}
	v.PeriodYear = v.VoucherDate.Year()
	v.PeriodMonth = int(v.VoucherDate.Month())
	v.Maker = opUser
	opID := NewOpID()
	return s.st.Mutate(func() {
		s.st.SetVoucher(v)
		s.audit.Log(opUser, "create", "voucher", v.ID, opID, "", jsonOf(v))
	})
}

// Update 修改凭证（仅草稿/已审核未过账可改）
func (s *VoucherService) Update(v *domain.Voucher, opUser string) error {
	old := s.st.GetVoucher(v.ID)
	if old == nil {
		return errors.New("凭证不存在")
	}
	if old.Status == domain.StatusPosted {
		return errors.New("已过账凭证不可修改")
	}
	if err := s.validateEntries(v, false); err != nil {
		return err
	}
	v.PeriodYear = v.VoucherDate.Year()
	v.PeriodMonth = int(v.VoucherDate.Month())
	v.Status = domain.StatusDraft // 修改后回到草稿
	opID := NewOpID()
	before := jsonOf(old)
	return s.st.Mutate(func() {
		s.st.SetVoucher(v)
		s.audit.Log(opUser, "update", "voucher", v.ID, opID, before, jsonOf(v))
	})
}

// Delete 删除凭证（未过账）
func (s *VoucherService) Delete(id, opUser string) error {
	old := s.st.GetVoucher(id)
	if old == nil {
		return errors.New("凭证不存在")
	}
	if old.Status == domain.StatusPosted {
		return errors.New("已过账凭证不可删除")
	}
	opID := NewOpID()
	before := jsonOf(old)
	return s.st.Mutate(func() {
		s.st.DeleteVoucher(id)
		s.audit.Log(opUser, "delete", "voucher", id, opID, before, "")
	})
}

// Approve 审核：草稿→已审核（须平衡）
func (s *VoucherService) Approve(id, opUser string) error {
	v := s.st.GetVoucher(id)
	if v == nil {
		return errors.New("凭证不存在")
	}
	if v.Status != domain.StatusDraft {
		return errors.New("仅草稿凭证可审核")
	}
	if err := s.validateEntries(v, true); err != nil {
		return err
	}
	opID := NewOpID()
	before := jsonOf(v)
	return s.st.Mutate(func() {
		v.Status = domain.StatusAudited
		v.Auditor = opUser
		s.st.SetVoucher(v)
		s.audit.Log(opUser, "approve", "voucher", id, opID, before, jsonOf(v))
	})
}

// Unapprove 反审核：已审核→草稿
func (s *VoucherService) Unapprove(id, opUser string) error {
	v := s.st.GetVoucher(id)
	if v == nil {
		return errors.New("凭证不存在")
	}
	if v.Status != domain.StatusAudited {
		return errors.New("仅已审核凭证可反审核")
	}
	opID := NewOpID()
	before := jsonOf(v)
	return s.st.Mutate(func() {
		v.Status = domain.StatusDraft
		v.Auditor = ""
		s.st.SetVoucher(v)
		s.audit.Log(opUser, "unapprove", "voucher", id, opID, before, jsonOf(v))
	})
}

// Post 过账：委托过账引擎
func (s *VoucherService) Post(id, opUser string) error {
	return s.engine.Post(id, opUser)
}

// Unpost 反过账（需期间开放）
func (s *VoucherService) Unpost(id, opUser string) error {
	return s.engine.Unpost(id, opUser)
}

// validateEntries 校验分录合法性。requireBalance 为 true 时强制借贷平衡
func (s *VoucherService) validateEntries(v *domain.Voucher, requireBalance bool) error {
	if len(v.Entries) < 2 {
		return errors.New("凭证至少需两行分录")
	}
	totalD, totalC := decimal.Zero, decimal.Zero
	for i, e := range v.Entries {
		acc := s.st.GetAccount(e.AccountCode)
		if acc == nil {
			return fmt.Errorf("第%d行：科目 %s 不存在", i+1, e.AccountCode)
		}
		if !acc.IsActive {
			return fmt.Errorf("第%d行：科目 %s 已停用", i+1, e.AccountCode)
		}
		if !acc.IsLeaf {
			return fmt.Errorf("第%d行：科目 %s 非末级", i+1, e.AccountCode)
		}
		if e.Debit.GreaterThan(decimal.Zero) && e.Credit.GreaterThan(decimal.Zero) {
			return fmt.Errorf("第%d行：借贷不能同时有值", i+1)
		}
		if e.Debit.LessThan(decimal.Zero) || e.Credit.LessThan(decimal.Zero) {
			return fmt.Errorf("第%d行：金额不能为负", i+1)
		}
		// 辅助核算引用校验
		for at, itemID := range e.AuxRefs {
			if !hasAuxType(acc.AuxTypes, at) {
				return fmt.Errorf("第%d行：科目 %s 未启用 %s 辅助核算", i+1, e.AccountCode, at)
			}
			if s.st.GetAuxItem(itemID) == nil {
				return fmt.Errorf("第%d行：辅助核算项 %s 不存在", i+1, itemID)
			}
		}
		totalD = totalD.Add(e.Debit)
		totalC = totalC.Add(e.Credit)
	}
	if requireBalance {
		if !totalD.Equal(totalC) {
			return fmt.Errorf("借贷不平衡：借 %s / 贷 %s", totalD.String(), totalC.String())
		}
		if totalD.IsZero() {
			return errors.New("凭证金额不能为零")
		}
	}
	return nil
}

func hasAuxType(types []domain.AuxType, t domain.AuxType) bool {
	for _, x := range types {
		if x == t {
			return true
		}
	}
	return false
}
