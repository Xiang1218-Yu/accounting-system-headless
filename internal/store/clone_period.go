package store

import "github.com/tog/accounting-system/internal/domain"

func clonePeriod(src *domain.Period) *domain.Period {
	if src == nil {
		return nil
	}
	dst := *src
	if src.ClosedAt != nil {
		closedAt := *src.ClosedAt
		dst.ClosedAt = &closedAt
	}
	dst.FXRateSnapshot = cloneRates(src.FXRateSnapshot)
	return &dst
}

func cloneBalance(src *domain.PeriodBalance) *domain.PeriodBalance {
	if src == nil {
		return nil
	}
	dst := *src
	return &dst
}
