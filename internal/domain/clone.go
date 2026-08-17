package domain

// 本文件提供各聚合根的深拷贝。Store 在读出时返回克隆、写入时再克隆一份落盘，
// 以保证调用方拿到或传入的对象与内存缓存互不别名——任何对返回值/入参的本地修改
// 都不会污染已保存数据，反之亦然。引用类型字段（切片、map）必须显式复制。

// Clone 返回科目的深拷贝。
func (a *Account) Clone() *Account {
	if a == nil {
		return nil
	}
	c := *a
	if a.AuxTypes != nil {
		c.AuxTypes = append([]AuxType(nil), a.AuxTypes...)
	}
	return &c
}

// Clone 返回凭证的深拷贝（含分录及其辅助核算引用）。
func (v *Voucher) Clone() *Voucher {
	if v == nil {
		return nil
	}
	c := *v
	if v.Entries != nil {
		c.Entries = make([]VoucherEntry, len(v.Entries))
		for i := range v.Entries {
			c.Entries[i] = v.Entries[i].Clone()
		}
	}
	if v.PostedAt != nil {
		t := *v.PostedAt
		c.PostedAt = &t
	}
	return &c
}

// Clone 返回分录的深拷贝（含辅助核算引用 map）。
func (e VoucherEntry) Clone() VoucherEntry {
	c := e
	if e.AuxRefs != nil {
		c.AuxRefs = make(map[AuxType]string, len(e.AuxRefs))
		for k, val := range e.AuxRefs {
			c.AuxRefs[k] = val
		}
	}
	return c
}

// Clone 返回辅助核算项的深拷贝。
func (a *AuxiliaryItem) Clone() *AuxiliaryItem {
	if a == nil {
		return nil
	}
	c := *a
	return &c
}

// Clone 返回会计期间的深拷贝（含期末汇率快照 map 与结账时间）。
func (p *Period) Clone() *Period {
	if p == nil {
		return nil
	}
	c := *p
	if p.FXRateSnapshot != nil {
		c.FXRateSnapshot = make(map[string]float64, len(p.FXRateSnapshot))
		for k, v := range p.FXRateSnapshot {
			c.FXRateSnapshot[k] = v
		}
	}
	if p.ClosedAt != nil {
		t := *p.ClosedAt
		c.ClosedAt = &t
	}
	return &c
}

// Clone 返回期间余额的深拷贝。
func (b *PeriodBalance) Clone() *PeriodBalance {
	if b == nil {
		return nil
	}
	c := *b
	return &c
}
