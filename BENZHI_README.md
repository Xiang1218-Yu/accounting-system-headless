# Accounting System · Benzhi 构建与运行说明

## 项目简介

Accounting System 是一个使用 JSON 文件作为本地持久化存储的 Go 会计核算系统。它提供科目管理、凭证录入、审核与过账、期间余额、结转、账簿查询以及资产负债表、利润表和现金流量表等服务层能力；同时包含基于 Fyne 的桌面端入口。

## 前置条件

- Go（以 `go.mod` 中声明的版本为准）
- Docker（仅在构建 Benzhi 评测镜像时需要）

## 标准构建命令

在项目根目录执行：

```bash
go build ./...
```

## 标准运行命令

启动桌面端入口：

```bash
go run ./cmd/accounting
```

## 标准测试命令

执行全部 Go 测试：

```bash
go test ./...
```

如需查看单个测试的详细输出，可使用：

```bash
go test -v ./...
```

## 构建 Benzhi Docker 镜像

`build_benzhi_docker.sh` 会使用项目根目录的 `benzhi.Dockerfile`，并默认构建 `linux/amd64` 镜像：

```bash
./build_benzhi_docker.sh accounting-system-headless
```

也可以显式指定镜像名和目标平台：

```bash
./build_benzhi_docker.sh accounting-system-headless linux/arm64
```

构建完成后，可进入容器执行构建、运行或测试命令：

```bash
docker run -it accounting-system-headless:latest
```
