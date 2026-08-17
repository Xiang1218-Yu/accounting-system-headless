package service

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/tog/accounting-system/internal/domain"
	"github.com/tog/accounting-system/internal/store"
)

// AuditService 操作审计
type AuditService struct {
	st *store.Store
}

// NewAuditService 构造
func NewAuditService(st *store.Store) *AuditService {
	return &AuditService{st: st}
}

// Log 记录审计日志（须在 store.Mutate 内调用）
func (a *AuditService) Log(user, action, entity, entityID, opID, before, after string) {
	a.st.AppendAudit(domain.AuditLog{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		User:      user,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Before:    before,
		After:     after,
		OpID:      opID,
	})
}

// NewOpID 生成一次业务操作的关联 ID
func NewOpID() string { return uuid.NewString() }

// jsonOf 序列化为 JSON 摘要
func jsonOf(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
