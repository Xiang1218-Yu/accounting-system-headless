# 凭证导入残缺草稿复现

## Bug 是什么

导入凭证表时，同一张凭证中只要存在一条无效科目分录，当前实现会忽略该无效行，仍把其余有效分录保存为草稿并返回成功。这会使一张本应整体导入失败的凭证留下残缺数据。

## 如何触发

准备一张包含多条分录的凭证，其中至少一条分录引用不存在的科目，然后执行导入。导入接口没有拒绝整张凭证，测试可观察到仍有一张凭证被保存。

## 运行指令

```bash
go test ./internal/service -run '^TestImportVouchersRejectsPartialVoucherWhenEntryIsInvalid$' -count=1
```

## 错误信息

测试断言要求包含无效分录的凭证导入失败，但实际导入结果仍为 1 张凭证。

## 错误堆栈

```text
--- FAIL: TestImportVouchersRejectsPartialVoucherWhenEntryIsInvalid (0.02s)
    import_vouchers_atomicity_test.go:33: 包含无效分录的凭证导入应失败，实际导入 1 张
FAIL
FAIL	github.com/tog/accounting-system/internal/service	0.809s
FAIL
```
