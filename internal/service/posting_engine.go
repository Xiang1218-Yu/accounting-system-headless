package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// PostingEngine 过账引擎：编号 + 余额更新
type PostingEngine struct {
	st    *store.Store
	audit *AuditService
}

// NewPostingEngine 构造
func NewPostingEngine(st *store.Store, audit *AuditService) *PostingEngine {
	return &PostingEngine{st: st, audit: audit}
}

// Post 过账
//
// 并发安全说明：月末批量过账时，操作员连点或两个任务同时提交同一张已审核凭证，
// 会导致同一张凭证被重复过账——余额按金额翻倍累计、审计出现重复过账记录。
// 根因有二：
//  1. 原先的状态校验（v.Status != StatusAudited）发生在 Mutate 写锁之外，两个并发
//     调用方都能在加锁前读到“已审核”，随后各自进入 Mutate 回调把金额累加两遍。
//  2. 在 Mutate 之外读取共享 *Voucher 的任何字段（状态、分录、before 快照）与另一个
//     调用方在 Mutate 内对该指针的写存在数据竞争（race detector 可证）。
//
// 这里把“读取凭证字段 + 校验 + 计算 before 快照 + 编号 + 余额更新 + 状态翻转 + 审计”
// 全部收敛进同一个 Mutate 回调，在写锁内执行：先于我们拿到锁的调用方会把状态置为
// StatusPosted，后到的调用方在锁内读到非 StatusAudited 即直接返回错误，不会触碰
// 余额、编号与审计。锁外只判断凭证是否存在（只比较指针，不读字段），给调用方清晰的
// “凭证不存在”，并避免无谓加写锁。Mutate 串行化回调执行，整体“检查-变更-落盘”原子。
func (e *PostingEngine) Post(id, opUser string) error {
	// 锁外只做不读字段的快速判定：GetVoucher 仅比较指针是否为 nil，不与 Mutate 内
	// 的字段写竞争。真正的状态/分录/期间校验都在 Mutate 回调内完成。
	if v := e.st.GetVoucher(id); v == nil {
		return errors.New("凭证不存在")
	}

	opID := NewOpID()
	var postErr error
	err := e.st.Mutate(func() {
		// 在写锁内读取凭证并做权威校验：这里的读-改-写不再可能被另一个并发过账
		// 穿插，是“只成功一次”的真正闸门。
		v := e.st.VoucherNL(id)
		if v == nil {
			postErr = errors.New("凭证不存在")
			return
		}
		if v.Status != domain.StatusAudited {
			// 已被并发调用方过账（StatusPosted）或状态被改回（StatusDraft），
			// 都不再重复过账。
			postErr = errors.New("仅已审核凭证可过账")
			return
		}
		if !v.IsBalanced() {
			postErr = errors.New("借贷不平衡，无法过账")
			return
		}
		p := e.st.PeriodNL(v.PeriodYear, v.PeriodMonth)
		if p != nil && p.Status == domain.PeriodClosed {
			postErr = errors.New("会计期间已结账，无法过账")
			return
		}
		for _, en := range v.Entries {
			acc := e.st.AccountNL(en.AccountCode)
			if acc == nil || !acc.IsActive || !acc.IsLeaf {
				postErr = fmt.Errorf("科目 %s 不可用", en.AccountCode)
				return
			}
		}

		// 在锁内计算变更前快照，确保 before/after 反映同一时刻状态。
		before := jsonOf(v)
		// 编号
		if v.Number == "" {
			seq := e.st.BumpVoucherSeq(v.PeriodKey())
			v.Number = fmt.Sprintf("记-%04d%02d-%04d", v.PeriodYear, v.PeriodMonth, seq)
		}
		// 余额更新
		for _, en := range v.Entries {
			acc := e.st.AccountNL(en.AccountCode)
			ak := AuxKey(en.AuxRefs)
			b := e.getOrCreateBalance(v.PeriodYear, v.PeriodMonth, en.AccountCode, ak, acc.Direction)
			b.PeriodDebit = b.PeriodDebit.Add(en.Debit)
			b.PeriodCredit = b.PeriodCredit.Add(en.Credit)
			recomputeClosing(b, acc.Direction)
			if acc.IsForeignCurrency {
				b.FCPeriodDebit = b.FCPeriodDebit.Add(en.DebitFC)
				b.FCPeriodCredit = b.FCPeriodCredit.Add(en.CreditFC)
				recomputeFCClosing(b, acc.Direction)
			}
			e.st.SetBalance(b)
			e.recomputeForward(v.PeriodYear, v.PeriodMonth, en.AccountCode, ak, acc.Direction)
		}
		now := time.Now()
		v.Status = domain.StatusPosted
		v.Poster = opUser
		v.PostedAt = &now
		e.st.SetVoucher(v)
		e.audit.Log(opUser, "post", "voucher", v.ID, opID, before, jsonOf(v))
	})
	// 优先返回业务校验错误，避免把落盘成功误判为过账成功。
	if postErr != nil {
		return postErr
	}
	return err
}

