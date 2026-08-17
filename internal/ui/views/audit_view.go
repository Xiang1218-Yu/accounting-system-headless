//go:build !linux

package views

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tog/accounting-system/internal/ui/widgets"
)

// BuildAuditView 操作审计视图
func BuildAuditView(a App) fyne.CanvasObject {
	logs := a.Svc().Store.ListAudit()
	headers := []string{"时间", "用户", "操作", "实体", "实体ID", "操作ID"}
	rows := make([][]string, 0, len(logs))
	for _, l := range logs {
		rows = append(rows, []string{
			l.Timestamp.Format("2006-01-02 15:04:05"), l.User, l.Action, l.Entity, l.EntityID, l.OpID,
		})
	}
	table := widgets.NewGridTable(headers, rows, []float32{150, 80, 120, 90, 120, 120})
	return container.NewBorder(
		widget.NewLabel("操作审计日志（共 "+strconv.Itoa(len(logs))+" 条）"),
		nil, nil, nil, container.NewVScroll(table),
	)
}
