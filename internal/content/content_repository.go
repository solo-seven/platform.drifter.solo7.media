package content

import "path/filepath"

type ContentRepository struct {
	gameMasterGuide *GameMasterGuideRepository
	playerGuide     *PlayerGuideRepository
	monsterManual   *MonsterManualRepository
	worldBooks      *WorldBookRepository
}

func NewContentRepository(contentPath string) *ContentRepository {
	return &ContentRepository{
		gameMasterGuide: NewGameMasterGuideRepository(filepath.Join(contentPath, "gamemaster_guide")),
		playerGuide:     NewPlayerGuideRepository(filepath.Join(contentPath, "player_guide")),
		monsterManual:   NewMonsterManualRepository(filepath.Join(contentPath, "monster_manual")),
		worldBooks:      NewWorldBookRepository(filepath.Join(contentPath, "world_books")),
	}
}
