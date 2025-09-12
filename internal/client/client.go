package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Message struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	ClientID  string      `json:"client_id"`
}

type Client struct {
	id          string
	writeQueue  chan Message
	readQueue   chan Message
	auditQueue  chan AuditEntry
	pendingReqs map[string]Message
	qp          *QueueProcessor
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewClient(id string, writeQueueSize, readQueueSize, auditQueueSize int, serverURL string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	writeQueue := make(chan Message, writeQueueSize)
	readQueue := make(chan Message, readQueueSize)
	auditQueue := make(chan AuditEntry, auditQueueSize)
	qp := NewQueueProcessor(writeQueue, readQueue, serverURL, ctx)
	qp.Start()
	return &Client{
		id:          id,
		writeQueue:  writeQueue,
		readQueue:   readQueue,
		auditQueue:  auditQueue,
		pendingReqs: make(map[string]Message),
		qp:          qp,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (c *Client) SendMessage(id string, payload interface{}) error {
	msg := Message{
		ID:        id,
		Type:      "request",
		Payload:   payload,
		Timestamp: time.Now(),
		ClientID:  c.id,
	}
	c.mu.Lock()
	c.pendingReqs[id] = msg
	c.mu.Unlock()

	select {
	case c.writeQueue <- msg:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
		return fmt.Errorf("write queue full")
	}
}

func (c *Client) StartReadProcessor() {
	go func() {
		for {
			select {
			case response := <-c.readQueue:
				c.ProcessResponse(response)
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

func (c *Client) ProcessResponse(response Message) {
	c.mu.Lock()
	request, exists := c.pendingReqs[response.ID]
	if exists {
		delete(c.pendingReqs, response.ID)
	}
	c.mu.Unlock()

	if exists {
		auditEntry := AuditEntry{
			RequestID:   request.ID,
			Request:     request,
			Response:    response,
			ProcessedAt: time.Now(),
		}

		select {
		case c.auditQueue <- auditEntry:
		case <-c.ctx.Done():
			return
		default:
			log.Printf("Audit queue full, dropping entry for request %s", response.ID)
		}
	}
}

func (c *Client) Cleanup() {
	c.cancel()
	close(c.writeQueue)
	close(c.readQueue)
	close(c.auditQueue)
}
