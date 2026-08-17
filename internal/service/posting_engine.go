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
func (e *PostingEngine) Post(id, opUser string) error {
	v := e.st.GetVoucher(id)
	if v == nil {
		return errors.New("凭证不存在")
	}
	if v.Status == domain.StatusPosted {
		return errors.New("凭证已过账")
	}
	if v.Status != domain.StatusAudited {
		return errors.New("仅已审核凭证可过账")
	}
	if !v.IsBalanced() {
		return errors.New("借贷不平衡，无法过账")
	}
	p := e.st.GetPeriod(v.PeriodYear, v.PeriodMonth)
	if p != nil && p.Status == domain.PeriodClosed {
		return errors.New("会计期间已结账，无法过账")
	}
	for _, en := range v.Entries {
		acc := e.st.GetAccount(en.AccountCode)
		if acc == nil || !acc.IsActive || !acc.IsLeaf {
			return fmt.Errorf("科目 %s 不可用", en.AccountCode)
		}
	}

	opID := NewOpID()
	before := jsonOf(v)
	return e.st.Mutate(func() {
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
}

// Unpost 反过账
func (e *PostingEngine) Unpost(id, opUser string) error {
	v := e.st.GetVoucher(id)
	if v == nil {
		return errors.New("凭证不存在")
	}
	if v.Status != domain.StatusPosted {
		return errors.New("仅已过账凭证可反过账")
	}
	p := e.st.GetPeriod(v.PeriodYear, v.PeriodMonth)
	if p != nil && p.Status == domain.PeriodClosed {
		return errors.New("会计期间已结账，无法反过账")
	}
	opID := NewOpID()
	before := jsonOf(v)
	return e.st.Mutate(func() {
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
