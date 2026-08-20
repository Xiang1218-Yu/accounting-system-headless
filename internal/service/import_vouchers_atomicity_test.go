package service

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestImportVouchersRejectsPartialVoucherWhenEntryIsInvalid(t *testing.T) {
	svc := newTestContainer(t)
	path := filepath.Join(t.TempDir(), "vouchers.xlsx")
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	rows := [][]any{
		{"凭证号", "日期", "类型", "状态", "摘要", "科目编码", "科目名称", "借方", "贷方"},
		{"IMP-001", "2026-08-08", "", "", "导入测试", "1002", "银行存款", 100, 0},
		{"IMP-001", "2026-08-08", "", "", "导入测试", "9999", "不存在科目", 0, 100},
	}
	for i, row := range rows {
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+1), &row); err != nil {
			t.Fatalf("write workbook row %d: %v", i+1, err)
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}

	count, err := svc.Excel.ImportVouchers(path)
	if err == nil {
		t.Fatalf("包含无效分录的凭证导入应失败，实际导入 %d 张", count)
	}
	if count != 0 {
		t.Fatalf("导入失败时不应保存残缺草稿，实际导入 %d 张", count)
	}
	if got := len(svc.Voucher.List()); got != 0 {
		t.Fatalf("导入失败时不应留下草稿凭证，实际有 %d 张", got)
	}
}
