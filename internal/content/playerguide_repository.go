package content

import "path/filepath"

type PlayerGuideRepository struct {
	fileLocation string
}

func NewPlayerGuideRepository(fileLocation string) *PlayerGuideRepository {
	return &PlayerGuideRepository{filepath.Join(fileLocation, "content.yaml")}
}
