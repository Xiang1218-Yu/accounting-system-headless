package service

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

// VerifyResult 一致性自检结果
type VerifyResult struct {
	OK     bool
	Checks []string
	Errors []string
}

// Verify 全量一致性自检
func (l *LedgerService) Verify(year, month int) *VerifyResult {
	r := &VerifyResult{OK: true}

	// 1. 所有已过账凭证借贷平衡
	for _, v := range l.st.ListVouchers() {
		if v.Status != domain.StatusPosted {
			continue
		}
		if !v.IsBalanced() {
			r.Errors = append(r.Errors, fmt.Sprintf("凭证 %s 借贷不平衡", v.Number))
			r.OK = false
		}
	}
	r.Checks = append(r.Checks, "已过账凭证借贷平衡校验")

	// 2. 余额表试算平衡：期末借方合计 == 期末贷方合计
	rows := l.TrialBalance(year, month)
	_, _, _, _, cd, cc := BalanceTotals(rows)
	if !cd.Equal(cc) {
		r.Errors = append(r.Errors, fmt.Sprintf("余额表期末借贷不平：借 %s / 贷 %s", cd.String(), cc.String()))
		r.OK = false
	}
	r.Checks = append(r.Checks, "余额表试算平衡")

	// 3. 资产负债表平衡
	bs := l.BuildBalanceSheet(year, month)
	if !bs.Balanced {
		r.Errors = append(r.Errors, fmt.Sprintf("资产负债表不平：资产 %s / 负债+权益 %s", bs.AssetTotal.String(), bs.LiabTotal.Add(bs.EquityTotal).String()))
		r.OK = false
	}
	r.Checks = append(r.Checks, "资产负债表平衡（资产=负债+权益）")

	// 4. 现金流量表对账
	cf := l.BuildCashFlow(year, month)
	if !cf.Reconciled {
		r.Errors = append(r.Errors, fmt.Sprintf("现金流量对账不平：期初+净增加 %s / 期末 %s", cf.CashBegin.Add(cf.NetIncrease).String(), cf.CashEnd.String()))
		r.OK = false
	}
	r.Checks = append(r.Checks, "现金流量表对账（期初+净增加=期末）")

	return r
}

// SumDecimal 求和辅助
func SumDecimal(ds []decimal.Decimal) decimal.Decimal {
	s := decimal.Zero
	for _, d := range ds {
		s = s.Add(d)
	}
	return s
}
