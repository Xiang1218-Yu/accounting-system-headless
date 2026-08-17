# 缺陷复现说明

## 缺陷现象
客户发来的备份包导入后，数据目录外竟然多出了一份 JSON 文件。这个入口会直接处理外部 zip，担心有人能借它覆盖别的业务文件；请修复恢复备份时的越界写入问题。

## 触发方式
# 1. 进入测试项目根目录；预期：后续命令在目标项目内执行。
cd /workplace/accounting-system-headless
# 2. target-verify 检查外部 zip 条目不能逃出数据目录；预期：未修复时越界文件断言失败。
go test -v ./internal/store -run '^TestRestoreRejectsArchiveEntriesOutsideDataDirectory$' -count=20
# 3. 修复后运行完整测试集；预期：备份恢复与其余功能均通过。
go test ./...

## 触发后的实际错误输出

```text
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
=== RUN   TestRestoreRejectsArchiveEntriesOutsideDataDirectory
    backup_restore_security_test.go:35: 恢复包内含越界路径时应被拒绝
--- FAIL: TestRestoreRejectsArchiveEntriesOutsideDataDirectory (0.00s)
FAIL
FAIL	github.com/tog/accounting-system/internal/store	0.007s
FAIL
```
