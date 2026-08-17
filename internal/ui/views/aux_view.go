//go:build !linux

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/ui/widgets"
)

// BuildAuxView 辅助核算视图
func BuildAuxView(a App) fyne.CanvasObject {
	typeSel := widget.NewSelect([]string{string(domain.AuxCustomer), string(domain.AuxSupplier), string(domain.AuxDepartment), string(domain.AuxProject)}, nil)
	typeSel.SetSelected(string(domain.AuxCustomer))

	wrap := container.NewStack()
	render := func() {
		t := domain.AuxType(typeSel.Selected)
		items := a.Svc().Aux.List(t)
		headers := []string{"编码", "名称", "启用"}
		rows := make([][]string, 0, len(items))
		for _, it := range items {
			rows = append(rows, []string{it.Code, it.Name, boolStr(it.IsActive)})
		}
		// 辅助余额表
		year, month := a.CurYear(), a.CurMonth()
		balRows := a.Svc().Aux.AuxBalanceReport(year, month, t, "")
		bHeaders := []string{"辅助项编码", "辅助项名称", "科目编码", "科目名称", "借方余额", "贷方余额"}
		bGrid := make([][]string, 0, len(balRows))
		for _, r := range balRows {
			bGrid = append(bGrid, []string{r.ItemCode, r.ItemName, r.AccountCode, r.AccountName, decStr(r.Debit), decStr(r.Credit)})
		}
		wrap.Objects = []fyne.CanvasObject{
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("%s 辅助项 (%d) | %04d-%02d 辅助余额表 (%d)", t, len(items), year, month, len(balRows))),
				widgets.NewGridTable(headers, rows, []float32{120, 200, 80}),
				widgets.NewGridTable(bHeaders, bGrid, []float32{120, 200, 90, 160, 130, 130}),
			),
		}
		wrap.Refresh()
	}
	typeSel.OnChanged = func(string) { render() }
	render()

	addBtn := widget.NewButton("新增辅助项", func() {
		code := widget.NewEntry()
		name := widget.NewEntry()
		form := dialog.NewForm("新增"+typeSel.Selected, "保存", "取消",
			[]*widget.FormItem{
				{Text: "编码", Widget: code},
				{Text: "名称", Widget: name},
			}, func(ok bool) {
				if !ok {
					return
				}
				err := a.Svc().Aux.Create(&domain.AuxiliaryItem{
					Type: domain.AuxType(typeSel.Selected), Code: code.Text, Name: name.Text,
				}, "user")
				if err != nil {
					dialog.ShowError(err, a.Win())
				}
				render()
			}, a.Win())
		form.Resize(fyne.NewSize(360, 220))
		form.Show()
	})

	return container.NewBorder(container.NewHBox(widget.NewLabel("类型"), typeSel, addBtn), nil, nil, nil, wrap)
}
