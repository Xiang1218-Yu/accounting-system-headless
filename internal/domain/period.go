package domain

import (
	"fmt"
	"time"
)

// Period 会计期间
type Period struct {
	Year           int                `json:"year"`
	Month          int                `json:"month"`
	Status         PeriodStatus       `json:"status"`
	ClosedAt       *time.Time         `json:"closed_at,omitempty"`
	ClosedBy       string             `json:"closed_by,omitempty"`
	FXRateSnapshot map[string]float64 `json:"fx_rate_snapshot,omitempty"`
	Note           string             `json:"note,omitempty"`
}

// Key 期间键 YYYY-MM
func (p Period) Key() string {
	return periodKey(p.Year, p.Month)
}

func periodKey(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}
