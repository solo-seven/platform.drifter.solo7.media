package content

type ContentRepository struct {
	gameMasterGuide *GameMasterGuideRepository
	playerGuide     *PlayerGuideRepository
	monsterManual   *MonsterManualRepository
	worldBooks      *WorldBookRepository
}

func NewContentRepository() *ContentRepository {
	return &ContentRepository{}
}
