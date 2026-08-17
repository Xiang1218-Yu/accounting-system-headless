package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// ClosingService 结转损益
type ClosingService struct {
	st     *store.Store
	ledger *LedgerService
	engine *PostingEngine
	audit  *AuditService
}

// NewClosingService 构造
func NewClosingService(st *store.Store, ledger *LedgerService, engine *PostingEngine, audit *AuditService) *ClosingService {
	return &ClosingService{st: st, ledger: ledger, engine: engine, audit: audit}
}

// HasClosed 本期是否已结转损益
func (c *ClosingService) HasClosed(year, month int) bool {
	for _, v := range c.st.ListVouchers() {
		if v.Type == domain.VoucherClosing && v.PeriodYear == year && v.PeriodMonth == month && v.Status == domain.StatusPosted {
			return true
		}
	}
	return false
}

// CloseProfitLoss 结转本期损益到本年利润(3103)
func (c *ClosingService) CloseProfitLoss(year, month int, opUser string) error {
	if c.HasClosed(year, month) {
		return errors.New("本期已结转损益")
	}
	p := c.st.GetPeriod(year, month)
	if p != nil && p.Status == domain.PeriodClosed {
		return errors.New("期间已结账，无法结转")
	}

	var entries []domain.VoucherEntry
	revenueTotal := decimal.Zero
	expenseTotal := decimal.Zero

	for _, acc := range c.st.ListAccounts() {
		if acc.Category != domain.CategoryProfitLoss || !acc.IsActive {
			continue
		}
		if acc.Code == "3103" {
			continue
		}
		cd, cc := c.ledger.closingAsOf(year, month, acc.Code)
		if acc.Direction == domain.DirCredit {
			// 收入类：贷方净额 → 借该科目 / 贷本年利润
			net := cc.Sub(cd)
			if net.IsZero() {
				continue
			}
			entries = append(entries, domain.VoucherEntry{AccountCode: acc.Code, Summary: "结转收入", Debit: net})
			revenueTotal = revenueTotal.Add(net)
		} else {
			// 费用类：借方净额 → 借本年利润 / 贷该科目
			net := cd.Sub(cc)
			if net.IsZero() {
				continue
			}
			entries = append(entries, domain.VoucherEntry{AccountCode: acc.Code, Summary: "结转费用", Credit: net})
			expenseTotal = expenseTotal.Add(net)
		}
	}

	if len(entries) == 0 {
		return errors.New("本期无可结转的损益余额")
	}

	// 本年利润：贷方记收入结转额，借方记费用结转额
	if revenueTotal.GreaterThan(decimal.Zero) {
		entries = append(entries, domain.VoucherEntry{AccountCode: "3103", Summary: "结转收入至本年利润", Credit: revenueTotal})
	}
	if expenseTotal.GreaterThan(decimal.Zero) {
		entries = append(entries, domain.VoucherEntry{AccountCode: "3103", Summary: "结转费用至本年利润", Debit: expenseTotal})
	}

	date := lastDayOfMonth(year, month)
	voucher := &domain.Voucher{
		ID:          uuid.NewString(),
		PeriodYear:  year,
		PeriodMonth: month,
		VoucherDate: date,
		Status:      domain.StatusAudited,
		Type:        domain.VoucherClosing,
		Summary:     "期末结转损益",
		Maker:       opUser,
		Auditor:     opUser,
		Entries:     entries,
	}

	opID := NewOpID()
	if err := c.st.Mutate(func() {
		c.st.SetVoucher(voucher)
		c.audit.Log(opUser, "create_closing", "voucher", voucher.ID, opID, "", jsonOf(voucher))
	}); err != nil {
		return err
	}
	return c.engine.Post(voucher.ID, opUser)
}

func lastDayOfMonth(year, month int) time.Time {
	if month == 12 {
		return time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)
	}
	return time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
}
