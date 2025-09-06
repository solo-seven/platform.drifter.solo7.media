package endtoend

import (
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/server"
)

func TestClientServerConnect(t *testing.T) {
	gameStateServer := server.NewGameStateServer()
	err := gameStateServer.Start("localhost:8080")
	if err != nil {
		t.Error(err)
	}
	err = gameStateServer.Stop()
	if err != nil {
		t.Error(err)
	}
}
