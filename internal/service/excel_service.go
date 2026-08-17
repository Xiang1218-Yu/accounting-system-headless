package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
	"github.com/xuri/excelize/v2"
)

// ExcelService Excel 导入导出
type ExcelService struct {
	st *store.Store
}

// NewExcelService 构造
func NewExcelService(st *store.Store) *ExcelService {
	return &ExcelService{st: st}
}

// ExportVouchers 导出凭证列表到 xlsx
func (x *ExcelService) ExportVouchers(path string, vouchers []*domain.Voucher) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	headers := []string{"凭证号", "日期", "类型", "状态", "摘要", "科目编码", "科目名称", "借方", "贷方"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	row := 2
	for _, v := range vouchers {
		for _, e := range v.Entries {
			acc := x.st.GetAccount(e.AccountCode)
			name := ""
			if acc != nil {
				name = acc.Name
			}
			x.setRow(f, sheet, row, v.Number, v.VoucherDate.Format("2006-01-02"), string(v.Type), string(v.Status), e.Summary, e.AccountCode, name, e.Debit.InexactFloat64(), e.Credit.InexactFloat64())
			row++
		}
	}
	return f.SaveAs(path)
}

// ExportTrialBalance 导出余额表
func (x *ExcelService) ExportTrialBalance(path string, rows []BalanceRow) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	headers := []string{"科目编码", "科目名称", "期初借方", "期初贷方", "本期借方", "本期贷方", "期末借方", "期末贷方"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, r := range rows {
		x.setRow(f, sheet, i+2, r.AccountCode, r.AccountName, r.OpeningDebit.InexactFloat64(), r.OpeningCredit.InexactFloat64(), r.PeriodDebit.InexactFloat64(), r.PeriodCredit.InexactFloat64(), r.ClosingDebit.InexactFloat64(), r.ClosingCredit.InexactFloat64())
	}
	return f.SaveAs(path)
}

// ExportBalanceSheet 导出资产负债表
func (x *ExcelService) ExportBalanceSheet(path string, bs *BalanceSheet) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", fmt.Sprintf("资产负债表 %s", bs.Period))
	r := 3
	writeBS := func(title string, rows []BSRow, total decimal.Decimal) {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), title)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), "金额")
		r++
		for _, row := range rows {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", r), row.Name)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", r), row.Amount.InexactFloat64())
			r++
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "合计")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), total.InexactFloat64())
		r += 2
	}
	writeBS("资产", bs.AssetRows, bs.AssetTotal)
	writeBS("负债", bs.LiabRows, bs.LiabTotal)
	writeBS("所有者权益", bs.EquityRows, bs.EquityTotal)
	return f.SaveAs(path)
}

// ExportIncomeStatement 导出利润表
func (x *ExcelService) ExportIncomeStatement(path string, is *IncomeStatement) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", fmt.Sprintf("利润表 %s", is.Period))
	f.SetCellValue(sheet, "A3", "项目")
	f.SetCellValue(sheet, "B3", "本期金额")
	f.SetCellValue(sheet, "C3", "本年累计")
	for i, r := range is.Rows {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+4), r.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+4), r.Current.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+4), r.YTD.InexactFloat64())
	}
	return f.SaveAs(path)
}

// ExportCashFlow 导出现金流量表
func (x *ExcelService) ExportCashFlow(path string, cf *CashFlowStatement) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "A1", fmt.Sprintf("现金流量表 %s", cf.Period))
	f.SetCellValue(sheet, "A3", "项目")
	f.SetCellValue(sheet, "B3", "金额")
	for i, r := range cf.Rows {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+4), r.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+4), r.Amount.InexactFloat64())
	}
	return f.SaveAs(path)
}

// ImportAccounts 从 xlsx 导入科目（编码|名称|方向|级别|父编码|辅助核算|外币|币种）
func (x *ExcelService) ImportAccounts(path string) (int, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, err
	}
	count := 0
	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}
		code, name, dir := row[0], row[1], row[2]
		acc := &domain.Account{
			Code: code, Name: name, Direction: domain.BalanceDirection(dir),
			Level: 1, IsLeaf: true, IsActive: true,
		}
		acc.Category = domain.CategoryByCode(code)
		if x.st.GetAccount(code) == nil {
			_ = x.st.Mutate(func() { x.st.SetAccount(acc) })
			count++
		}
	}
	return count, nil
}

// ImportVouchers 从 xlsx 导入凭证（每行一条分录，按凭证号聚合，导入为草稿）
func (x *ExcelService) ImportVouchers(path string) (int, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, err
	}
	type tmp struct {
		date    string
		summary string
		entries []domain.VoucherEntry
	}
	groups := map[string]*tmp{}
	order := []string{}
	for i, row := range rows {
		if i == 0 || len(row) < 8 {
			continue
		}
		num, date, summary, code := row[0], row[1], row[4], row[5]
		debit := parseDec(row[7])
		credit := parseDec(row[8])
		if x.st.GetAccount(code) == nil {
			continue
		}
		g, ok := groups[num]
		if !ok {
			g = &tmp{date: date, summary: summary}
			groups[num] = g
			order = append(order, num)
		}
		g.entries = append(g.entries, domain.VoucherEntry{AccountCode: code, Summary: summary, Debit: debit, Credit: credit})
	}
	count := 0
	for _, num := range order {
		g := groups[num]
		t, err := time.Parse("2006-01-02", g.date)
		if err != nil {
			t = time.Now()
		}
		v := &domain.Voucher{
			ID:          uuid.NewString(),
			VoucherDate: t,
			Status:      domain.StatusDraft,
			Type:        domain.VoucherNormal,
			Summary:     g.summary,
			Entries:     g.entries,
		}
		v.PeriodYear = t.Year()
		v.PeriodMonth = int(t.Month())
		if err := x.st.Mutate(func() { x.st.SetVoucher(v) }); err != nil {
			continue
		}
		count++
	}
	return count, nil
}

func (x *ExcelService) setRow(f *excelize.File, sheet string, row int, vals ...any) {
	for i, v := range vals {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheet, cell, v)
	}
}

func parseDec(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}
