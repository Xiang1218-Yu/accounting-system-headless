//go:build !linux

package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// BuildBackupView 备份恢复视图
func BuildBackupView(a App) fyne.CanvasObject {
	backupBtn := widget.NewButton("备份到 zip", func() {
		saveDialog(a, "accounting-backup.zip", func(p string) error {
			return a.Svc().Backup.Backup(p)
		})
	})

	restoreBtn := widget.NewButton("从 zip 恢复", func() {
		dialog.ShowConfirm("恢复", "恢复将覆盖当前数据，且需重启应用生效。继续？", func(ok bool) {
			if !ok {
				return
			}
			dlg := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
				if err != nil || rc == nil {
					return
				}
				path := rc.URI().Path()
				rc.Close()
				if e := a.Svc().Backup.Restore(path); e != nil {
					dialog.ShowError(e, a.Win())
					return
				}
				dialog.ShowInformation("完成", "恢复成功，请重启应用以重载数据", a.Win())
			}, a.Win())
			dlg.Show()
		}, a.Win())
	})

	desc := widget.NewLabel("数据目录：" + a.Svc().Store.Dir())
	return container.NewVBox(
		widget.NewLabel("数据备份与恢复"),
		desc,
		backupBtn, restoreBtn,
	)
}
