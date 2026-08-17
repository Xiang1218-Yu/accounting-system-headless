package service

import (
	"github.com/tog/accounting-system/internal/store"
)

// BackupService 数据备份与恢复
type BackupService struct {
	st *store.Store
}

// NewBackupService 构造
func NewBackupService(st *store.Store) *BackupService {
	return &BackupService{st: st}
}

// Backup 备份全部数据到 zip 文件
func (b *BackupService) Backup(dest string) error {
	return b.st.Backup(dest)
}

// Restore 从 zip 恢复（恢复后需重新 New Store 重载）
func (b *BackupService) Restore(zipPath string) error {
	return store.Restore(zipPath, b.st.Dir())
}
