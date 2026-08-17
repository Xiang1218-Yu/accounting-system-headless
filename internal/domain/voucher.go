package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Voucher 记账凭证
type Voucher struct {
	ID          string         `json:"id"`
	PeriodYear  int            `json:"period_year"`
	PeriodMonth int            `json:"period_month"`
	VoucherDate time.Time      `json:"voucher_date"`
	Number      string         `json:"number"`
	Status      VoucherStatus  `json:"status"`
	Type        VoucherType    `json:"type"`
	Summary     string         `json:"summary,omitempty"`
	Maker       string         `json:"maker,omitempty"`
	Auditor     string         `json:"auditor,omitempty"`
	Poster      string         `json:"poster,omitempty"`
	PostedAt    *time.Time     `json:"posted_at,omitempty"`
	Entries     []VoucherEntry `json:"entries"`
}

// TotalDebit 借方合计
func (v *Voucher) TotalDebit() decimal.Decimal {
	t := decimal.Zero
	for _, e := range v.Entries {
		t = t.Add(e.Debit)
	}
	return t
}

// TotalCredit 贷方合计
func (v *Voucher) TotalCredit() decimal.Decimal {
	t := decimal.Zero
	for _, e := range v.Entries {
		t = t.Add(e.Credit)
	}
	return t
}

// IsBalanced 借贷是否平衡
func (v *Voucher) IsBalanced() bool {
	return v.TotalDebit().Equal(v.TotalCredit()) && v.TotalDebit().GreaterThan(decimal.Zero)
}

// PeriodKey 期间键 YYYY-MM
func (v *Voucher) PeriodKey() string {
	return periodKey(v.PeriodYear, v.PeriodMonth)
}
