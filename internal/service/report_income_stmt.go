package service

import (
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

// IncomeStmtRow 利润表行
type IncomeStmtRow struct {
	Name    string
	Current decimal.Decimal // 本期金额
	YTD     decimal.Decimal // 本年累计
}

// IncomeStatement 利润表
type IncomeStatement struct {
	Period    string
	Rows      []IncomeStmtRow
	NetProfit decimal.Decimal
}

// BuildIncomeStatement 生成利润表（取发生额，排除结转损益凭证）
func (l *LedgerService) BuildIncomeStatement(year, month int) *IncomeStatement {
	is := &IncomeStatement{Period: formatPeriod2(year, month)}
	add := func(name string, cur, ytd decimal.Decimal) {
		is.Rows = append(is.Rows, IncomeStmtRow{Name: name, Current: cur, YTD: ytd})
	}

	// 收入（贷方净额）
	opRevenue := l.netCredit(year, month, "5001").Add(l.netCredit(year, month, "5051"))
	opRevenueYTD := l.netCreditYTD(year, month, "5001").Add(l.netCreditYTD(year, month, "5051"))
	add("一、营业收入", opRevenue, opRevenueYTD)

	opCost := l.netDebit(year, month, "5401").Add(l.netDebit(year, month, "5402"))
	opCostYTD := l.netDebitYTD(year, month, "5401").Add(l.netDebitYTD(year, month, "5402"))
	add("减：营业成本", opCost, opCostYTD)

	tax := l.netDebit(year, month, "5403")
	taxYTD := l.netDebitYTD(year, month, "5403")
	add("营业税金及附加", tax, taxYTD)

	sell := l.netDebit(year, month, "5601")
	sellYTD := l.netDebitYTD(year, month, "5601")
	add("销售费用", sell, sellYTD)

	admin := l.netDebit(year, month, "5602")
	adminYTD := l.netDebitYTD(year, month, "5602")
	add("管理费用", admin, adminYTD)

	fin := l.netDebit(year, month, "5603")
	finYTD := l.netDebitYTD(year, month, "5603")
	add("财务费用", fin, finYTD)

	invest := l.netCredit(year, month, "5101")
	investYTD := l.netCreditYTD(year, month, "5101")
	add("加：投资收益", invest, investYTD)

	opProfit := opRevenue.Sub(opCost).Sub(tax).Sub(sell).Sub(admin).Sub(fin).Add(invest)
	opProfitYTD := opRevenueYTD.Sub(opCostYTD).Sub(taxYTD).Sub(sellYTD).Sub(adminYTD).Sub(finYTD).Add(investYTD)
	add("二、营业利润", opProfit, opProfitYTD)

	nonOpIn := l.netCredit(year, month, "5301")
	nonOpInYTD := l.netCreditYTD(year, month, "5301")
	add("加：营业外收入", nonOpIn, nonOpInYTD)

	nonOpOut := l.netDebit(year, month, "5711")
	nonOpOutYTD := l.netDebitYTD(year, month, "5711")
	add("减：营业外支出", nonOpOut, nonOpOutYTD)

	totalProfit := opProfit.Add(nonOpIn).Sub(nonOpOut)
	totalProfitYTD := opProfitYTD.Add(nonOpInYTD).Sub(nonOpOutYTD)
	add("三、利润总额", totalProfit, totalProfitYTD)

	incomeTax := l.netDebit(year, month, "5801")
	incomeTaxYTD := l.netDebitYTD(year, month, "5801")
	add("减：所得税费用", incomeTax, incomeTaxYTD)

	netProfit := totalProfit.Sub(incomeTax)
	netProfitYTD := totalProfitYTD.Sub(incomeTaxYTD)
	add("四、净利润", netProfit, netProfitYTD)

	is.NetProfit = netProfit
	return is
}

// netCredit 贷方净额（期间，排除结转损益凭证）
func (l *LedgerService) netCredit(year, month int, code string) decimal.Decimal {
	d, c := l.entryTotals(year, month, month, code)
	return c.Sub(d)
}

// netDebit 借方净额
func (l *LedgerService) netDebit(year, month int, code string) decimal.Decimal {
	d, c := l.entryTotals(year, month, month, code)
	return d.Sub(c)
}

// netCreditYTD 贷方净额（本年累计）
func (l *LedgerService) netCreditYTD(year, month int, code string) decimal.Decimal {
	d, c := l.entryTotals(year, 1, month, code)
	return c.Sub(d)
}

// netDebitYTD 借方净额（本年累计）
func (l *LedgerService) netDebitYTD(year, month int, code string) decimal.Decimal {
	d, c := l.entryTotals(year, 1, month, code)
	return d.Sub(c)
}

// entryTotals 汇总某科目在 [fromMonth,toMonth] 已过账、非结转损益凭证的借贷发生额
func (l *LedgerService) entryTotals(year, fromMonth, toMonth int, code string) (decimal.Decimal, decimal.Decimal) {
	var d, c decimal.Decimal
	for _, v := range l.st.ListVouchers() {
		if v.Status != domain.StatusPosted {
			continue
		}
		if v.Type == domain.VoucherClosing {
			continue
		}
		if v.PeriodYear != year || v.PeriodMonth < fromMonth || v.PeriodMonth > toMonth {
			continue
		}
		for _, e := range v.Entries {
			if e.AccountCode == code {
				d = d.Add(e.Debit)
				c = c.Add(e.Credit)
			}
		}
	}
	return d, c
}
