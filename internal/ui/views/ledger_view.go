//go:build !linux

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tog/accounting-system/internal/ui/widgets"
)

// BuildLedgerView 总账/明细账视图
func BuildLedgerView(a App) fyne.CanvasObject {
	accs := a.Svc().Account.List()
	opts := make([]string, 0, len(accs))
	codeOf := map[string]string{}
	for _, ac := range accs {
		if ac.IsLeaf {
			opt := ac.Code + " " + ac.Name
			opts = append(opts, opt)
			codeOf[opt] = ac.Code
		}
	}
	sel := widget.NewSelect(opts, nil)
	if len(opts) > 0 {
		sel.SetSelected(opts[0])
	}

	wrap := container.NewStack()
	render := func() {
		if sel.SelectedIndex() < 0 || sel.SelectedIndex() >= len(opts) {
			return
		}
		code := codeOf[opts[sel.SelectedIndex()]]
		year, month := a.CurYear(), a.CurMonth()
		lines := a.Svc().Ledger.DetailLedger(year, month, code)
		headers := []string{"日期", "凭证号", "摘要", "借方", "贷方", "方向", "余额"}
		grid := make([][]string, 0, len(lines))
		for _, l := range lines {
			grid = append(grid, []string{l.Date, l.VoucherNumber, l.Summary, decStr(l.Debit), decStr(l.Credit), string(l.Direction), l.Balance.StringFixed(2)})
		}
		wrap.Objects = []fyne.CanvasObject{widgets.NewGridTable(headers, grid, []float32{100, 140, 200, 110, 110, 60, 120})}
		wrap.Refresh()
	}
	sel.OnChanged = func(string) { render() }
	render()

	// 总账区
	glCode := func() string {
		if sel.SelectedIndex() < 0 || sel.SelectedIndex() >= len(opts) {
			return ""
		}
		return codeOf[opts[sel.SelectedIndex()]]
	}
	glBtn := widget.NewButton("查看全年总账", func() {
		code := glCode()
		if code == "" {
			return
		}
		year := a.CurYear()
		od, oc, glRows, cd, cc := a.Svc().Ledger.GeneralLedger(year, code)
		headers := []string{"月份", "本期借方", "本期贷方", "期末借方", "期末贷方"}
		grid := make([][]string, 0, len(glRows)+2)
		grid = append(grid, []string{"期初", "", "", decStr(od), decStr(oc)})
		for _, r := range glRows {
			grid = append(grid, []string{fmt.Sprintf("%02d月", r.Month), decStr(r.PeriodDebit), decStr(r.PeriodCredit), decStr(r.ClosingDebit), decStr(r.ClosingCredit)})
		}
		grid = append(grid, []string{"年末", "", "", decStr(cd), decStr(cc)})
		dlg := dialog.NewCustom("总账 "+code, "关闭", widgets.NewGridTable(headers, grid, []float32{80, 130, 130, 130, 130}), a.Win())
		dlg.Resize(fyne.NewSize(620, 480))
		dlg.Show()
	})

	toolbar := container.NewHBox(widget.NewLabel("科目"), sel, glBtn)
	return container.NewBorder(toolbar, nil, nil, nil, wrap)
}
