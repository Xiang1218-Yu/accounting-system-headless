package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// PeriodBalance 科目期间余额（含辅助核算维度）
type PeriodBalance struct {
	Year          int             `json:"year"`
	Month         int             `json:"month"`
	AccountCode   string          `json:"account_code"`
	AuxKey        string          `json:"aux_key"` // 辅助核算组合键，无辅助时为空
	OpeningDebit  decimal.Decimal `json:"opening_debit"`
	OpeningCredit decimal.Decimal `json:"opening_credit"`
	PeriodDebit   decimal.Decimal `json:"period_debit"`  // 期间发生额（永久保留）
	PeriodCredit  decimal.Decimal `json:"period_credit"` // 期间发生额（永久保留）
	ClosingDebit  decimal.Decimal `json:"closing_debit"`
	ClosingCredit decimal.Decimal `json:"closing_credit"`
	// 外币余额（仅外币科目使用）
	FCOpeningDebit  decimal.Decimal `json:"fc_opening_debit,omitempty"`
	FCOpeningCredit decimal.Decimal `json:"fc_opening_credit,omitempty"`
	FCPeriodDebit   decimal.Decimal `json:"fc_period_debit,omitempty"`
	FCPeriodCredit  decimal.Decimal `json:"fc_period_credit,omitempty"`
	FCClosingDebit  decimal.Decimal `json:"fc_closing_debit,omitempty"`
	FCClosingCredit decimal.Decimal `json:"fc_closing_credit,omitempty"`
}

// Key 余额记录主键
func (b PeriodBalance) Key() string {
	return fmt.Sprintf("%04d-%02d|%s|%s", b.Year, b.Month, b.AccountCode, b.AuxKey)
}

// NetDebit 借方净额
func (b PeriodBalance) NetDebit() decimal.Decimal {
	return b.ClosingDebit.Sub(b.ClosingCredit)
}

// NetCredit 贷方净额
func (b PeriodBalance) NetCredit() decimal.Decimal {
	return b.ClosingCredit.Sub(b.ClosingDebit)
}
