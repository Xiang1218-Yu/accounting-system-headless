# Accounting System

一个使用 JSON 文件存储数据的本地会计核算系统。项目提供凭证录入、审核与过账、期间余额、结转、账簿和三大报表等服务层能力，并包含 Fyne 桌面端入口。

## 本地验证

```bash
go build ./...
go test ./...
```

## 运行

```bash
go run ./cmd/accounting
```

## Linux 容器构建

Linux 环境默认编译无界面 CLI 入口，桌面端 Fyne 代码仅在非 Linux 平台参与构建，因此容器可以直接执行：

```bash
go build ./...
go test ./...
```
