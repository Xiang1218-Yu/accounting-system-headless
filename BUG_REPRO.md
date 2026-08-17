# 缺陷复现说明

## 缺陷现象
财务同事在科目和凭证详情页只是改了本地展示数据，后面再打开却发现账套里的原数据也被悄悄改掉了，连已经设置好的汇率快照都会跟着变。请修复这种读取或传参后污染已保存数据的问题。

## 触发方式
# 1. 进入测试项目根目录；预期：后续命令在目标项目内执行。
cd /workplace/accounting-system-headless
# 2. target-verify 检查读取结果和传入引用不会污染已保存状态；预期：未修复时断言失败。
go test -v ./internal/service -run '^TestReadModelsDoNotMutatePersistedState$' -count=20
# 3. 修复后运行完整测试集；预期：服务与存储相关回归均通过。
go test ./...

## 触发后的实际错误输出

```text
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
=== RUN   TestReadModelsDoNotMutatePersistedState
    store_snapshot_test.go:17: 读取科目后修改返回值污染了存储: &domain.Account{Code:"1002", Name:"被外部覆盖", Category:"资产", Direction:"借", Level:1, ParentCode:"", IsLeaf:true, IsForeignCurrency:false, Currency:"", AuxTypes:[]domain.AuxType{"客户"}, IsActive:true}
--- FAIL: TestReadModelsDoNotMutatePersistedState (0.00s)
FAIL
FAIL	github.com/tog/accounting-system/internal/service	0.030s
FAIL
```
