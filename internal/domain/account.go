package domain

// Account 会计科目
type Account struct {
	Code              string           `json:"code"`
	Name              string           `json:"name"`
	Category          Category         `json:"category"`
	Direction         BalanceDirection `json:"direction"`
	Level             int              `json:"level"`
	ParentCode        string           `json:"parent_code,omitempty"`
	IsLeaf            bool             `json:"is_leaf"`
	IsForeignCurrency bool             `json:"is_foreign_currency,omitempty"`
	Currency          string           `json:"currency,omitempty"`
	AuxTypes          []AuxType        `json:"aux_types,omitempty"`
	IsActive          bool             `json:"is_active"`
}
