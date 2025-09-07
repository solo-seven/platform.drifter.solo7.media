package content

import (
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/content"
)

func TestFullSimpleContentLoading(t *testing.T) {
	contentRepository := content.NewContentRepository("./examples/simple/")
	if contentRepository == nil {
		t.Error("Content repository is nil")
	} else {
		gameMasterGuide := contentRepository.GameMasterGuide()
		if gameMasterGuide == nil {
			t.Error("Game master guide is nil")
		} else {
			t.Log("Verify the Game Master Guide Here")
		}
		playerGuide := contentRepository.PlayerGuide()
		if playerGuide == nil {
			t.Error("Player guide is nil")
		} else {
			t.Log("Verify the Player Guide Here")
		}
		monsterManual := contentRepository.MonsterManual()
		if monsterManual == nil {
			t.Error("Monster manual is nil")
		} else {
			t.Log("Verify the Monster Manual Here")
		}
		worldBooks := contentRepository.WorldBooks()
		if worldBooks == nil {
			t.Error("World books is nil")
		} else {
			t.Log("Verify the World Books Here")
		}
	}
}
