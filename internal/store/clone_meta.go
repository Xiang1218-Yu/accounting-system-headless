package store

func cloneMeta(src Meta) Meta {
	dst := src
	if src.VoucherSeq == nil {
		return dst
	}
	dst.VoucherSeq = make(map[string]int, len(src.VoucherSeq))
	for key, value := range src.VoucherSeq {
		dst.VoucherSeq[key] = value
	}
	return dst
}

func cloneRates(src map[string]float64) map[string]float64 {
	if src == nil {
		return nil
	}
	dst := make(map[string]float64, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
