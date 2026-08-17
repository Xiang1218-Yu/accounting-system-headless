package domain

import "time"

// AuditLog 操作审计日志
type AuditLog struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Before    string    `json:"before,omitempty"`
	After     string    `json:"after,omitempty"`
	OpID      string    `json:"op_id,omitempty"`
}
