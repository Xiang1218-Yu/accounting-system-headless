package store

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/tog/accounting-system/internal/domain"
)

// entityDirty 标记哪些聚合需要落盘
type entityDirty string

const (
	dirtyAccounts entityDirty = "accounts"
	dirtyVouchers entityDirty = "vouchers"
	dirtyAux      entityDirty = "auxitems"
	dirtyPeriods  entityDirty = "periods"
	dirtyBalances entityDirty = "balances"
	dirtyAudit    entityDirty = "audit"
	dirtyMeta     entityDirty = "meta"
)

// Meta 应用元数据
type Meta struct {
	Version      string         `json:"version"`
	CurrentYear  int            `json:"current_year"`
	CurrentMonth int            `json:"current_month"`
	VoucherSeq   map[string]int `json:"voucher_seq"` // 期间键 -> 已用最大序号
}

// Store JSON 本地存储：内存缓存 + 原子落盘
type Store struct {
	dir string
	mu  sync.RWMutex

	accounts map[string]*domain.Account
	vouchers map[string]*domain.Voucher
	auxitems map[string]*domain.AuxiliaryItem
	periods  map[string]*domain.Period
	balances map[string]*domain.PeriodBalance
	audit    []domain.AuditLog
	meta     Meta

	dirty map[entityDirty]bool
}

// New 创建存储并加载数据
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		dir:      dir,
		accounts: map[string]*domain.Account{},
		vouchers: map[string]*domain.Voucher{},
		auxitems: map[string]*domain.AuxiliaryItem{},
		periods:  map[string]*domain.Period{},
		balances: map[string]*domain.PeriodBalance{},
		audit:    []domain.AuditLog{},
		meta:     Meta{Version: "1.0.0", VoucherSeq: map[string]int{}},
		dirty:    map[entityDirty]bool{},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	if err := loadJSON(s.path("accounts.json"), &s.accounts); err != nil {
		return err
	}
	if err := loadJSON(s.path("vouchers.json"), &s.vouchers); err != nil {
		return err
	}
	if err := loadJSON(s.path("auxitems.json"), &s.auxitems); err != nil {
		return err
	}
	if err := loadJSON(s.path("periods.json"), &s.periods); err != nil {
		return err
	}
	if err := loadJSON(s.path("balances.json"), &s.balances); err != nil {
		return err
	}
	if err := loadJSON(s.path("audit.json"), &s.audit); err != nil {
		return err
	}
	if err := loadJSON(s.path("meta.json"), &s.meta); err != nil {
		return err
	}
	if s.meta.VoucherSeq == nil {
		s.meta.VoucherSeq = map[string]int{}
	}
	if s.meta.CurrentYear == 0 {
		s.meta.CurrentYear = defaultYear()
		s.meta.CurrentMonth = defaultMonth()
	}
	return nil
}

func (s *Store) path(name string) string { return filepath.Join(s.dir, name) }

// Dir 数据目录
func (s *Store) Dir() string { return s.dir }

// Mutate 在写锁内执行业务变更并一次性落盘所有脏聚合
func (s *Store) Mutate(fn func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn()
	for e := range s.dirty {
		if err := s.saveEntity(e); err != nil {
			return err
		}
	}
	s.dirty = map[entityDirty]bool{}
	return nil
}

func (s *Store) saveEntity(e entityDirty) error {
	switch e {
	case dirtyAccounts:
		return saveJSON(s.path("accounts.json"), s.accounts)
	case dirtyVouchers:
		return saveJSON(s.path("vouchers.json"), s.vouchers)
	case dirtyAux:
		return saveJSON(s.path("auxitems.json"), s.auxitems)
	case dirtyPeriods:
		return saveJSON(s.path("periods.json"), s.periods)
	case dirtyBalances:
		return saveJSON(s.path("balances.json"), s.balances)
	case dirtyAudit:
		return saveJSON(s.path("audit.json"), s.audit)
	case dirtyMeta:
		return saveJSON(s.path("meta.json"), s.meta)
	}
	return nil
}

func (s *Store) markDirty(e ...entityDirty) {
	for _, d := range e {
		s.dirty[d] = true
	}
}

// ----- 内部写方法（须在 Mutate 内调用，不再加锁） -----

func (s *Store) setAccount(a *domain.Account) {
	// 克隆后落盘，避免调用方继续修改入参时污染内存缓存
	s.accounts[a.Code] = a.Clone()
	s.markDirty(dirtyAccounts)
}

func (s *Store) deleteAccount(code string) bool {
	if _, ok := s.accounts[code]; !ok {
		return false
	}
	delete(s.accounts, code)
	s.markDirty(dirtyAccounts)
	return true
}

