package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// FXService 期末调汇
type FXService struct {
	st     *store.Store
	ledger *LedgerService
	engine *PostingEngine
	audit  *AuditService
}

// NewFXService 构造
func NewFXService(st *store.Store, ledger *LedgerService, engine *PostingEngine, audit *AuditService) *FXService {
	return &FXService{st: st, ledger: ledger, engine: engine, audit: audit}
}

// HasAdjusted 本期是否已调汇
func (f *FXService) HasAdjusted(year, month int) bool {
	for _, v := range f.st.ListVouchers() {
		if v.Type == domain.VoucherFX && v.PeriodYear == year && v.PeriodMonth == month && v.Status == domain.StatusPosted {
			return true
		}
	}
	return false
}

// AdjustFX 期末调汇：对外币资产/负债科目按期末汇率重估本币，差额入财务费用-汇兑损益
func (f *FXService) AdjustFX(year, month int, opUser string) error {
	if f.HasAdjusted(year, month) {
		return errors.New("本期已调汇")
	}
	p := f.st.GetPeriod(year, month)
	if p == nil || p.FXRateSnapshot == nil || len(p.FXRateSnapshot) == 0 {
		return errors.New("请先设置期末汇率")
	}

	// 聚合本期间各外币科目的外币/本币期末余额（跨辅助维度）
	type agg struct {
		fcDebit, fcCredit, rmbDebit, rmbCredit decimal.Decimal
	}
	aggs := map[string]*agg{}
	for _, b := range f.st.ListBalances() {
		if b.Year != year || b.Month != month {
			continue
		}
		acc := f.st.GetAccount(b.AccountCode)
		if acc == nil || !acc.IsForeignCurrency {
			continue
		}
		a := aggs[b.AccountCode]
		if a == nil {
			a = &agg{}
			aggs[b.AccountCode] = a
		}
		a.fcDebit = a.fcDebit.Add(b.FCClosingDebit)
		a.fcCredit = a.fcCredit.Add(b.FCClosingCredit)
		a.rmbDebit = a.rmbDebit.Add(b.ClosingDebit)
		a.rmbCredit = a.rmbCredit.Add(b.ClosingCredit)
	}

	var entries []domain.VoucherEntry
	for code, a := range aggs {
		acc := f.st.GetAccount(code)
		rate, ok := p.FXRateSnapshot[acc.Currency]
		if !ok || rate <= 0 {
			continue
		}
		rateDec := decimal.NewFromFloat(rate)

		var fcQty, rmbBook decimal.Decimal
		if acc.Direction == domain.DirDebit {
			fcQty = a.fcDebit.Sub(a.fcCredit) // 持有的外币资产
			rmbBook = a.rmbDebit.Sub(a.rmbCredit)
		} else {
			fcQty = a.fcCredit.Sub(a.fcDebit) // 外币负债
			rmbBook = a.rmbCredit.Sub(a.rmbDebit)
		}
		if fcQty.IsZero() {
			continue
		}
		revalued := fcQty.Mul(rateDec)
		diff := revalued.Sub(rmbBook)
		if diff.IsZero() {
			continue
		}
		absDiff := diff.Abs()
		if acc.Direction == domain.DirDebit {
			// 资产：差额>0 增值 → 借资产/贷财务费用；<0 → 借财务费用/贷资产
			if diff.GreaterThan(decimal.Zero) {
				entries = append(entries, domain.VoucherEntry{AccountCode: code, Summary: "期末调汇", Debit: absDiff})
				entries = append(entries, domain.VoucherEntry{AccountCode: "5603", Summary: "汇兑收益", Credit: absDiff})
			} else {
				entries = append(entries, domain.VoucherEntry{AccountCode: code, Summary: "期末调汇", Credit: absDiff})
				entries = append(entries, domain.VoucherEntry{AccountCode: "5603", Summary: "汇兑损失", Debit: absDiff})
			}
		} else {
			// 负债：差额>0 增值 → 贷负债/借财务费用；<0 → 借负债/贷财务费用
			if diff.GreaterThan(decimal.Zero) {
				entries = append(entries, domain.VoucherEntry{AccountCode: code, Summary: "期末调汇", Credit: absDiff})
				entries = append(entries, domain.VoucherEntry{AccountCode: "5603", Summary: "汇兑损失", Debit: absDiff})
			} else {
				entries = append(entries, domain.VoucherEntry{AccountCode: code, Summary: "期末调汇", Debit: absDiff})
				entries = append(entries, domain.VoucherEntry{AccountCode: "5603", Summary: "汇兑收益", Credit: absDiff})
			}
		}
	}

	if len(entries) == 0 {
		return errors.New("本期无需调汇的外币余额")
	}

	voucher := &domain.Voucher{
		ID:          uuid.NewString(),
		PeriodYear:  year,
		PeriodMonth: month,
		VoucherDate: lastDayOfMonth(year, month),
		Status:      domain.StatusAudited,
		Type:        domain.VoucherFX,
		Summary:     fmt.Sprintf("%04d年%02d月期末调汇", year, month),
		Maker:       opUser,
		Auditor:     opUser,
		Entries:     entries,
	}
	opID := NewOpID()
	if err := f.st.Mutate(func() {
		f.st.SetVoucher(voucher)
		f.audit.Log(opUser, "create_fx", "voucher", voucher.ID, opID, "", jsonOf(voucher))
	}); err != nil {
		return err
	}
	return f.engine.Post(voucher.ID, opUser)
}
