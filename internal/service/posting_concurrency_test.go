package service

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tog/accounting-system/internal/domain"
)

func TestConcurrentPostAppliesVoucherOnlyOnce(t *testing.T) {
	svc := newTestContainer(t)
	voucher := &domain.Voucher{VoucherDate: parseT("2026-08-09"), Type: domain.VoucherNormal, Entries: []domain.VoucherEntry{
		{AccountCode: "1002", Debit: decimal.NewFromInt(480)},
		{AccountCode: "3001", Credit: decimal.NewFromInt(480)},
	}}
	if err := svc.Voucher.Create(voucher, "maker"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.Voucher.Approve(voucher.ID, "auditor"); err != nil {
		t.Fatalf("approve: %v", err)
	}

	const callers = 32
	start := make(chan struct{})
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- svc.Engine.Post(voucher.ID, "poster")
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("同一张凭证并发过账只能成功一次，实际成功 %d 次", successes)
	}

	balance := svc.Store.GetBalance(2026, 8, "1002", "")
	if balance == nil || !balance.PeriodDebit.Equal(decimal.NewFromInt(480)) {
		t.Fatalf("并发过账重复累计了余额: %#v", balance)
	}
}
