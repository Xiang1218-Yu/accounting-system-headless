//go:build !linux

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// BuildPeriodView 期末处理视图
func BuildPeriodView(a App) fyne.CanvasObject {
	year, month := a.CurYear(), a.CurMonth()
	info := widget.NewLabel("")

	refreshInfo := func() {
		p := a.Svc().Period.Get(year, month)
		status := "未建账"
		if p != nil {
			status = string(p.Status)
		}
		info.SetText(fmt.Sprintf("期间：%04d-%02d   状态：%s   已结转损益：%s   已调汇：%s",
			year, month, status,
			boolStr(a.Svc().Closing.HasClosed(year, month)),
			boolStr(a.Svc().FX.HasAdjusted(year, month))))
	}
	refreshInfo()

	// 期末汇率设置
	usdRate := widget.NewEntry()
	usdRate.SetPlaceHolder("如 7.2")
	setRateBtn := widget.NewButton("设置USD期末汇率", func() {
		rate := atofSafe(usdRate.Text)
		if rate <= 0 {
			dialog.ShowError(fmt.Errorf("汇率无效"), a.Win())
			return
		}
		if err := a.Svc().Period.SetFXRate(year, month, map[string]float64{"USD": rate}, "user"); err != nil {
			dialog.ShowError(err, a.Win())
		} else {
			dialog.ShowInformation("完成", "期末汇率已设置", a.Win())
		}
		refreshInfo()
	})

	fxBtn := widget.NewButton("期末调汇", func() {
		if err := a.Svc().FX.AdjustFX(year, month, "user"); err != nil {
			dialog.ShowError(err, a.Win())
		} else {
			dialog.ShowInformation("完成", "期末调汇凭证已生成并过账", a.Win())
		}
		refreshInfo()
	})

	closeBtn := widget.NewButton("结转损益", func() {
		if err := a.Svc().Closing.CloseProfitLoss(year, month, "user"); err != nil {
			dialog.ShowError(err, a.Win())
		} else {
			dialog.ShowInformation("完成", "结转损益凭证已生成并过账", a.Win())
		}
		refreshInfo()
	})

	monthCloseBtn := widget.NewButton("月末结账", func() {
		dialog.ShowConfirm("月末结账", fmt.Sprintf("确定结账 %04d-%02d？结账后该期凭证不可修改。", year, month), func(ok bool) {
			if !ok {
				return
			}
			if err := a.Svc().Period.Close(year, month, "user"); err != nil {
				dialog.ShowError(err, a.Win())
			} else {
				dialog.ShowInformation("完成", "已结账", a.Win())
			}
			refreshInfo()
		}, a.Win())
	})

	reopenBtn := widget.NewButton("反结账", func() {
		if err := a.Svc().Period.Reopen(year, month, "user"); err != nil {
			dialog.ShowError(err, a.Win())
		}
		refreshInfo()
	})

	verifyBtn := widget.NewButton("一致性自检", func() {
		r := a.Svc().Ledger.Verify(year, month)
		msg := fmt.Sprintf("结果：%s\n检查项：%v", boolStr(r.OK), r.Checks)
		if len(r.Errors) > 0 {
			msg += "\n错误：" + fmt.Sprint(r.Errors)
		}
		dialog.ShowInformation("自检", msg, a.Win())
	})

	toolbar := container.NewVBox(
		info,
		container.NewHBox(usdRate, setRateBtn, fxBtn, closeBtn, monthCloseBtn, reopenBtn, verifyBtn),
	)
	return container.NewBorder(toolbar, widget.NewLabel("建议顺序：设置汇率 → 期末调汇 → 结转损益 → 月末结账 → 一致性自检"), nil, nil, nil)
}

func atofSafe(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
