package content

import (
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/content"
)

func TestFullSimpleContentLoading(m *testing.T) {
	contentRepository := content.NewContentRepository()
	if contentRepository == nil {
		m.Error("Content repository is nil")
	}
}
