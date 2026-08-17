package domain

import "github.com/shopspring/decimal"

// VoucherEntry 凭证分录
type VoucherEntry struct {
	AccountCode string             `json:"account_code"`
	Summary     string             `json:"summary,omitempty"`
	Debit       decimal.Decimal    `json:"debit"`
	Credit      decimal.Decimal    `json:"credit"`
	DebitFC     decimal.Decimal    `json:"debit_fc,omitempty"`  // 外币借方（外币科目时填写）
	CreditFC    decimal.Decimal    `json:"credit_fc,omitempty"` // 外币贷方
	AuxRefs     map[AuxType]string `json:"aux_refs,omitempty"`
}
