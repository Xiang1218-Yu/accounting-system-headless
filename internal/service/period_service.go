package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// PeriodService 会计期间管理
type PeriodService struct {
	st    *store.Store
	audit *AuditService
}

// NewPeriodService 构造
func NewPeriodService(st *store.Store, audit *AuditService) *PeriodService {
	return &PeriodService{st: st, audit: audit}
}

// List 期间列表
func (s *PeriodService) List() []*domain.Period { return s.st.ListPeriods() }

// Get 取期间（不存在返回 nil）
func (s *PeriodService) Get(year, month int) *domain.Period { return s.st.GetPeriod(year, month) }

// EnsurePeriod 确保期间存在（不存在则创建为开放）
func (s *PeriodService) EnsurePeriod(year, month int) *domain.Period {
	p := s.st.GetPeriod(year, month)
	if p != nil {
		return p
	}
	p = &domain.Period{Year: year, Month: month, Status: domain.PeriodOpen}
	_ = s.st.Mutate(func() {
		s.st.SetPeriod(p)
	})
	return p
}

// SetFXRate 设置期间期末汇率快照
func (s *PeriodService) SetFXRate(year, month int, rates map[string]float64, opUser string) error {
	p := s.EnsurePeriod(year, month)
	if p.Status == domain.PeriodClosed {
		return errors.New("期间已结账，无法修改汇率")
	}
	opID := NewOpID()
	before := jsonOf(p)
	return s.st.Mutate(func() {
		p.FXRateSnapshot = rates
		s.st.SetPeriod(p)
		s.audit.Log(opUser, "set_fx_rate", "period", p.Key(), opID, before, jsonOf(p))
	})
}

// Close 月末结账：要求该期间所有凭证已过账
func (s *PeriodService) Close(year, month int, opUser string) error {
	p := s.st.GetPeriod(year, month)
	if p == nil {
		return errors.New("期间不存在")
	}
	if p.Status == domain.PeriodClosed {
		return errors.New("期间已结账")
	}
	for _, v := range s.st.ListVouchers() {
		if v.PeriodYear == year && v.PeriodMonth == month && v.Status != domain.StatusPosted {
			return fmt.Errorf("凭证 %s 未过账，无法结账", v.Number)
		}
	}
	opID := NewOpID()
	before := jsonOf(p)
	return s.st.Mutate(func() {
		now := time.Now()
		p.Status = domain.PeriodClosed
		p.ClosedAt = &now
		p.ClosedBy = opUser
		s.st.SetPeriod(p)
		s.audit.Log(opUser, "close_period", "period", p.Key(), opID, before, jsonOf(p))
	})
}

// Reopen 反结账
func (s *PeriodService) Reopen(year, month int, opUser string) error {
	p := s.st.GetPeriod(year, month)
	if p == nil {
		return errors.New("期间不存在")
	}
	if p.Status != domain.PeriodClosed {
		return errors.New("期间未结账")
	}
	opID := NewOpID()
	before := jsonOf(p)
	return s.st.Mutate(func() {
		p.Status = domain.PeriodOpen
		p.ClosedAt = nil
		p.ClosedBy = ""
		s.st.SetPeriod(p)
		s.audit.Log(opUser, "reopen_period", "period", p.Key(), opID, before, jsonOf(p))
	})
}

// Current 当前期间
func (s *PeriodService) Current() (int, int) {
	m := s.st.Meta()
	return m.CurrentYear, m.CurrentMonth
}

// SetCurrent 设置当前期间
func (s *PeriodService) SetCurrent(year, month int, opUser string) error {
	m := s.st.Meta()
	m.CurrentYear = year
	m.CurrentMonth = month
	opID := NewOpID()
	return s.st.Mutate(func() {
		s.st.SetMeta(m)
		s.audit.Log(opUser, "set_current_period", "meta", fmt.Sprintf("%04d-%02d", year, month), opID, "", "")
	})
}
