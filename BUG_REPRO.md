# 非现金固定资产购置现金流复现

## Bug 是什么

现金流量表在统计购建长期资产支付的现金时，只根据固定资产等业务科目判断，未确认整张凭证是否实际包含现金科目分录。因此，借记固定资产、贷记应付账款的非现金购置也被当作现金流出，导致现金流量虚增。

## 如何触发

分别录入一张现金购置固定资产的凭证和一张仅挂应付账款的固定资产购置凭证，再生成现金流量表。后一张凭证没有现金分录，却仍被累计到购建长期资产支付的现金项目中。

## 运行指令

```bash
go test ./internal/service -run '^TestCashFlowExcludesNonCashAssetPurchase$' -count=1
```

## 错误信息

测试要求不含现金分录的应付购置不得计入现金流出，但报表结果把两张购置凭证都累计，金额为 200。

## 错误堆栈

```text
--- FAIL: TestCashFlowExcludesNonCashAssetPurchase (0.10s)
    cash_flow_non_cash_purchase_test.go:44: 应付购置没有现金分录，不应计入购建长期资产支付现金，实际 200
FAIL
FAIL	github.com/tog/accounting-system/internal/service	0.603s
FAIL
```
