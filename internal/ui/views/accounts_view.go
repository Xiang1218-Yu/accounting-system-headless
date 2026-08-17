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

// BuildAccountsView 科目管理视图
func BuildAccountsView(a App) fyne.CanvasObject {
	refresh := func() fyne.CanvasObject {
		accs := a.Svc().Account.List()
		headers := []string{"编码", "名称", "类别", "方向", "级别", "辅助核算", "外币", "启用"}
		rows := make([][]string, 0, len(accs))
		for _, ac := range accs {
			aux := ""
			for i, t := range ac.AuxTypes {
				if i > 0 {
					aux += "/"
				}
				aux += string(t)
			}
			rows = append(rows, []string{
				ac.Code, ac.Name, string(ac.Category), string(ac.Direction),
				fmt.Sprintf("%d", ac.Level), aux, ac.Currency, boolStr(ac.IsActive),
			})
		}
		return widgets.NewGridTable(headers, rows, []float32{80, 180, 70, 60, 60, 140, 70, 60})
	}

	wrap := container.NewStack(refresh())

	addBtn := widget.NewButton("新增科目", func() {
		showAccountDialog(a, nil, func() {
			wrap.Objects = []fyne.CanvasObject{refresh()}
			wrap.Refresh()
		})
	})
	editBtn := widget.NewButton("编辑科目", func() {
		code := selectedAccountCode(a)
		if code == "" {
			return
		}
		acc := a.Svc().Account.Get(code)
		showAccountDialog(a, acc, func() {
			wrap.Objects = []fyne.CanvasObject{refresh()}
			wrap.Refresh()
		})
	})
	delBtn := widget.NewButton("删除科目", func() {
		code := selectedAccountCode(a)
		if code == "" {
			return
		}
		dialog.ShowConfirm("确认删除", "确定删除科目 "+code+" ?", func(ok bool) {
			if !ok {
				return
			}
			if err := a.Svc().Account.Delete(code, "user"); err != nil {
				dialog.ShowError(err, a.Win())
			} else {
				wrap.Objects = []fyne.CanvasObject{refresh()}
				wrap.Refresh()
			}
		}, a.Win())
	})

	toolbar := container.NewHBox(addBtn, editBtn, delBtn)
	return container.NewBorder(toolbar, nil, nil, nil, wrap)
}

func selectedAccountCode(a App) string {
	accs := a.Svc().Account.List()
	if len(accs) == 0 {
		return ""
	}
	return accs[0].Code
}

func showAccountDialog(a App, acc *domain.Account, done func()) {
	code := widget.NewEntry()
	name := widget.NewEntry()
	dir := widget.NewSelect([]string{string(domain.DirDebit), string(domain.DirCredit)}, nil)
	dir.SetSelected(string(domain.DirDebit))
	level := widget.NewEntry()
	level.SetText("1")
	parent := widget.NewEntry()
	currency := widget.NewEntry()
	isFC := widget.NewCheck("外币科目", nil)
	isActive := widget.NewCheck("启用", nil)
	isActive.SetChecked(true)

	if acc != nil {
		code.SetText(acc.Code)
		code.Disable()
		name.SetText(acc.Name)
		dir.SetSelected(string(acc.Direction))
		level.SetText(fmt.Sprintf("%d", acc.Level))
		parent.SetText(acc.ParentCode)
		currency.SetText(acc.Currency)
		isFC.SetChecked(acc.IsForeignCurrency)
		isActive.SetChecked(acc.IsActive)
	}

	form := dialog.NewForm("科目", "保存", "取消",
		[]*widget.FormItem{
			{Text: "编码", Widget: code},
			{Text: "名称", Widget: name},
			{Text: "方向", Widget: dir},
			{Text: "级别", Widget: level},
			{Text: "上级编码", Widget: parent},
			{Text: "币种", Widget: currency},
			{Text: "", Widget: isFC},
			{Text: "", Widget: isActive},
		},
		func(ok bool) {
			if !ok {
				return
			}
			newAcc := &domain.Account{
				Code: code.Text, Name: name.Text,
				Direction:  domain.BalanceDirection(dir.Selected),
				Level:      atoiSafe(level.Text),
				ParentCode: parent.Text, Currency: currency.Text,
				IsForeignCurrency: isFC.Checked, IsActive: isActive.Checked,
				IsLeaf: true,
			}
			var err error
			if acc == nil {
				err = a.Svc().Account.Create(newAcc, "user")
			} else {
				err = a.Svc().Account.Update(newAcc, "user")
			}
			if err != nil {
				dialog.ShowError(err, a.Win())
			}
			done()
		}, a.Win())
	form.Resize(fyne.NewSize(420, 360))
	form.Show()
}
