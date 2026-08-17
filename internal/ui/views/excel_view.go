//go:build !linux

package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// BuildExcelView Excel 导入导出视图
func BuildExcelView(a App) fyne.CanvasObject {
	exportVouchersBtn := widget.NewButton("导出全部凭证", func() {
		saveDialog(a, "凭证.xlsx", func(p string) error {
			return a.Svc().Excel.ExportVouchers(p, a.Svc().Voucher.List())
		})
	})
	exportTrialBtn := widget.NewButton("导出余额表", func() {
		year, month := a.CurYear(), a.CurMonth()
		saveDialog(a, "余额表.xlsx", func(p string) error {
			return a.Svc().Excel.ExportTrialBalance(p, a.Svc().Ledger.TrialBalance(year, month))
		})
	})

	importVouchersBtn := widget.NewButton("导入凭证(xlsx)", func() {
		dlg := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			path := rc.URI().Path()
			rc.Close()
			n, e := a.Svc().Excel.ImportVouchers(path)
			if e != nil {
				dialog.ShowError(e, a.Win())
				return
			}
			dialog.ShowInformation("完成", "已导入凭证草稿", a.Win())
			_ = n
		}, a.Win())
		dlg.Show()
	})

	importAccountsBtn := widget.NewButton("导入科目(xlsx)", func() {
		dlg := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			path := rc.URI().Path()
			rc.Close()
			n, e := a.Svc().Excel.ImportAccounts(path)
			if e != nil {
				dialog.ShowError(e, a.Win())
				return
			}
			dialog.ShowInformation("完成", "已导入科目", a.Win())
			_ = n
		}, a.Win())
		dlg.Show()
	})

	desc := widget.NewLabel("导入凭证模板：凭证号|日期|类型|摘要|摘要|科目编码|科目名称|借方|贷方（首行表头）")
	return container.NewVBox(
		widget.NewLabel("Excel 导入导出"),
		desc,
		exportVouchersBtn, exportTrialBtn,
		importVouchersBtn, importAccountsBtn,
	)
}
