package store

import "github.com/tog/accounting-system/internal/domain"

func cloneAccount(src *domain.Account) *domain.Account {
	if src == nil {
		return nil
	}
	dst := *src
	dst.AuxTypes = append([]domain.AuxType(nil), src.AuxTypes...)
	return &dst
}

func cloneAuxItem(src *domain.AuxiliaryItem) *domain.AuxiliaryItem {
	if src == nil {
		return nil
	}
	dst := *src
	return &dst
}
