package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tog/accounting-system/internal/service"
)

// runHeadless 无界面 CLI
func runHeadless(svc *service.Container, args []string) {
	if len(args) == 0 {
		printUsage()
		return
	}
	switch args[0] {
	case "seed":
		ok, err := svc.Seed.SeedIfEmpty()
		if err != nil {
			fmt.Fprintln(os.Stderr, "err:", err)
			os.Exit(1)
		}
		if ok {
			fmt.Println("已初始化小企业会计准则科目表")
		} else {
			fmt.Println("科目已存在，跳过初始化")
		}
	case "verify":
		year, month := svc.Period.Current()
		if year == 0 {
			year, month = 2026, 1
		}
		r := svc.Ledger.Verify(year, month)
		fmt.Println("一致性自检结果:", passStr(r.OK))
		for _, c := range r.Checks {
			fmt.Println("  ✓", c)
		}
		for _, e := range r.Errors {
			fmt.Println("  ✗", e)
		}
		if !r.OK {
			os.Exit(1)
		}
	case "report":
		reportCmd(svc, args[1:])
	case "backup":
		backupCmd(svc, args[1:])
	case "restore":
		restoreCmd(svc, args[1:])
	case "demo":
		demoCmd(svc)
	case "vouchers":
		for _, v := range svc.Voucher.List() {
			fmt.Printf("%s  %s  %s  借:%s 贷:%s\n", v.Number, v.VoucherDate.Format("2006-01-02"), v.Status, v.TotalDebit().StringFixed(2), v.TotalCredit().StringFixed(2))
		}
	default:
		printUsage()
	}
}

func reportCmd(svc *service.Container, args []string) {
	typ := "balance"
	period := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--type" && i+1 < len(args) {
			typ = args[i+1]
			i++
		}
		if args[i] == "--period" && i+1 < len(args) {
			period = args[i+1]
			i++
		}
	}
	year, month := parsePeriod(period)
	switch typ {
	case "balance":
		bs := svc.Ledger.BuildBalanceSheet(year, month)
		fmt.Printf("资产负债表 %s  资产=%s 负债=%s 权益=%s 平衡=%s\n", bs.Period, bs.AssetTotal.StringFixed(2), bs.LiabTotal.StringFixed(2), bs.EquityTotal.StringFixed(2), passStr(bs.Balanced))
	case "income":
		is := svc.Ledger.BuildIncomeStatement(year, month)
		fmt.Printf("利润表 %s  净利润=%s\n", is.Period, is.NetProfit.StringFixed(2))
		for _, r := range is.Rows {
			fmt.Printf("  %-16s 本期:%s 累计:%s\n", r.Name, r.Current.StringFixed(2), r.YTD.StringFixed(2))
		}
	case "cashflow":
		cf := svc.Ledger.BuildCashFlow(year, month)
		fmt.Printf("现金流量表 %s  净增加=%s 期末=%s 对账=%s\n", cf.Period, cf.NetIncrease.StringFixed(2), cf.CashEnd.StringFixed(2), passStr(cf.Reconciled))
	case "trial":
		rows := svc.Ledger.TrialBalance(year, month)
		_, _, _, _, cd, cc := service.BalanceTotals(rows)
		fmt.Printf("余额表 %04d-%02d  共%d行  期末借=%s 贷=%s 平衡=%s\n", year, month, len(rows), cd.StringFixed(2), cc.StringFixed(2), passStr(cd.Equal(cc)))
	default:
		fmt.Println("未知报表类型:", typ)
	}
}

func backupCmd(svc *service.Container, args []string) {
	out := "accounting-backup.zip"
	if len(args) > 0 {
		out = args[0]
	}
	if err := svc.Backup.Backup(out); err != nil {
		fmt.Fprintln(os.Stderr, "err:", err)
		os.Exit(1)
	}
	fmt.Println("已备份到", out)
}

func restoreCmd(svc *service.Container, args []string) {
	src := "accounting-backup.zip"
	if len(args) > 0 {
		src = args[0]
	}
	if err := svc.Backup.Restore(src); err != nil {
		fmt.Fprintln(os.Stderr, "err:", err)
		os.Exit(1)
	}
	fmt.Println("已从", src, "恢复，请重启应用以重载数据")
}

func parsePeriod(p string) (int, int) {
	if p == "" {
		return 2026, 1
	}
	parts := strings.Split(p, "-")
	if len(parts) != 2 {
		return 2026, 1
	}
	y, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	if y == 0 {
		y = 2026
	}
	if m == 0 {
		m = 1
	}
	return y, m
}

func passStr(ok bool) string {
	if ok {
		return "通过"
	}
	return "未通过"
}

func printUsage() {
	fmt.Println("会计核算系统 CLI")
	fmt.Println("用法: accounting --headless <命令>")
	fmt.Println("命令:")
	fmt.Println("  seed                 初始化科目表")
	fmt.Println("  verify               一致性自检")
	fmt.Println("  report --type <balance|income|cashflow|trial> --period YYYY-MM")
	fmt.Println("  backup [文件名.zip]   备份数据")
	fmt.Println("  restore [文件名.zip]  从备份恢复")
	fmt.Println("  demo                 生成演示凭证并过账")
	fmt.Println("  vouchers             列出凭证")
}
