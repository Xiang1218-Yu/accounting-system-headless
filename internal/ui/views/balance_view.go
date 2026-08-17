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

// BuildBalanceView 余额表视图
func BuildBalanceView(a App) fyne.CanvasObject {
	year, month := a.CurYear(), a.CurMonth()
	rows := a.Svc().Ledger.TrialBalance(year, month)
	od, oc, pd, pc, cd, cc := service.BalanceTotals(rows)

	headers := []string{"编码", "名称", "期初借", "期初贷", "本期借", "本期贷", "期末借", "期末贷"}
	grid := make([][]string, 0, len(rows)+1)
	for _, r := range rows {
		grid = append(grid, []string{
			r.AccountCode, r.AccountName,
			decStr(r.OpeningDebit), decStr(r.OpeningCredit),
			decStr(r.PeriodDebit), decStr(r.PeriodCredit),
			decStr(r.ClosingDebit), decStr(r.ClosingCredit),
		})
	}
	grid = append(grid, []string{"合计", "", decStr(od), decStr(oc), decStr(pd), decStr(pc), decStr(cd), decStr(cc)})
	table := widgets.NewGridTable(headers, grid, []float32{70, 160, 110, 110, 110, 110, 110, 110})

	info := widget.NewLabel(fmt.Sprintf("期间：%04d-%02d   试算平衡：%s（期末借 %s / 贷 %s）",
		year, month, boolStr(cd.Equal(cc)), cd.StringFixed(2), cc.StringFixed(2)))
	return container.NewBorder(info, nil, nil, nil, table)
}
