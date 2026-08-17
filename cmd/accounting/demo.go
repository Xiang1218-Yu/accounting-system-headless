package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/service"
)

// demoCmd 生成演示数据并过账，便于运行时验证
func demoCmd(svc *service.Container) {
	year, month := 2026, 8
	svc.Period.EnsurePeriod(year, month)

	type demo struct {
		date    string
		summary string
		entries []domain.VoucherEntry
	}
	demos := []demo{
		{"2026-08-05", "股东投入资本", []domain.VoucherEntry{
			{AccountCode: "1002", Summary: "收到投资款", Debit: dec("500000")},
			{AccountCode: "3001", Summary: "实收资本", Credit: dec("500000")},
		}},
		{"2026-08-08", "销售商品收款", []domain.VoucherEntry{
			{AccountCode: "1002", Summary: "销售收款", Debit: dec("113000")},
			{AccountCode: "5001", Summary: "主营业务收入", Credit: dec("100000")},
			{AccountCode: "2221", Summary: "应交增值税", Credit: dec("13000")},
		}},
		{"2026-08-10", "采购商品付款", []domain.VoucherEntry{
			{AccountCode: "1405", Summary: "入库", Debit: dec("60000")},
			{AccountCode: "2221", Summary: "进项税", Debit: dec("7800")},
			{AccountCode: "1002", Summary: "付款", Credit: dec("67800")},
		}},
		{"2026-08-15", "结转销售成本", []domain.VoucherEntry{
			{AccountCode: "5401", Summary: "主营业务成本", Debit: dec("55000")},
			{AccountCode: "1405", Summary: "出库", Credit: dec("55000")},
		}},
		{"2026-08-20", "支付办公租金", []domain.VoucherEntry{
			{AccountCode: "5602", Summary: "管理费用-租金", Debit: dec("8000")},
			{AccountCode: "1002", Summary: "付款", Credit: dec("8000")},
		}},
		{"2026-08-25", "支付员工薪酬", []domain.VoucherEntry{
			{AccountCode: "5602", Summary: "管理费用-薪酬", Debit: dec("20000")},
			{AccountCode: "2211", Summary: "应付职工薪酬", Credit: dec("20000")},
		}},
		{"2026-08-28", "计提折旧", []domain.VoucherEntry{
			{AccountCode: "5602", Summary: "管理费用-折旧", Debit: dec("3000")},
			{AccountCode: "1602", Summary: "累计折旧", Credit: dec("3000")},
		}},
	}

	count := 0
	for _, d := range demos {
		t, _ := time.Parse("2006-01-02", d.date)
		v := &domain.Voucher{
			VoucherDate: t,
			Type:        domain.VoucherNormal,
			Summary:     d.summary,
			Entries:     d.entries,
		}
		if err := svc.Voucher.Create(v, "demo"); err != nil {
			fmt.Fprintln(os.Stderr, "创建凭证失败:", err)
			continue
		}
		if err := svc.Voucher.Approve(v.ID, "demo"); err != nil {
			fmt.Fprintln(os.Stderr, "审核失败:", err)
			continue
		}
		if err := svc.Voucher.Post(v.ID, "demo"); err != nil {
			fmt.Fprintln(os.Stderr, "过账失败:", err)
			continue
		}
		count++
	}
	fmt.Printf("已生成并过账 %d 张演示凭证（%04d-%02d）\n", count, year, month)

	// 结转损益
	if err := svc.Closing.CloseProfitLoss(year, month, "demo"); err != nil {
		fmt.Fprintln(os.Stderr, "结转损益:", err)
	} else {
		fmt.Println("已结转损益")
	}

	// 自检
	r := svc.Ledger.Verify(year, month)
	fmt.Println("一致性自检:", passStr(r.OK))
	for _, e := range r.Errors {
		fmt.Println("  ✗", e)
	}
}

func dec(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
