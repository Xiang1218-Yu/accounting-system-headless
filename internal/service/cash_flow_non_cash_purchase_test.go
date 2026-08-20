package service

import (
	"testing"

	"github.com/tog/accounting-system/internal/domain"
)

func TestCashFlowExcludesNonCashAssetPurchase(t *testing.T) {
	svc := newTestContainer(t)
	svc.Period.EnsurePeriod(2026, 8)
	post := func(summary string, entries []domain.VoucherEntry) {
		t.Helper()
		v := &domain.Voucher{
			VoucherDate: parseT("2026-08-08"),
			Type:        domain.VoucherNormal,
			Summary:     summary,
			Entries:     entries,
		}
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

	post("现金购置设备", []domain.VoucherEntry{
		{AccountCode: "1601", Debit: dec("100")},
		{AccountCode: "1002", Credit: dec("100")},
	})
	post("应付购置设备", []domain.VoucherEntry{
		{AccountCode: "1601", Debit: dec("100")},
		{AccountCode: "2202", Credit: dec("100")},
	})

	cf := svc.Ledger.BuildCashFlow(2026, 8)
	for _, row := range cf.Rows {
		if row.Name == "购建固定资产、无形资产和其他长期资产支付的现金" {
			if !row.Amount.Equal(dec("100")) {
				t.Fatalf("应付购置没有现金分录，不应计入购建长期资产支付现金，实际 %s", row.Amount)
			}
			return
		}
	}
	t.Fatal("现金流量表缺少购建长期资产支付现金项目")
}
