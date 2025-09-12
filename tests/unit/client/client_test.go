package client

import (
	"testing"

	client2 "github.com/solo-seven/platform.drifter.solo7.media/internal/client"
)

func TestSimpleClientMessage(t *testing.T) {
	client := client2.NewClient("test", 5, 5, 5, "localhost:8080")
	if client == nil {
		t.Error("Client is nil")
		return
	}
	client.StartReadProcessor()
}
