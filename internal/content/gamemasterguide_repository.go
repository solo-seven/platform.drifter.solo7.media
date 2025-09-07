package content

import "path/filepath"

type GameMasterGuideRepository struct {
	fileLocation string
}

func NewGameMasterGuideRepository(fileLocation string) *GameMasterGuideRepository {
	return &GameMasterGuideRepository{filepath.Join(fileLocation, "content.yaml")}
}
