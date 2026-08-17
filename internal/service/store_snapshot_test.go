package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

func TestReadModelsDoNotMutatePersistedState(t *testing.T) {
	svc := newTestContainer(t)

	account := svc.Account.Get("1002")
	account.Name = "被外部覆盖"
	account.AuxTypes = append(account.AuxTypes, domain.AuxCustomer)
	if got := svc.Account.Get("1002"); got.Name == "被外部覆盖" || len(got.AuxTypes) != 0 {
		t.Fatalf("读取科目后修改返回值污染了存储: %#v", got)
	}

	rates := map[string]float64{"USD": 7.10}
	if err := svc.Period.SetFXRate(2026, 8, rates, "tester"); err != nil {
		t.Fatalf("set rates: %v", err)
	}
	rates["USD"] = 9.99
	if got := svc.Period.Get(2026, 8).FXRateSnapshot["USD"]; got != 7.10 {
		t.Fatalf("调用方修改汇率 map 污染了期间快照: got %v", got)
	}

	voucher := &domain.Voucher{VoucherDate: parseT("2026-08-08"), Type: domain.VoucherNormal, Entries: []domain.VoucherEntry{
		{AccountCode: "1002", Debit: decimal.NewFromInt(200)},
		{AccountCode: "3001", Credit: decimal.NewFromInt(200)},
	}}
	if err := svc.Voucher.Create(voucher, "tester"); err != nil {
		t.Fatalf("create: %v", err)
	}
	voucher.Entries[0].Debit = decimal.NewFromInt(999)
	stored := svc.Voucher.Get(voucher.ID)
	if !stored.Entries[0].Debit.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("调用方修改凭证入参污染了存储: %#v", stored.Entries[0])
	}

	listed := svc.Voucher.List()
	listed[0].Entries[0].Debit = decimal.Zero
	if svc.Voucher.Get(voucher.ID).Entries[0].Debit.IsZero() {
		t.Fatal("修改列表返回的嵌套分录污染了存储")
	}
}
