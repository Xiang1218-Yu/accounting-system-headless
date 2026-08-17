//go:build !linux

package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tog/accounting-system/internal/service"
	"github.com/tog/accounting-system/internal/ui/views"
)

// App 桌面应用上下文
type App struct {
	svc     *service.Container
	fyneApp fyne.App
	win     fyne.Window

	curYear  int
	curMonth int

	content     *fyne.Container
	periodLabel *widget.Label
	builders    []func() fyne.CanvasObject
	curIdx      int
}

// New 构造 UI 应用
func New(svc *service.Container, fyneApp fyne.App) *App {
	y, m := svc.Period.Current()
	if y == 0 {
		y = 2026
		m = 1
	}
	return &App{svc: svc, fyneApp: fyneApp, curYear: y, curMonth: m}
}

// Svc 服务容器
func (a *App) Svc() *service.Container { return a.svc }

// Win 主窗口
func (a *App) Win() fyne.Window { return a.win }

// CurYear 当前期间年
func (a *App) CurYear() int { return a.curYear }

// CurMonth 当前期间月
func (a *App) CurMonth() int { return a.curMonth }

// Run 启动主窗口
func (a *App) Run() {
	a.win = a.fyneApp.NewWindow("会计核算系统")
	a.win.Resize(fyne.NewSize(1200, 760))

	a.periodLabel = widget.NewLabel("")
	a.refreshPeriodLabel()

	navItems := []struct {
		name  string
		build func() fyne.CanvasObject
	}{
		{"科目管理", func() fyne.CanvasObject { return views.BuildAccountsView(a) }},
		{"凭证录入", func() fyne.CanvasObject { return views.BuildVoucherEditView(a) }},
		{"凭证列表", func() fyne.CanvasObject { return views.BuildVoucherListView(a) }},
		{"总账/明细账", func() fyne.CanvasObject { return views.BuildLedgerView(a) }},
		{"余额表", func() fyne.CanvasObject { return views.BuildBalanceView(a) }},
		{"三大报表", func() fyne.CanvasObject { return views.BuildReportsView(a) }},
		{"辅助核算", func() fyne.CanvasObject { return views.BuildAuxView(a) }},
		{"期末处理", func() fyne.CanvasObject { return views.BuildPeriodView(a) }},
		{"Excel导入导出", func() fyne.CanvasObject { return views.BuildExcelView(a) }},
		{"备份恢复", func() fyne.CanvasObject { return views.BuildBackupView(a) }},
		{"操作审计", func() fyne.CanvasObject { return views.BuildAuditView(a) }},
	}

	a.builders = make([]func() fyne.CanvasObject, len(navItems))
	for i, it := range navItems {
		a.builders[i] = it.build
	}
	a.content = container.NewStack(a.builders[0]())
	list := widget.NewList(
		func() int { return len(navItems) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(navItems[i].name)
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		a.curIdx = id
		a.content.Objects = []fyne.CanvasObject{a.builders[id]()}
		a.content.Refresh()
	}

	prevBtn := widget.NewButton("上一期间", func() { a.shiftPeriod(-1) })
	nextBtn := widget.NewButton("下一期间", func() { a.shiftPeriod(1) })
	topBar := container.NewHBox(prevBtn, a.periodLabel, nextBtn,
		widget.NewButton("刷新", func() { a.rebuild() }))

	left := container.NewBorder(topBar, nil, nil, nil, list)
	split := container.NewHSplit(left, a.content)
	split.Offset = 0.16
	a.win.SetContent(container.NewBorder(nil, nil, nil, nil, split))

	list.Select(0)
	a.win.ShowAndRun()
}

func (a *App) refreshPeriodLabel() {
	a.periodLabel.SetText(fmt.Sprintf("当前期间：%04d年%02d月", a.curYear, a.curMonth))
}

func (a *App) shiftPeriod(delta int) {
	a.curMonth += delta
	if a.curMonth > 12 {
		a.curMonth = 1
		a.curYear++
	}
	if a.curMonth < 1 {
		a.curMonth = 12
		a.curYear--
	}
	a.svc.Period.SetCurrent(a.curYear, a.curMonth, "user")
	a.refreshPeriodLabel()
	a.refreshContent()
}

func (a *App) refreshContent() {
	a.rebuild()
}

func (a *App) rebuild() {
	if a.content == nil || a.curIdx >= len(a.builders) {
		return
	}
	a.content.Objects = []fyne.CanvasObject{a.builders[a.curIdx]()}
	a.content.Refresh()
}
