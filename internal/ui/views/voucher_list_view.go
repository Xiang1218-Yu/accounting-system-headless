//go:build !linux

package views

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tog/accounting-system/internal/ui/widgets"
)

// BuildVoucherListView 凭证列表视图
func BuildVoucherListView(a App) fyne.CanvasObject {
	buildTable := func() fyne.CanvasObject {
		vs := a.Svc().Voucher.List()
		headers := []string{"凭证号", "日期", "类型", "状态", "摘要", "借方合计", "贷方合计"}
		rows := make([][]string, 0, len(vs))
		for _, v := range vs {
			rows = append(rows, []string{
				v.Number, v.VoucherDate.Format("2006-01-02"),
				string(v.Type), string(v.Status), v.Summary,
				v.TotalDebit().StringFixed(2), v.TotalCredit().StringFixed(2),
			})
		}
		return widgets.NewGridTable(headers, rows, []float32{140, 100, 80, 70, 200, 110, 110})
	}

	wrap := container.NewStack(buildTable())

	sel := widget.NewSelect([]string{"(无凭证)"}, nil)
	refreshSel := func() {
		vs := a.Svc().Voucher.List()
		opts := make([]string, 0, len(vs))
		for _, v := range vs {
			label := v.Number
			if v.Number == "" {
				label = v.ID[:8]
			}
			opts = append(opts, fmt.Sprintf("%s | %s | %s", label, v.VoucherDate.Format("01-02"), string(v.Status)))
		}
		if len(opts) == 0 {
			opts = []string{"(无凭证)"}
		}
		sel.Options = opts
		sel.ClearSelected()
		sel.Refresh()
	}
	refreshSel()

	selectedID := func() string {
		vs := a.Svc().Voucher.List()
		if sel.SelectedIndex() < 0 || sel.SelectedIndex() >= len(vs) {
			return ""
		}
		return vs[sel.SelectedIndex()].ID
	}

	refreshAll := func() {
		wrap.Objects = []fyne.CanvasObject{buildTable()}
		wrap.Refresh()
		refreshSel()
	}

	approveBtn := widget.NewButton("审核", func() {
		if id := selectedID(); id != "" {
			if err := a.Svc().Voucher.Approve(id, "user"); err != nil {
				dialog.ShowError(err, a.Win())
			}
			refreshAll()
		}
	})
	unapproveBtn := widget.NewButton("反审核", func() {
		if id := selectedID(); id != "" {
			if err := a.Svc().Voucher.Unapprove(id, "user"); err != nil {
				dialog.ShowError(err, a.Win())
			}
			refreshAll()
		}
	})
	postBtn := widget.NewButton("过账", func() {
		if id := selectedID(); id != "" {
			if err := a.Svc().Voucher.Post(id, "user"); err != nil {
				dialog.ShowError(err, a.Win())
			} else {
				dialog.ShowInformation("过账成功", "凭证已过账，余额已更新", a.Win())
			}
			refreshAll()
		}
	})
	unpostBtn := widget.NewButton("反过账", func() {
		if id := selectedID(); id != "" {
			if err := a.Svc().Voucher.Unpost(id, "user"); err != nil {
				dialog.ShowError(err, a.Win())
			}
			refreshAll()
		}
	})
	delBtn := widget.NewButton("删除", func() {
		if id := selectedID(); id != "" {
			dialog.ShowConfirm("确认", "确定删除该凭证?", func(ok bool) {
				if !ok {
					return
				}
				if err := a.Svc().Voucher.Delete(id, "user"); err != nil {
					dialog.ShowError(err, a.Win())
				}
				refreshAll()
			}, a.Win())
		}
	})

	toolbar := container.NewVBox(
		container.NewHBox(sel, approveBtn, unapproveBtn, postBtn, unpostBtn, delBtn),
	)
	return container.NewBorder(toolbar, nil, nil, nil, wrap)
}

// ensureVoucherPeriodValid 工具：构造日期
func voucherDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now()
	}
	return t
}
