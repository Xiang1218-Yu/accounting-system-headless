package service

// bsItem 资产负债表行项目
type bsItem struct {
	Name     string
	AddCodes []string
	SubCodes []string
	Memo     bool // 备注行，不计入合计
}

// balanceSheetAssets 资产行项目
func balanceSheetAssets() []bsItem {
	return []bsItem{
		{Name: "货币资金", AddCodes: []string{"1001", "1002", "1012"}},
		{Name: "短期投资", AddCodes: []string{"1101"}},
		{Name: "应收票据", AddCodes: []string{"1121"}},
		{Name: "应收账款", AddCodes: []string{"1122"}, SubCodes: []string{"1231"}},
		{Name: "预付账款", AddCodes: []string{"1123"}},
		{Name: "应收股利", AddCodes: []string{"1131"}},
		{Name: "应收利息", AddCodes: []string{"1132"}},
		{Name: "其他应收款", AddCodes: []string{"1221"}},
		{Name: "存货", AddCodes: []string{"1401", "1402", "1403", "1404", "1405", "1411", "4001"}, SubCodes: []string{"1407"}},
		{Name: "长期债券投资", AddCodes: []string{"1501"}},
		{Name: "长期股权投资", AddCodes: []string{"1511"}},
		{Name: "固定资产", AddCodes: []string{"1601"}, SubCodes: []string{"1602"}},
		{Name: "其中：固定资产原价", AddCodes: []string{"1601"}, Memo: true},
		{Name: "减：累计折旧", AddCodes: []string{"1602"}, Memo: true},
		{Name: "在建工程", AddCodes: []string{"1604"}},
		{Name: "工程物资", AddCodes: []string{"1605"}},
		{Name: "固定资产清理", AddCodes: []string{"1606"}},
		{Name: "无形资产", AddCodes: []string{"1701"}, SubCodes: []string{"1702"}},
		{Name: "长期待摊费用", AddCodes: []string{"1801"}},
		{Name: "待处理财产损溢", AddCodes: []string{"1901"}},
	}
}

// balanceSheetLiabEquity 负债与权益行项目
func balanceSheetLiabEquity() ([]bsItem, []bsItem) {
	liab := []bsItem{
		{Name: "短期借款", AddCodes: []string{"2001"}},
		{Name: "应付票据", AddCodes: []string{"2201"}},
		{Name: "应付账款", AddCodes: []string{"2202"}},
		{Name: "预收账款", AddCodes: []string{"2203"}},
		{Name: "应付职工薪酬", AddCodes: []string{"2211"}},
		{Name: "应交税费", AddCodes: []string{"2221"}},
		{Name: "应付利息", AddCodes: []string{"2231"}},
		{Name: "应付利润", AddCodes: []string{"2232"}},
		{Name: "其他应付款", AddCodes: []string{"2241"}},
		{Name: "长期借款", AddCodes: []string{"2501"}},
		{Name: "长期应付款", AddCodes: []string{"2502"}},
	}
	equity := []bsItem{
		{Name: "实收资本", AddCodes: []string{"3001"}},
		{Name: "资本公积", AddCodes: []string{"3002"}},
		{Name: "盈余公积", AddCodes: []string{"3101"}},
		{Name: "未分配利润", AddCodes: []string{"3103", "3104"}},
	}
	return liab, equity
}

// cashAccounts 现金及现金等价物科目
var cashAccounts = []string{"1001", "1002", "1012"}

// cfGroup 现金流量分组（direction 决定流入/流出；finBorrow 双向）
type cfGroup int

const (
	cfGrpNone        cfGroup = iota
	cfGrpOpSales             // 销售商品提供劳务收到的现金（流入）
	cfGrpOpOtherIn           // 收到其他与经营活动有关的现金（流入）
	cfGrpOpPurchase          // 购买商品接受劳务支付的现金（流出）
	cfGrpOpSalary            // 支付给职工的现金（流出）
	cfGrpOpTax               // 支付的各项税费（流出）
	cfGrpOpOtherOut          // 支付其他与经营活动有关的现金（流出）
	cfGrpInvRecover          // 收回投资收到的现金（流入）
	cfGrpInvIncome           // 取得投资收益收到的现金（流入）
	cfGrpInvBuy              // 购建长期资产支付的现金（流出）
	cfGrpFinBorrow           // 借款（流入=取得，流出=偿还）
	cfGrpFinDividend         // 分配股利利润偿付利息支付的现金（流出）
	cfGrpOther               // 其他经营活动
)

// classifyCF 按对方科目归类现金流量分组
func classifyCF(code string) cfGroup {
	switch code {
	case "5001", "5051", "1122", "2203", "1131", "1132":
		return cfGrpOpSales
	case "1221", "5301":
		return cfGrpOpOtherIn
	case "1401", "1402", "1403", "1404", "1405", "1411", "2202", "1123", "4001", "4101":
		return cfGrpOpPurchase
	case "2211":
		return cfGrpOpSalary
	case "2221", "5403":
		return cfGrpOpTax
	case "5601", "5602", "5603", "5711", "2241":
		return cfGrpOpOtherOut
	case "1101", "1501", "1511":
		return cfGrpInvRecover
	case "5101":
		return cfGrpInvIncome
	case "1601", "1604", "1605", "1701":
		return cfGrpInvBuy
	case "2001", "2501":
		return cfGrpFinBorrow
	case "2232", "2231":
		return cfGrpFinDividend
	}
	return cfGrpOther
}

// isCashAccount 是否现金科目
func isCashAccount(code string) bool {
	for _, c := range cashAccounts {
		if c == code {
			return true
		}
	}
	return false
}
