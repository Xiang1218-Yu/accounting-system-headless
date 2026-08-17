package service

import (
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// SeedService 初始化小企业会计准则科目表
type SeedService struct {
	st *store.Store
}

// NewSeedService 构造
func NewSeedService(st *store.Store) *SeedService {
	return &SeedService{st: st}
}

// SeedIfEmpty 首次启动初始化科目
func (s *SeedService) SeedIfEmpty() (bool, error) {
	if len(s.st.ListAccounts()) > 0 {
		return false, nil
	}
	return true, s.st.Mutate(func() {
		for _, a := range defaultChart() {
			s.st.SetAccount(a)
		}
	})
}

// defaultChart 小企业会计准则（2013）科目表
func defaultChart() []*domain.Account {
	type def struct {
		code, name string
		dir        domain.BalanceDirection
		aux        []domain.AuxType
	}
	defs := []def{
		{"1001", "库存现金", domain.DirDebit, nil},
		{"1002", "银行存款", domain.DirDebit, nil},
		{"1012", "其他货币资金", domain.DirDebit, nil},
		{"1101", "短期投资", domain.DirDebit, nil},
		{"1121", "应收票据", domain.DirDebit, nil},
		{"1122", "应收账款", domain.DirDebit, []domain.AuxType{domain.AuxCustomer}},
		{"1123", "预付账款", domain.DirDebit, []domain.AuxType{domain.AuxSupplier}},
		{"1131", "应收股利", domain.DirDebit, nil},
		{"1132", "应收利息", domain.DirDebit, nil},
		{"1221", "其他应收款", domain.DirDebit, nil},
		{"1231", "坏账准备", domain.DirCredit, nil},
		{"1401", "材料采购", domain.DirDebit, nil},
		{"1402", "在途物资", domain.DirDebit, nil},
		{"1403", "原材料", domain.DirDebit, nil},
		{"1404", "材料成本差异", domain.DirDebit, nil},
		{"1405", "库存商品", domain.DirDebit, nil},
		{"1407", "商品进销差价", domain.DirCredit, nil},
		{"1411", "周转材料", domain.DirDebit, nil},
		{"1501", "长期债券投资", domain.DirDebit, nil},
		{"1511", "长期股权投资", domain.DirDebit, nil},
		{"1601", "固定资产", domain.DirDebit, nil},
		{"1602", "累计折旧", domain.DirCredit, nil},
		{"1604", "在建工程", domain.DirDebit, nil},
		{"1605", "工程物资", domain.DirDebit, nil},
		{"1606", "固定资产清理", domain.DirDebit, nil},
		{"1701", "无形资产", domain.DirDebit, nil},
		{"1702", "累计摊销", domain.DirCredit, nil},
		{"1801", "长期待摊费用", domain.DirDebit, nil},
		{"1901", "待处理财产损溢", domain.DirDebit, nil},
		{"2001", "短期借款", domain.DirCredit, nil},
		{"2201", "应付票据", domain.DirCredit, nil},
		{"2202", "应付账款", domain.DirCredit, []domain.AuxType{domain.AuxSupplier}},
		{"2203", "预收账款", domain.DirCredit, []domain.AuxType{domain.AuxCustomer}},
		{"2211", "应付职工薪酬", domain.DirCredit, nil},
		{"2221", "应交税费", domain.DirCredit, nil},
		{"2231", "应付利息", domain.DirCredit, nil},
		{"2232", "应付利润", domain.DirCredit, nil},
		{"2241", "其他应付款", domain.DirCredit, nil},
		{"2501", "长期借款", domain.DirCredit, nil},
		{"2502", "长期应付款", domain.DirCredit, nil},
		{"3001", "实收资本", domain.DirCredit, nil},
		{"3002", "资本公积", domain.DirCredit, nil},
		{"3101", "盈余公积", domain.DirCredit, nil},
		{"3103", "本年利润", domain.DirCredit, nil},
		{"3104", "利润分配", domain.DirCredit, nil},
		{"4001", "生产成本", domain.DirDebit, []domain.AuxType{domain.AuxDepartment, domain.AuxProject}},
		{"4101", "制造费用", domain.DirDebit, []domain.AuxType{domain.AuxDepartment}},
		{"4301", "研发支出", domain.DirDebit, nil},
		{"5001", "主营业务收入", domain.DirCredit, []domain.AuxType{domain.AuxCustomer, domain.AuxDepartment}},
		{"5051", "其他业务收入", domain.DirCredit, nil},
		{"5101", "投资收益", domain.DirCredit, nil},
		{"5301", "营业外收入", domain.DirCredit, nil},
		{"5401", "主营业务成本", domain.DirDebit, nil},
		{"5402", "其他业务成本", domain.DirDebit, nil},
		{"5403", "营业税金及附加", domain.DirDebit, nil},
		{"5601", "销售费用", domain.DirDebit, []domain.AuxType{domain.AuxDepartment}},
		{"5602", "管理费用", domain.DirDebit, []domain.AuxType{domain.AuxDepartment}},
		{"5603", "财务费用", domain.DirDebit, nil},
		{"5711", "营业外支出", domain.DirDebit, nil},
		{"5801", "所得税费用", domain.DirDebit, nil},
	}
	out := make([]*domain.Account, 0, len(defs))
	for _, d := range defs {
		a := &domain.Account{
			Code:      d.code,
			Name:      d.name,
			Category:  domain.CategoryByCode(d.code),
			Direction: d.dir,
			Level:     1,
			IsLeaf:    true,
			AuxTypes:  d.aux,
			IsActive:  true,
		}
		out = append(out, a)
	}
	return out
}
