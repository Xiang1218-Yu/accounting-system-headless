//go:build !linux

package views

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

type entryRow struct {
	account *widget.Select
	summary *widget.Entry
	debit   *widget.Entry
	credit  *widget.Entry
}

// BuildVoucherEditView 凭证录入视图
func BuildVoucherEditView(a App) fyne.CanvasObject {
	accountOpts := func() []string {
		accs := a.Svc().Account.List()
		opts := make([]string, 0, len(accs))
		for _, ac := range accs {
			if ac.IsActive && ac.IsLeaf {
				opts = append(opts, ac.Code+" "+ac.Name)
			}
		}
		return opts
	}()

	dateEntry := widget.NewEntry()
	dateEntry.SetText(time.Now().Format("2006-01-02"))
	summaryEntry := widget.NewEntry()

	rows := []*entryRow{}
	rowsBox := container.NewVBox()
	scroll := container.NewVScroll(rowsBox)
	scroll.SetMinSize(fyne.NewSize(800, 300))

	balanceLabel := widget.NewLabel("借：0.00  贷：0.00  差额：0.00")

	recompute := func() {
		td, tc := decimal.Zero, decimal.Zero
		for _, r := range rows {
			td = td.Add(parseDecSafe(r.debit.Text))
			tc = tc.Add(parseDecSafe(r.credit.Text))
		}
		diff := td.Sub(tc)
		balanceLabel.SetText(fmt.Sprintf("借：%s  贷：%s  差额：%s", td.StringFixed(2), tc.StringFixed(2), diff.StringFixed(2)))
		if diff.IsZero() && td.GreaterThan(decimal.Zero) {
			balanceLabel.TextStyle = fyne.TextStyle{Bold: true}
		}
	}

	var rebuildRows func()
	rebuildRows = func() {
		rowsBox.Objects = nil
		for _, r := range rows {
			rr := r
			rm := widget.NewButton("删除", func() {
				for i, x := range rows {
					if x == rr {
						rows = append(rows[:i], rows[i+1:]...)
						break
					}
				}
				rebuildRows()
				recompute()
			})
			row := container.NewHBox(
				container.NewMax(container.NewGridWithColumns(1, rr.account)),
				rr.summary, rr.debit, rr.credit, rm,
			)
			rowsBox.Add(row)
		}
		rowsBox.Refresh()
	}

	addRow := func() {
		r := &entryRow{
			account: widget.NewSelect(accountOpts, nil),
			summary: widget.NewEntry(),
			debit:   widget.NewEntry(),
			credit:  widget.NewEntry(),
		}
		r.debit.OnChanged = func(string) { recompute() }
		r.credit.OnChanged = func(string) { recompute() }
		if len(accountOpts) > 0 {
			r.account.SetSelected(accountOpts[0])
		}
		rows = append(rows, r)
		rebuildRows()
		recompute()
	}
	addRow()
	addRow()

	saveBtn := widget.NewButton("保存草稿", func() {
		entries := make([]domain.VoucherEntry, 0, len(rows))
		for _, r := range rows {
			code := accountCodeFromOpt(r.account.Selected)
			if code == "" {
				continue
			}
			entries = append(entries, domain.VoucherEntry{
				AccountCode: code,
				Summary:     r.summary.Text,
				Debit:       parseDecSafe(r.debit.Text),
				Credit:      parseDecSafe(r.credit.Text),
			})
		}
		v := &domain.Voucher{
			VoucherDate: voucherDate(dateEntry.Text),
			Type:        domain.VoucherNormal,
			Summary:     summaryEntry.Text,
			Entries:     entries,
		}
		if err := a.Svc().Voucher.Create(v, "user"); err != nil {
			dialog.ShowError(err, a.Win())
			return
		}
		dialog.ShowInformation("已保存", "凭证已保存为草稿，请到凭证列表审核过账", a.Win())
	})

	addBtn := widget.NewButton("+ 增加分录", addRow)

	header := container.NewHBox(
		widget.NewLabel("日期"), dateEntry,
		widget.NewLabel("凭证摘要"), summaryEntry,
	)
	colHeader := container.NewHBox(
		widget.NewLabel("科目"), widget.NewLabel("              "),
		widget.NewLabel("摘要"), widget.NewLabel("    "),
		widget.NewLabel("借方"), widget.NewLabel("    "),
		widget.NewLabel("贷方"), widget.NewLabel("    "),
	)
	toolbar := container.NewHBox(addBtn, saveBtn, balanceLabel)

	return container.NewBorder(
		container.NewVBox(header, colHeader, toolbar),
		nil, nil, nil, scroll,
	)
}

func accountCodeFromOpt(opt string) string {
	if len(opt) < 4 {
		return ""
	}
	return opt[:4]
}

func parseDecSafe(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}
