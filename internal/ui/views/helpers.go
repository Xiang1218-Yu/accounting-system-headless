//go:build !linux

package views

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/service"
)

// App 视图所需的应用上下文接口（打破与 ui 包的循环依赖）
type App interface {
	Svc() *service.Container
	Win() fyne.Window
	CurYear() int
	CurMonth() int
}

func boolStr(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return n
}

func decStr(d decimal.Decimal) string {
	if d.IsZero() {
		return ""
	}
	return d.StringFixed(2)
}

// saveDialog 弹出文件保存对话框并执行保存
func saveDialog(a App, defaultName string, save func(string) error) {
	dlg := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		path := wc.URI().Path()
		wc.Close()
		if e := save(path); e != nil {
			dialog.ShowError(e, a.Win())
			return
		}
		dialog.ShowInformation("完成", "已保存到 "+path, a.Win())
	}, a.Win())
	dlg.SetFileName(defaultName)
	dlg.Show()
}
