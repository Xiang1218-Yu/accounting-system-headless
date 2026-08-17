//go:build !linux

package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// NewGridTable 由表头与数据行构建带粘性表头的表格
func NewGridTable(headers []string, rows [][]string, widths []float32) *widget.Table {
	data := append([][]string{headers}, rows...)
	t := widget.NewTable(
		func() (rows int, cols int) { return len(data), len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			if id.Row < len(data) && id.Col < len(data[id.Row]) {
				lbl.SetText(data[id.Row][id.Col])
			} else {
				lbl.SetText("")
			}
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				lbl.TextStyle = fyne.TextStyle{}
			}
		},
	)
	for i, w := range widths {
		if i < len(headers) {
			t.SetColumnWidth(i, w)
		}
	}
	// 行高
	t.CreateHeader = func() fyne.CanvasObject { return widget.NewLabel("") }
	return t
}
