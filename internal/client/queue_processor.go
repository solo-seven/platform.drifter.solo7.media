package client

import (
	"context"
	"log"
	"time"
)

type QueueProcessor struct {
	writeQueue chan Message
	readQueue  chan Message
	serverURL  string
	ctx        context.Context
}

func NewQueueProcessor(writeQueue, readQueue chan Message, serverURL string, ctx context.Context) *QueueProcessor {
	return &QueueProcessor{
		writeQueue: writeQueue,
		readQueue:  readQueue,
		serverURL:  serverURL,
		ctx:        ctx,
	}
}

func (qp *QueueProcessor) Start() {
	go func() {
		for {
			select {
			case msg := <-qp.writeQueue:
				qp.processMessage(msg)
			case <-qp.ctx.Done():
				return
			}
		}
	}()
}

func (qp *QueueProcessor) processMessage(msg Message) {
	response := qp.sendToServer(msg)

	select {
	case qp.readQueue <- response:
	case <-qp.ctx.Done():
		return
	default:
		log.Printf("Read queue full, dropping response for request %s", msg.ID)
	}
}

func (qp *QueueProcessor) sendToServer(msg Message) Message {
	// TODO : Implement server connection
	return Message{
		ID:        msg.ID,
		Type:      "response",
		Payload:   msg.Payload,
		Timestamp: time.Now(),
		ClientID:  msg.ClientID,
	}
}
