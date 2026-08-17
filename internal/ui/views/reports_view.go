//go:build !linux

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tog/accounting-system/internal/service"
	"github.com/tog/accounting-system/internal/ui/widgets"
)

// BuildReportsView 三大报表视图
func BuildReportsView(a App) fyne.CanvasObject {
	year, month := a.CurYear(), a.CurMonth()

	bs := a.Svc().Ledger.BuildBalanceSheet(year, month)
	bsView := buildBSTable(bs)
	bsBtn := widget.NewButton("导出 Excel", func() {
		saveDialog(a, "资产负债表.xlsx", func(p string) error { return a.Svc().Excel.ExportBalanceSheet(p, bs) })
	})

	is := a.Svc().Ledger.BuildIncomeStatement(year, month)
	isView := buildISTable(is)
	isBtn := widget.NewButton("导出 Excel", func() {
		saveDialog(a, "利润表.xlsx", func(p string) error { return a.Svc().Excel.ExportIncomeStatement(p, is) })
	})

	cf := a.Svc().Ledger.BuildCashFlow(year, month)
	cfView := buildCFTable(cf)
	cfBtn := widget.NewButton("导出 Excel", func() {
		saveDialog(a, "现金流量表.xlsx", func(p string) error { return a.Svc().Excel.ExportCashFlow(p, cf) })
	})

	return container.NewAppTabs(
		container.NewTabItem("资产负债表", container.NewBorder(container.NewHBox(bsBtn), nil, nil, nil, bsView)),
		container.NewTabItem("利润表", container.NewBorder(container.NewHBox(isBtn), nil, nil, nil, isView)),
		container.NewTabItem("现金流量表", container.NewBorder(container.NewHBox(cfBtn), nil, nil, nil, cfView)),
	)
}

func buildBSTable(bs *service.BalanceSheet) fyne.CanvasObject {
	headers := []string{"资产", "金额", "负债与权益", "金额"}
	rows := [][]string{}
	assetRows := bs.AssetRows
	liabRows := bs.LiabRows
	eqRows := bs.EquityRows
	maxLen := len(assetRows)
	if l := len(liabRows) + len(eqRows) + 1; l > maxLen {
		maxLen = l
	}
	for i := 0; i < maxLen; i++ {
		row := []string{"", "", "", ""}
		if i < len(assetRows) {
			row[0] = assetRows[i].Name
			row[1] = decStr(assetRows[i].Amount)
		}
		if i == len(assetRows) {
			row[0] = "资产总计"
			row[1] = decStr(bs.AssetTotal)
		}
		if i < len(liabRows) {
			row[2] = liabRows[i].Name
			row[3] = decStr(liabRows[i].Amount)
		} else if j := i - len(liabRows); j >= 0 && j < len(eqRows) {
			row[2] = eqRows[j].Name
			row[3] = decStr(eqRows[j].Amount)
		} else if i == len(liabRows)+len(eqRows) {
			row[2] = "负债+权益总计"
			row[3] = decStr(bs.LiabTotal.Add(bs.EquityTotal))
		}
		rows = append(rows, row)
	}
	status := fmt.Sprintf("（%s）", boolStr(bs.Balanced))
	if bs.Balanced {
		status = "（平衡 ✓）"
	} else {
		status = "（不平 ✗）"
	}
	return container.NewBorder(widget.NewLabel("资产负债表 "+bs.Period+status), nil, nil, nil,
		widgets.NewGridTable(headers, rows, []float32{180, 130, 180, 130}))
}

func buildISTable(is *service.IncomeStatement) fyne.CanvasObject {
	headers := []string{"项目", "本期金额", "本年累计"}
	rows := make([][]string, 0, len(is.Rows))
	for _, r := range is.Rows {
		rows = append(rows, []string{r.Name, decStr(r.Current), decStr(r.YTD)})
	}
	return container.NewBorder(widget.NewLabel("利润表 "+is.Period), nil, nil, nil,
		widgets.NewGridTable(headers, rows, []float32{260, 140, 140}))
}

func buildCFTable(cf *service.CashFlowStatement) fyne.CanvasObject {
	headers := []string{"项目", "金额"}
	rows := make([][]string, 0, len(cf.Rows))
	for _, r := range cf.Rows {
		name := r.Name
		if r.Subtotal {
			name = "  " + r.Name
		}
		rows = append(rows, []string{name, decStr(r.Amount)})
	}
	status := "（对账平衡 ✓）"
	if !cf.Reconciled {
		status = "（对账不平 ✗）"
	}
	return container.NewBorder(widget.NewLabel("现金流量表 "+cf.Period+status), nil, nil, nil,
		widgets.NewGridTable(headers, rows, []float32{360, 150}))
}
