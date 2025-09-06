package player_stories

import (
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/server"
)

func TestPlayerLogin(m *testing.T) {
	gameStateServer := server.NewGameStateServer()
	err := gameStateServer.Start("localhost:8080")
	if err != nil {
		m.Error(err)
	}
	err = gameStateServer.Stop()
	if err != nil {
		m.Error(err)
	}
}
