package store

import "github.com/tog/accounting-system/internal/domain"

// 以下导出方法仅供 service 在 Store.Mutate 回调内调用，
// 以修改内存缓存并标记脏数据。切勿在 Mutate 之外调用。

// SetAccount 写入科目
func (s *Store) SetAccount(a *domain.Account) { s.setAccount(a) }

// DeleteAccount 删除科目
func (s *Store) DeleteAccount(code string) bool { return s.deleteAccount(code) }

// SetVoucher 写入凭证
func (s *Store) SetVoucher(v *domain.Voucher) { s.setVoucher(v) }

// DeleteVoucher 删除凭证
func (s *Store) DeleteVoucher(id string) bool { return s.deleteVoucher(id) }

// SetAuxItem 写入辅助项
func (s *Store) SetAuxItem(a *domain.AuxiliaryItem) { s.setAuxItem(a) }

// DeleteAuxItem 删除辅助项
func (s *Store) DeleteAuxItem(id string) bool { return s.deleteAuxItem(id) }

// SetPeriod 写入期间
func (s *Store) SetPeriod(p *domain.Period) { s.setPeriod(p) }

// SetBalance 写入余额
func (s *Store) SetBalance(b *domain.PeriodBalance) { s.setBalance(b) }

// AppendAudit 追加审计日志
func (s *Store) AppendAudit(l domain.AuditLog) { s.appendAudit(l) }

// SetMeta 写入元数据
func (s *Store) SetMeta(m Meta) { s.setMeta(m) }

// BumpVoucherSeq 递增并返回期间凭证序号
func (s *Store) BumpVoucherSeq(periodKey string) int { return s.bumpVoucherSeq(periodKey) }
