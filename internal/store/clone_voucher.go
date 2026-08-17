package store

import "github.com/tog/accounting-system/internal/domain"

func cloneVoucher(src *domain.Voucher) *domain.Voucher {
	if src == nil {
		return nil
	}
	dst := *src
	if src.PostedAt != nil {
		postedAt := *src.PostedAt
		dst.PostedAt = &postedAt
	}
	dst.Entries = make([]domain.VoucherEntry, len(src.Entries))
	copy(dst.Entries, src.Entries)
	for i := range dst.Entries {
		dst.Entries[i].AuxRefs = cloneAuxRefs(src.Entries[i].AuxRefs)
	}
	return &dst
}

func cloneAuxRefs(src map[domain.AuxType]string) map[domain.AuxType]string {
	if src == nil {
		return nil
	}
	dst := make(map[domain.AuxType]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
