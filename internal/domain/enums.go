package domain

// Category 科目类别
type Category string

const (
	CategoryAsset      Category = "资产"
	CategoryLiability  Category = "负债"
	CategoryEquity     Category = "权益"
	CategoryCost       Category = "成本"
	CategoryProfitLoss Category = "损益"
)

// BalanceDirection 余额方向
type BalanceDirection string

const (
	DirDebit  BalanceDirection = "借"
	DirCredit BalanceDirection = "贷"
)

// VoucherStatus 凭证状态
type VoucherStatus string

const (
	StatusDraft   VoucherStatus = "草稿"
	StatusAudited VoucherStatus = "已审核"
	StatusPosted  VoucherStatus = "已过账"
)

// VoucherType 凭证类型
type VoucherType string

const (
	VoucherNormal  VoucherType = "普通"
	VoucherClosing VoucherType = "结转损益"
	VoucherFX      VoucherType = "期末调汇"
)

// PeriodStatus 会计期间状态
type PeriodStatus string

const (
	PeriodOpen   PeriodStatus = "开放"
	PeriodClosed PeriodStatus = "已结账"
)

// AuxType 辅助核算类型
type AuxType string

const (
	AuxCustomer   AuxType = "客户"
	AuxSupplier   AuxType = "供应商"
	AuxDepartment AuxType = "部门"
	AuxProject    AuxType = "项目"
)

// AllAuxTypes 全部辅助核算类型
var AllAuxTypes = []AuxType{AuxCustomer, AuxSupplier, AuxDepartment, AuxProject}

// CategoryByCode 根据科目编码前缀推断类别（小企业会计准则）
func CategoryByCode(code string) Category {
	if len(code) == 0 {
		return CategoryAsset
	}
	switch code[:1] {
	case "1":
		return CategoryAsset
	case "2":
		return CategoryLiability
	case "3":
		return CategoryEquity
	case "4":
		return CategoryCost
	case "5":
		return CategoryProfitLoss
	}
	return CategoryAsset
}
