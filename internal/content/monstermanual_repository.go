package content

import "path/filepath"

type MonsterManualRepository struct {
	fileLocation string
}

func NewMonsterManualRepository(fileLocation string) *MonsterManualRepository {
	return &MonsterManualRepository{filepath.Join(fileLocation, "content.yaml")}
}
