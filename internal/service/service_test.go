package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

func newTestContainer(t *testing.T) *Container {
	t.Helper()
	dir := t.TempDir()
	st, err := store.New(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	svc := NewContainer(st)
	if _, err := svc.Seed.SeedIfEmpty(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return svc
}

func TestPostAndReports(t *testing.T) {
	svc := newTestContainer(t)
	year, month := 2026, 8
	svc.Period.EnsurePeriod(year, month)

	mk := func(date, summary string, entries []domain.VoucherEntry) {
		v := &domain.Voucher{VoucherDate: parseT(date), Type: domain.VoucherNormal, Summary: summary, Entries: entries}
		if err := svc.Voucher.Create(v, "tester"); err != nil {
			t.Fatalf("create %s: %v", summary, err)
		}
		if err := svc.Voucher.Approve(v.ID, "tester"); err != nil {
			t.Fatalf("approve %s: %v", summary, err)
		}
		if err := svc.Voucher.Post(v.ID, "tester"); err != nil {
			t.Fatalf("post %s: %v", summary, err)
		}
	}

	mk("2026-08-05", "投入资本", []domain.VoucherEntry{
		{AccountCode: "1002", Debit: dec("500000")},
		{AccountCode: "3001", Credit: dec("500000")},
	})
	mk("2026-08-10", "销售收款", []domain.VoucherEntry{
		{AccountCode: "1002", Debit: dec("113000")},
		{AccountCode: "5001", Credit: dec("100000")},
		{AccountCode: "2221", Credit: dec("13000")},
	})
	mk("2026-08-20", "支付费用", []domain.VoucherEntry{
		{AccountCode: "5602", Debit: dec("8000")},
		{AccountCode: "1002", Credit: dec("8000")},
	})

	// 余额表试算平衡
	rows := svc.Ledger.TrialBalance(year, month)
	_, _, _, _, cd, cc := BalanceTotals(rows)
	if !cd.Equal(cc) {
		t.Fatalf("试算不平衡: 借 %s 贷 %s", cd, cc)
	}

	// 资产负债表（结转损益前损益类科目有余额，BS 在结转后才平衡）
	bs := svc.Ledger.BuildBalanceSheet(year, month)
	// 货币资金应为 500000 + 113000 - 8000 = 605000
	if !bs.AssetRows[0].Amount.Equal(dec("605000")) {
		t.Fatalf("货币资金=%s，期望 605000", bs.AssetRows[0].Amount)
	}

	// 结转损益后利润表仍正确
	if err := svc.Closing.CloseProfitLoss(year, month, "tester"); err != nil {
		t.Fatalf("closing: %v", err)
	}
	// 结转后资产负债表应平衡
	bs2 := svc.Ledger.BuildBalanceSheet(year, month)
	if !bs2.Balanced {
		t.Fatalf("结转后资产负债表不平衡: 资产 %s 负债+权益 %s", bs2.AssetTotal, bs2.LiabTotal.Add(bs2.EquityTotal))
	}
	is := svc.Ledger.BuildIncomeStatement(year, month)
	// 净利润 = 100000 - 8000 = 92000
	if !is.NetProfit.Equal(dec("92000")) {
		t.Fatalf("净利润=%s，期望 92000", is.NetProfit)
	}

	// 现金流量表对账
	cf := svc.Ledger.BuildCashFlow(year, month)
	if !cf.Reconciled {
		t.Fatalf("现金流量表对账不平")
	}

	// 自检
	r := svc.Ledger.Verify(year, month)
	if !r.OK {
		t.Fatalf("verify 失败: %v", r.Errors)
	}
}

func TestUnbalancedVoucherRejected(t *testing.T) {
	svc := newTestContainer(t)
	v := &domain.Voucher{
		VoucherDate: parseT("2026-08-01"),
		Type:        domain.VoucherNormal,
		Entries: []domain.VoucherEntry{
			{AccountCode: "1002", Debit: dec("100")},
			{AccountCode: "3001", Credit: dec("99")},
		},
	}
	if err := svc.Voucher.Create(v, "tester"); err != nil {
		t.Fatalf("草稿应允许不平衡: %v", err)
	}
	if err := svc.Voucher.Approve(v.ID, "tester"); err == nil {
		t.Fatalf("不平衡凭证不应通过审核")
	}
}

func TestBackupRestore(t *testing.T) {
	svc := newTestContainer(t)
	svc.Period.EnsurePeriod(2026, 8)
	v := &domain.Voucher{VoucherDate: parseT("2026-08-05"), Type: domain.VoucherNormal,
		Entries: []domain.VoucherEntry{
			{AccountCode: "1002", Debit: dec("1000")},
			{AccountCode: "3001", Credit: dec("1000")},
		}}
	_ = svc.Voucher.Create(v, "tester")
	_ = svc.Voucher.Approve(v.ID, "tester")
	_ = svc.Voucher.Post(v.ID, "tester")

	zip := filepath.Join(t.TempDir(), "b.zip")
	if err := svc.Backup.Backup(zip); err != nil {
		t.Fatalf("backup: %v", err)
	}
	if _, err := os.Stat(zip); err != nil {
		t.Fatalf("zip not created: %v", err)
	}
	dir2 := t.TempDir()
	if err := store.Restore(zip, dir2); err != nil {
		t.Fatalf("restore: %v", err)
	}
	st2, _ := store.New(dir2)
	svc2 := NewContainer(st2)
	r := svc2.Ledger.Verify(2026, 8)
	if !r.OK {
		t.Fatalf("恢复后自检失败: %v", r.Errors)
	}
}

func parseT(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func dec(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
