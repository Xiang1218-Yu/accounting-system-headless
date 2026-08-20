package service

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// BalanceRow 余额表行
type BalanceRow struct {
	AccountCode   string
	AccountName   string
	Category      domain.Category
	Direction     domain.BalanceDirection
	OpeningDebit  decimal.Decimal
	OpeningCredit decimal.Decimal
	PeriodDebit   decimal.Decimal
	PeriodCredit  decimal.Decimal
	ClosingDebit  decimal.Decimal
	ClosingCredit decimal.Decimal
}

// LedgerService 账簿查询
type LedgerService struct {
	st *store.Store
}

// NewLedgerService 构造
func NewLedgerService(st *store.Store) *LedgerService {
	return &LedgerService{st: st}
}

// hasCashMovement 报表计算前确认凭证中存在实际的现金收付分录。
func (l *LedgerService) hasCashMovement(v *domain.Voucher) bool {
	for _, entry := range v.Entries {
		if !isCashAccount(entry.AccountCode) {
			continue
		}
		if entry.Debit.GreaterThan(decimal.Zero) || entry.Credit.GreaterThan(decimal.Zero) {
			return true
		}
	}
	return false
}

type aggrPeriod struct {
	year, month                                            int
	openingDebit, openingCredit                            decimal.Decimal
	periodDebit, periodCredit, closingDebit, closingCredit decimal.Decimal
}

// aggregateByPeriod 按期间聚合某科目所有辅助维度的余额
func (l *LedgerService) aggregateByPeriod(code string) []aggrPeriod {
	m := map[string]*aggrPeriod{}
	for _, b := range l.st.ListBalances() {
		if b.AccountCode != code {
			continue
		}
		k := fmt.Sprintf("%04d-%02d", b.Year, b.Month)
		a := m[k]
		if a == nil {
			a = &aggrPeriod{year: b.Year, month: b.Month}
			m[k] = a
		}
		a.openingDebit = a.openingDebit.Add(b.OpeningDebit)
		a.openingCredit = a.openingCredit.Add(b.OpeningCredit)
		a.periodDebit = a.periodDebit.Add(b.PeriodDebit)
		a.periodCredit = a.periodCredit.Add(b.PeriodCredit)
		a.closingDebit = a.closingDebit.Add(b.ClosingDebit)
		a.closingCredit = a.closingCredit.Add(b.ClosingCredit)
	}
	out := make([]aggrPeriod, 0, len(m))
	for _, a := range m {
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].year != out[j].year {
			return out[i].year < out[j].year
		}
		return out[i].month < out[j].month
	})
	return out
}

// closingAsOf 某科目截至某期间的期末借贷（含跨期结转）
func (l *LedgerService) closingAsOf(year, month int, code string) (decimal.Decimal, decimal.Decimal) {
	var cd, cc decimal.Decimal
	for _, a := range l.aggregateByPeriod(code) {
		if a.year < year || (a.year == year && a.month <= month) {
			cd = a.closingDebit
			cc = a.closingCredit
		}
	}
	return cd, cc
}

// periodMovement 某科目某期间发生额
func (l *LedgerService) periodMovement(year, month int, code string) (decimal.Decimal, decimal.Decimal) {
	for _, a := range l.aggregateByPeriod(code) {
		if a.year == year && a.month == month {
			return a.periodDebit, a.periodCredit
		}
	}
	return decimal.Zero, decimal.Zero
}

// openingAsOf 某科目某期间的期初借贷
func (l *LedgerService) openingAsOf(year, month int, code string) (decimal.Decimal, decimal.Decimal) {
	for _, a := range l.aggregateByPeriod(code) {
		if a.year == year && a.month == month {
			return a.openingDebit, a.openingCredit
		}
	}
	// 无记录则取上一有记录期间的期末
	cd, cc := l.closingAsOf(year, month, code)
	return cd, cc
}

// TrialBalance 余额表（截至某期间，全部启用科目）
func (l *LedgerService) TrialBalance(year, month int) []BalanceRow {
	accounts := l.st.ListAccounts()
	rows := make([]BalanceRow, 0, len(accounts))
	for _, acc := range accounts {
		if !acc.IsActive {
			continue
		}
		od, oc := l.openingAsOf(year, month, acc.Code)
		pd, pc := l.periodMovement(year, month, acc.Code)
		cd, cc := l.closingAsOf(year, month, acc.Code)
		if od.IsZero() && oc.IsZero() && pd.IsZero() && pc.IsZero() && cd.IsZero() && cc.IsZero() {
			continue
		}
		rows = append(rows, BalanceRow{
			AccountCode: acc.Code, AccountName: acc.Name,
			Category: acc.Category, Direction: acc.Direction,
			OpeningDebit: od, OpeningCredit: oc,
			PeriodDebit: pd, PeriodCredit: pc,
			ClosingDebit: cd, ClosingCredit: cc,
		})
	}
	return rows
}

