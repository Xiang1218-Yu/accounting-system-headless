package service

import (
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

// BSRow 资产负债表行
type BSRow struct {
	Name   string
	Amount decimal.Decimal
	Memo   bool
}

// BalanceSheet 资产负债表
type BalanceSheet struct {
	Period      string
	AssetRows   []BSRow
	LiabRows    []BSRow
	EquityRows  []BSRow
	AssetTotal  decimal.Decimal
	LiabTotal   decimal.Decimal
	EquityTotal decimal.Decimal
	Balanced    bool // 资产总计 == 负债+权益
}

// BuildBalanceSheet 生成资产负债表
func (l *LedgerService) BuildBalanceSheet(year, month int) *BalanceSheet {
	bs := &BalanceSheet{Period: fmtPeriod(year, month)}
	for _, it := range balanceSheetAssets() {
		amt := l.bsItemAmount(year, month, it)
		bs.AssetRows = append(bs.AssetRows, BSRow{Name: it.Name, Amount: amt, Memo: it.Memo})
		if !it.Memo {
			bs.AssetTotal = bs.AssetTotal.Add(amt)
		}
	}
	liab, equity := balanceSheetLiabEquity()
	for _, it := range liab {
		amt := l.bsItemAmount(year, month, it)
		bs.LiabRows = append(bs.LiabRows, BSRow{Name: it.Name, Amount: amt})
		bs.LiabTotal = bs.LiabTotal.Add(amt)
	}
	for _, it := range equity {
		amt := l.bsItemAmount(year, month, it)
		bs.EquityRows = append(bs.EquityRows, BSRow{Name: it.Name, Amount: amt})
		bs.EquityTotal = bs.EquityTotal.Add(amt)
	}
	bs.Balanced = bs.AssetTotal.Equal(bs.LiabTotal.Add(bs.EquityTotal))
	return bs
}

// bsItemAmount 计算行项目金额
func (l *LedgerService) bsItemAmount(year, month int, it bsItem) decimal.Decimal {
	amt := decimal.Zero
	for _, c := range it.AddCodes {
		amt = amt.Add(l.netFor(year, month, c))
	}
	for _, c := range it.SubCodes {
		amt = amt.Sub(l.netFor(year, month, c))
	}
	return amt
}

// netFor 某科目截至某期间按方向的净额（资产取借方净，负债权益取贷方净）
func (l *LedgerService) netFor(year, month int, code string) decimal.Decimal {
	cd, cc := l.closingAsOf(year, month, code)
	acc := l.st.GetAccount(code)
	if acc != nil && acc.Direction == domain.DirCredit {
		return cc.Sub(cd)
	}
	return cd.Sub(cc)
}

func fmtPeriod(year, month int) string {
	return formatPeriod2(year, month)
}