func (s *Store) setVoucher(v *domain.Voucher) {
	s.vouchers[v.ID] = v.Clone()
	s.markDirty(dirtyVouchers)
}

func (s *Store) deleteVoucher(id string) bool {
	if _, ok := s.vouchers[id]; !ok {
		return false
	}
	delete(s.vouchers, id)
	s.markDirty(dirtyVouchers)
	return true
}

func (s *Store) setAuxItem(a *domain.AuxiliaryItem) {
	s.auxitems[a.ID] = a.Clone()
	s.markDirty(dirtyAux)
}

func (s *Store) deleteAuxItem(id string) bool {
	if _, ok := s.auxitems[id]; !ok {
		return false
	}
	delete(s.auxitems, id)
	s.markDirty(dirtyAux)
	return true
}

func (s *Store) setPeriod(p *domain.Period) {
	s.periods[p.Key()] = p.Clone()
	s.markDirty(dirtyPeriods)
}

func (s *Store) setBalance(b *domain.PeriodBalance) {
	s.balances[b.Key()] = b.Clone()
	s.markDirty(dirtyBalances)
}

func (s *Store) appendAudit(l domain.AuditLog) {
	s.audit = append(s.audit, l)
	s.markDirty(dirtyAudit)
}

func (s *Store) setMeta(m Meta) {
	// 深拷贝 VoucherSeq，避免调用方修改其 map 污染缓存
	s.meta = m
	if m.VoucherSeq != nil {
		s.meta.VoucherSeq = make(map[string]int, len(m.VoucherSeq))
		for k, v := range m.VoucherSeq {
			s.meta.VoucherSeq[k] = v
		}
	}
	s.markDirty(dirtyMeta)
}

func (s *Store) bumpVoucherSeq(periodKey string) int {
	seq := s.meta.VoucherSeq[periodKey] + 1
	s.meta.VoucherSeq[periodKey] = seq
	s.markDirty(dirtyMeta)
	return seq
}

// ----- 只读方法（RLock） -----

// ListAccounts 按编码排序的科目。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) ListAccounts() []*domain.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		out = append(out, a.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// GetAccount 取科目。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) GetAccount(code string) *domain.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accounts[code].Clone()
}

// ListVouchers 按日期/编号排序的凭证。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) ListVouchers() []*domain.Voucher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.Voucher, 0, len(s.vouchers))
	for _, v := range s.vouchers {
		out = append(out, v.Clone())
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

// GetVoucher 取凭证。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) GetVoucher(id string) *domain.Voucher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vouchers[id].Clone()
}

// ListAuxItems 辅助项（可按类型过滤）。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) ListAuxItems(t domain.AuxType) []*domain.AuxiliaryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.AuxiliaryItem, 0, len(s.auxitems))
	for _, a := range s.auxitems {
		if t != "" && a.Type != t {
			continue
		}
		out = append(out, a.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return out[i].Code < out[j].Code
	})
	return out
}

// GetAuxItem 取辅助项。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) GetAuxItem(id string) *domain.AuxiliaryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.auxitems[id].Clone()
}

// ListPeriods 期间列表。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) ListPeriods() []*domain.Period {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.Period, 0, len(s.periods))
	for _, p := range s.periods {
		out = append(out, p.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out
}

// GetPeriod 取期间。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) GetPeriod(year, month int) *domain.Period {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.periods[periodKeyStr(year, month)].Clone()
}

// ListBalances 余额列表。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) ListBalances() []*domain.PeriodBalance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.PeriodBalance, 0, len(s.balances))
	for _, b := range s.balances {
		out = append(out, b.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out
}

// GetBalance 取余额。返回克隆，调用方修改不会污染已保存数据。
func (s *Store) GetBalance(year, month int, accountCode, auxKey string) *domain.PeriodBalance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := balanceKey(year, month, accountCode, auxKey)
	return s.balances[key].Clone()
}

// ListAudit 审计日志（倒序）
func (s *Store) ListAudit() []domain.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AuditLog, len(s.audit))
	copy(out, s.audit)
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp.After(out[j].Timestamp) })
	return out
}

// Meta 取元数据副本（VoucherSeq 深拷贝，调用方修改不会污染已保存数据）
func (s *Store) Meta() Meta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := s.meta
	if m.VoucherSeq != nil {
		m.VoucherSeq = make(map[string]int, len(s.meta.VoucherSeq))
		for k, v := range s.meta.VoucherSeq {
			m.VoucherSeq[k] = v
		}
	} else {
		m.VoucherSeq = map[string]int{}
	}
	return m
}

func periodKeyStr(year, month int) string {
	return formatPeriod(year, month)
}

func balanceKey(year, month int, accountCode, auxKey string) string {
	return formatPeriod(year, month) + "|" + accountCode + "|" + auxKey
}
