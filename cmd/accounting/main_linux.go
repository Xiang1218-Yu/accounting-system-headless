//go:build linux

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tog/accounting-system/internal/service"
	"github.com/tog/accounting-system/internal/store"
)

func main() {
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
	runHeadless(svc, flag.Args())
}

func defaultDataDir() string {
	c, err := os.UserConfigDir()
	if err != nil || c == "" {
		c, _ = os.Getwd()
	}
	return filepath.Join(c, "accounting-system", "data")
}
