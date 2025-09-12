package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AuditEntry struct {
	RequestID   string    `json:"request_id"`
	Request     Message   `json:"request"`
	Response    Message   `json:"response"`
	ProcessedAt time.Time `json:"processed_at"`
}

type AuditLogger struct {
	auditQueue chan AuditEntry
	logFile    string
	ctx        context.Context
}

func NewAuditLogger(auditQueue chan AuditEntry, logFile string, ctx context.Context) *AuditLogger {
	return &AuditLogger{
		auditQueue: auditQueue,
		logFile:    logFile,
		ctx:        ctx,
	}
}

func (al *AuditLogger) Start() {
	go func() {
		for {
			select {
			case entry := <-al.auditQueue:
				al.writeAuditEntry(entry)
			case <-al.ctx.Done():
				return
			}
		}
	}()
}

func (al *AuditLogger) writeAuditEntry(entry AuditEntry) {
	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("AUDIT: %s\n", data)
}
