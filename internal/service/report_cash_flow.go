package service

import (
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

// CFRow 现金流量表行
type CFRow struct {
	Name     string
	Amount   decimal.Decimal
	Subtotal bool
}

// CashFlowStatement 现金流量表
type CashFlowStatement struct {
	Period      string
	Rows        []CFRow
	NetIncrease decimal.Decimal
	CashBegin   decimal.Decimal
	CashEnd     decimal.Decimal
	Reconciled  bool // 净增加额+期初 == 期末
}

// BuildCashFlow 生成现金流量表（直接法，仅普通凭证；调汇单列汇率影响）
func (l *LedgerService) BuildCashFlow(year, month int) *CashFlowStatement {
	cf := &CashFlowStatement{Period: formatPeriod2(year, month)}

	var opSales, opOtherIn, opPurchase, opSalary, opTax, opOtherOut decimal.Decimal
	var invRecover, invIncome, invBuy decimal.Decimal
	var finBorrowIn, finRepayOut, finDividend decimal.Decimal

	for _, v := range l.st.ListVouchers() {
		if v.Status != domain.StatusPosted || v.Type != domain.VoucherNormal {
			continue
		}
		if v.PeriodYear != year || v.PeriodMonth != month {
			continue
		}
		// 直接法只反映涉及现金的凭证；纯非现金凭证（如挂应付账款购置资产）
		// 没有现金腿，不应产生任何现金流量，避免现金流出虚增。
		if !voucherHasCashLeg(v) {
			continue
		}
		for _, e := range v.Entries {
			if isCashAccount(e.AccountCode) {
				continue // 现金腿跳过
			}
			switch classifyCF(e.AccountCode) {
			case cfGrpOpSales:
				opSales = opSales.Add(e.Credit.Sub(e.Debit))
			case cfGrpOpOtherIn:
				opOtherIn = opOtherIn.Add(e.Credit.Sub(e.Debit))
			case cfGrpOpPurchase:
				opPurchase = opPurchase.Add(e.Debit.Sub(e.Credit))
			case cfGrpOpSalary:
				opSalary = opSalary.Add(e.Debit.Sub(e.Credit))
			case cfGrpOpTax:
				opTax = opTax.Add(e.Debit.Sub(e.Credit))
			case cfGrpOpOtherOut:
				opOtherOut = opOtherOut.Add(e.Debit.Sub(e.Credit))
			case cfGrpInvRecover:
				invRecover = invRecover.Add(e.Credit.Sub(e.Debit))
			case cfGrpInvIncome:
				invIncome = invIncome.Add(e.Credit.Sub(e.Debit))
			case cfGrpInvBuy:
				invBuy = invBuy.Add(e.Debit.Sub(e.Credit))
			case cfGrpFinBorrow:
				finBorrowIn = finBorrowIn.Add(e.Credit)
				finRepayOut = finRepayOut.Add(e.Debit)
			case cfGrpFinDividend:
				finDividend = finDividend.Add(e.Debit.Sub(e.Credit))
			case cfGrpOther:
				// 其他：贷方归入其他流入，借方归入其他流出
				if e.Credit.GreaterThan(decimal.Zero) {
					opOtherIn = opOtherIn.Add(e.Credit)
				} else {
					opOtherOut = opOtherOut.Add(e.Debit)
				}
			}
		}
	}

	// 汇率变动对现金的影响（期末调汇凭证中现金科目的本币变动）
	var fxEffect decimal.Decimal
	for _, v := range l.st.ListVouchers() {
		if v.Status != domain.StatusPosted || v.Type != domain.VoucherFX {
			continue
		}
		if v.PeriodYear != year || v.PeriodMonth != month {
			continue
		}
		for _, e := range v.Entries {
			if isCashAccount(e.AccountCode) {
				fxEffect = fxEffect.Add(e.Debit.Sub(e.Credit))
			}
		}
	}

	opInTotal := opSales.Add(opOtherIn)
	opOutTotal := opPurchase.Add(opSalary).Add(opTax).Add(opOtherOut)
	opNet := opInTotal.Sub(opOutTotal)

	invInTotal := invRecover.Add(invIncome)
	invOutTotal := invBuy
	invNet := invInTotal.Sub(invOutTotal)

	finInTotal := finBorrowIn
	finOutTotal := finRepayOut.Add(finDividend)
	finNet := finInTotal.Sub(finOutTotal)

	add := func(name string, amt decimal.Decimal, sub bool) {
		cf.Rows = append(cf.Rows, CFRow{Name: name, Amount: amt, Subtotal: sub})
	}

	add("一、经营活动产生的现金流量", decimal.Zero, false)
	add("销售商品、提供劳务收到的现金", opSales, false)
	add("收到其他与经营活动有关的现金", opOtherIn, false)
	add("现金流入小计", opInTotal, true)
	add("购买商品、接受劳务支付的现金", opPurchase, false)
	add("支付给职工以及为职工支付的现金", opSalary, false)
	add("支付的各项税费", opTax, false)
	add("支付其他与经营活动有关的现金", opOtherOut, false)
	add("现金流出小计", opOutTotal, true)
	add("经营活动产生的现金流量净额", opNet, true)

	add("二、投资活动产生的现金流量", decimal.Zero, false)
	add("收回投资收到的现金", invRecover, false)
	add("取得投资收益收到的现金", invIncome, false)
	add("现金流入小计", invInTotal, true)
	add("购建固定资产、无形资产和其他长期资产支付的现金", invBuy, false)
	add("现金流出小计", invOutTotal, true)
	add("投资活动产生的现金流量净额", invNet, true)

	add("三、筹资活动产生的现金流量", decimal.Zero, false)
	add("取得借款收到的现金", finBorrowIn, false)
	add("现金流入小计", finInTotal, true)
	add("偿还债务支付的现金", finRepayOut, false)
	add("分配股利、利润或偿付利息支付的现金", finDividend, false)
	add("现金流出小计", finOutTotal, true)
	add("筹资活动产生的现金流量净额", finNet, true)

	add("四、汇率变动对现金的影响", fxEffect, false)
	cf.NetIncrease = opNet.Add(invNet).Add(finNet).Add(fxEffect)
	add("五、现金及现金等价物净增加额", cf.NetIncrease, true)

	cf.CashBegin = l.cashBalanceAsOf(year, month, true)
	cf.CashEnd = l.cashBalanceAsOf(year, month, false)
	add("加：期初现金余额", cf.CashBegin, false)
	add("六、期末现金余额", cf.CashEnd, true)

	cf.Reconciled = cf.CashBegin.Add(cf.NetIncrease).Equal(cf.CashEnd)
	return cf
}

// cashBalanceAsOf 现金科目期初/期末净额
func (l *LedgerService) cashBalanceAsOf(year, month int, opening bool) decimal.Decimal {
	total := decimal.Zero
	for _, code := range cashAccounts {
		if opening {
			od, oc := l.openingAsOf(year, month, code)
			total = total.Add(od.Sub(oc))
		} else {
			cd, cc := l.closingAsOf(year, month, code)
			total = total.Add(cd.Sub(cc))
		}
	}
	return total
}
