package service

import "github.com/tog/accounting-system/internal/store"

// Container 聚合全部业务服务，便于在 UI/CLI 间共享
type Container struct {
	Store   *store.Store
	Audit   *AuditService
	Seed    *SeedService
	Account *AccountService
	Voucher *VoucherService
	Engine  *PostingEngine
	Ledger  *LedgerService
	Aux     *AuxService
	Period  *PeriodService
	Closing *ClosingService
	FX      *FXService
	Excel   *ExcelService
	Backup  *BackupService
}

// NewContainer 构造服务容器
func NewContainer(st *store.Store) *Container {
	audit := NewAuditService(st)
	engine := NewPostingEngine(st, audit)
	ledger := NewLedgerService(st)
	return &Container{
		Store:   st,
		Audit:   audit,
		Seed:    NewSeedService(st),
		Account: NewAccountService(st, audit),
		Voucher: NewVoucherService(st, audit, engine),
		Engine:  engine,
		Ledger:  ledger,
		Aux:     NewAuxService(st, audit),
		Period:  NewPeriodService(st, audit),
		Closing: NewClosingService(st, ledger, engine, audit),
		FX:      NewFXService(st, ledger, engine, audit),
		Excel:   NewExcelService(st),
		Backup:  NewBackupService(st),
	}
}