// Unpost 反过账
//
// 与 Post 同理：状态校验与 before 快照都在 Mutate 写锁内完成，避免并发反过账把
// 余额重复扣减，也避免锁外读共享 *Voucher 字段造成的数据竞争。
func (e *PostingEngine) Unpost(id, opUser string) error {
	// 锁外只判断凭证是否存在（只比较指针，不读字段）。
	if v := e.st.GetVoucher(id); v == nil {
		return errors.New("凭证不存在")
	}

	opID := NewOpID()
	var postErr error
	err := e.st.Mutate(func() {
		v := e.st.VoucherNL(id)
		if v == nil {
			postErr = errors.New("凭证不存在")
			return
		}
		if v.Status != domain.StatusPosted {
			// 已被并发调用方反过账回已审核，不重复扣减余额。
			postErr = errors.New("仅已过账凭证可反过账")
			return
		}
		p := e.st.PeriodNL(v.PeriodYear, v.PeriodMonth)
		if p != nil && p.Status == domain.PeriodClosed {
			postErr = errors.New("会计期间已结账，无法反过账")
			return
		}
		before := jsonOf(v)
		for _, en := range v.Entries {
			acc := e.st.AccountNL(en.AccountCode)
			ak := AuxKey(en.AuxRefs)
			b := e.st.BalanceNL(v.PeriodYear, v.PeriodMonth, en.AccountCode, ak)
			if b != nil {
				b.PeriodDebit = b.PeriodDebit.Sub(en.Debit)
				b.PeriodCredit = b.PeriodCredit.Sub(en.Credit)
				recomputeClosing(b, acc.Direction)
				if acc.IsForeignCurrency {
					b.FCPeriodDebit = b.FCPeriodDebit.Sub(en.DebitFC)
					b.FCPeriodCredit = b.FCPeriodCredit.Sub(en.CreditFC)
					recomputeFCClosing(b, acc.Direction)
				}
				e.st.SetBalance(b)
				e.recomputeForward(v.PeriodYear, v.PeriodMonth, en.AccountCode, ak, acc.Direction)
			}
		}
		v.Status = domain.StatusAudited
		v.Poster = ""
		v.PostedAt = nil
		e.st.SetVoucher(v)
		e.audit.Log(opUser, "unpost", "voucher", v.ID, opID, before, jsonOf(v))
	})
	if postErr != nil {
		return postErr
	}
	return err
}