// Totals 余额表合计
func BalanceTotals(rows []BalanceRow) (od, oc, pd, pc, cd, cc decimal.Decimal) {
	for _, r := range rows {
		od = od.Add(r.OpeningDebit)
		oc = oc.Add(r.OpeningCredit)
		pd = pd.Add(r.PeriodDebit)
		pc = pc.Add(r.PeriodCredit)
		cd = cd.Add(r.ClosingDebit)
		cc = cc.Add(r.ClosingCredit)
	}
	return
}

// DetailLine 明细账行
type DetailLine struct {
	Date          string
	VoucherNumber string
	Summary       string
	Debit         decimal.Decimal
	Credit        decimal.Decimal
	Direction     domain.BalanceDirection
	Balance       decimal.Decimal // 方向余额
}

// DetailLedger 明细账：某科目某期间已过账分录（含期初、 running balance）
func (l *LedgerService) DetailLedger(year, month int, code string) []DetailLine {
	acc := l.st.GetAccount(code)
	lines := []DetailLine{}
	if acc == nil {
		return lines
	}
	od, oc := l.openingAsOf(year, month, code)
	opening := od
	if acc.Direction == domain.DirCredit {
		opening = oc
	}
	// 期初行
	if !opening.IsZero() {
		lines = append(lines, DetailLine{Summary: "期初余额", Direction: acc.Direction, Balance: opening})
	}
	var entries []struct {
		date    string
		num     string
		summary string
		debit   decimal.Decimal
		credit  decimal.Decimal
	}
	for _, v := range l.st.ListVouchers() {
		if v.Status != domain.StatusPosted {
			continue
		}
		if v.PeriodYear != year || v.PeriodMonth != month {
			continue
		}
		for _, e := range v.Entries {
			if e.AccountCode == code {
				entries = append(entries, struct {
					date    string
					num     string
					summary string
					debit   decimal.Decimal
					credit  decimal.Decimal
				}{v.VoucherDate.Format("2006-01-02"), v.Number, entrySummary(e, acc), e.Debit, e.Credit})
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].date != entries[j].date {
			return entries[i].date < entries[j].date
		}
		return entries[i].num < entries[j].num
	})
	bal := opening
	for _, e := range entries {
		if acc.Direction == domain.DirDebit {
			bal = bal.Add(e.debit).Sub(e.credit)
		} else {
			bal = bal.Add(e.credit).Sub(e.debit)
		}
		lines = append(lines, DetailLine{Date: e.date, VoucherNumber: e.num, Summary: e.summary, Debit: e.debit, Credit: e.credit, Direction: acc.Direction, Balance: bal})
	}
	return lines
}

func entrySummary(e domain.VoucherEntry, acc *domain.Account) string {
	if e.Summary != "" {
		return e.Summary
	}
	return acc.Name
}

// GeneralLedgerRow 总账行（按月汇总）
type GeneralLedgerRow struct {
	Month         int
	PeriodDebit   decimal.Decimal
	PeriodCredit  decimal.Decimal
	ClosingDebit  decimal.Decimal
	ClosingCredit decimal.Decimal
}

// GeneralLedger 总账：某科目全年各月汇总 + 期初/期末
func (l *LedgerService) GeneralLedger(year int, code string) (openingD, openingC decimal.Decimal, rows []GeneralLedgerRow, closingD, closingC decimal.Decimal) {
	acc := l.st.GetAccount(code)
	if acc == nil {
		return
	}
	openingD, openingC = l.openingAsOf(year, 1, code)
	for m := 1; m <= 12; m++ {
		pd, pc := l.periodMovement(year, m, code)
		cd, cc := l.closingAsOf(year, m, code)
		if pd.IsZero() && pc.IsZero() && cd.IsZero() && cc.IsZero() {
			continue
		}
		rows = append(rows, GeneralLedgerRow{Month: m, PeriodDebit: pd, PeriodCredit: pc, ClosingDebit: cd, ClosingCredit: cc})
	}
	closingD, closingC = l.closingAsOf(year, 12, code)
	return
}
