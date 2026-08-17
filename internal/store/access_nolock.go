package store

import (
	"sort"

	"github.com/tog/accounting-system/internal/domain"
)

// 以下 NoLock 方法不加锁，仅供 service 在 Store.Mutate 回调内读取缓存。
// 在 Mutate 之外请使用对应的加锁读方法（GetAccount / ListAccounts 等）。

// AccountNL 无锁取科目
func (s *Store) AccountNL(code string) *domain.Account { return s.accounts[code] }

// VoucherNL 无锁取凭证
func (s *Store) VoucherNL(id string) *domain.Voucher { return s.vouchers[id] }

// AuxItemNL 无锁取辅助项
func (s *Store) AuxItemNL(id string) *domain.AuxiliaryItem { return s.auxitems[id] }

// PeriodNL 无锁取期间
func (s *Store) PeriodNL(year, month int) *domain.Period { return s.periods[formatPeriod(year, month)] }

// BalanceNL 无锁取余额
func (s *Store) BalanceNL(year, month int, code, auxKey string) *domain.PeriodBalance {
	return s.balances[balanceKey(year, month, code, auxKey)]
}

// AccountsNL 无锁全部科目（已排序）
func (s *Store) AccountsNL() []*domain.Account {
	out := make([]*domain.Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// VouchersNL 无锁全部凭证（已排序）
func (s *Store) VouchersNL() []*domain.Voucher {
	out := make([]*domain.Voucher, 0, len(s.vouchers))
	for _, v := range s.vouchers {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PeriodYear != out[j].PeriodYear {
			return out[i].PeriodYear < out[j].PeriodYear
		}
		if out[i].PeriodMonth != out[j].PeriodMonth {
			return out[i].PeriodMonth < out[j].PeriodMonth
		}
		return out[i].Number < out[j].Number
	})
	return out
}

// BalancesNL 无锁全部余额（已排序）
func (s *Store) BalancesNL() []*domain.PeriodBalance {
	out := make([]*domain.PeriodBalance, 0, len(s.balances))
	for _, b := range s.balances {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out
}

// MetaNL 无锁取元数据
func (s *Store) MetaNL() Meta { return s.meta }