// getOrCreateBalance 取或建余额记录，期初取自上期期末
func (e *PostingEngine) getOrCreateBalance(year, month int, code, auxKey string, direction domain.BalanceDirection) *domain.PeriodBalance {
	if b := e.st.BalanceNL(year, month, code, auxKey); b != nil {
		return b
	}
	py, pm := prevPeriod(year, month)
	pb := e.st.BalanceNL(py, pm, code, auxKey)
	b := &domain.PeriodBalance{Year: year, Month: month, AccountCode: code, AuxKey: auxKey}
	if pb != nil {
		b.OpeningDebit = pb.ClosingDebit
		b.OpeningCredit = pb.ClosingCredit
		b.FCOpeningDebit = pb.FCClosingDebit
		b.FCOpeningCredit = pb.FCClosingCredit
	}
	b.ClosingDebit = b.OpeningDebit
	b.ClosingCredit = b.OpeningCredit
	b.FCClosingDebit = b.FCOpeningDebit
	b.FCClosingCredit = b.FCOpeningCredit
	return b
}

// recomputeForward 级联重算后续期间同账户同辅助的期初/期末
func (e *PostingEngine) recomputeForward(year, month int, code, auxKey string, direction domain.BalanceDirection) {
	var chain []*domain.PeriodBalance
	for _, b := range e.st.BalancesNL() {
		if b.AccountCode == code && b.AuxKey == auxKey && (b.Year > year || (b.Year == year && b.Month > month)) {
			chain = append(chain, b)
		}
	}
	sort.Slice(chain, func(i, j int) bool {
		if chain[i].Year != chain[j].Year {
			return chain[i].Year < chain[j].Year
		}
		return chain[i].Month < chain[j].Month
	})
	prev := e.st.BalanceNL(year, month, code, auxKey)
	for _, b := range chain {
		if prev != nil {
			b.OpeningDebit = prev.ClosingDebit
			b.OpeningCredit = prev.ClosingCredit
			b.FCOpeningDebit = prev.FCClosingDebit
			b.FCOpeningCredit = prev.FCClosingCredit
		}
		recomputeClosing(b, direction)
		recomputeFCClosing(b, direction)
		e.st.SetBalance(b)
		prev = b
	}
}

func recomputeClosing(b *domain.PeriodBalance, direction domain.BalanceDirection) {
	if direction == domain.DirDebit {
		net := b.OpeningDebit.Add(b.PeriodDebit).Sub(b.PeriodCredit)
		if net.IsNegative() {
			b.ClosingDebit = decimal.Zero
			b.ClosingCredit = net.Neg()
		} else {
			b.ClosingDebit = net
			b.ClosingCredit = decimal.Zero
		}
	} else {
		net := b.OpeningCredit.Add(b.PeriodCredit).Sub(b.PeriodDebit)
		if net.IsNegative() {
			b.ClosingCredit = decimal.Zero
			b.ClosingDebit = net.Neg()
		} else {
			b.ClosingCredit = net
			b.ClosingDebit = decimal.Zero
		}
	}
}

func prevPeriod(year, month int) (int, int) {
	if month <= 1 {
		return year - 1, 12
	}
	return year, month - 1
}

func recomputeFCClosing(b *domain.PeriodBalance, direction domain.BalanceDirection) {
	if direction == domain.DirDebit {
		net := b.FCOpeningDebit.Add(b.FCPeriodDebit).Sub(b.FCPeriodCredit)
		if net.IsNegative() {
			b.FCClosingDebit = decimal.Zero
			b.FCClosingCredit = net.Neg()
		} else {
			b.FCClosingDebit = net
			b.FCClosingCredit = decimal.Zero
		}
	} else {
		net := b.FCOpeningCredit.Add(b.FCPeriodCredit).Sub(b.FCPeriodDebit)
		if net.IsNegative() {
			b.FCClosingCredit = decimal.Zero
			b.FCClosingDebit = net.Neg()
		} else {
			b.FCClosingCredit = net
			b.FCClosingDebit = decimal.Zero
		}
	}
}

// AuxKey 由辅助核算引用生成稳定组合键
func AuxKey(refs map[domain.AuxType]string) string {
	if len(refs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(refs))
	for k := range refs {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	out := ""
	for _, k := range keys {
		out += k + ":" + refs[domain.AuxType(k)] + "|"
	}
	return out
}
