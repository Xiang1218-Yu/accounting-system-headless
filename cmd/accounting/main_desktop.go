//go:build !linux

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"github.com/tog/accounting-system/internal/service"
	"github.com/tog/accounting-system/internal/store"
	"github.com/tog/accounting-system/internal/ui"
)

func main() {
	headless := flag.Bool("headless", false, "无界面 CLI 模式")
	datadir := flag.String("datadir", "", "数据目录（默认用户配置目录）")
	flag.Parse()

	dir := *datadir
	if dir == "" {
		dir = defaultDataDir()
	}

	st, err := store.New(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "初始化存储失败:", err)
		os.Exit(1)
	}
	svc := service.NewContainer(st)

	if _, err := svc.Seed.SeedIfEmpty(); err != nil {
		fmt.Fprintln(os.Stderr, "初始化科目失败:", err)
		os.Exit(1)
	}

	if *headless {
		runHeadless(svc, flag.Args())
		return
	}

	fyneApp := app.NewWithID("com.tog.accounting")
	a := ui.New(svc, fyneApp)
	a.Run()
}

func defaultDataDir() string {
	c, err := os.UserConfigDir()
	if err != nil || c == "" {
		c, _ = os.Getwd()
	}
	return filepath.Join(c, "accounting-system", "data")
}
