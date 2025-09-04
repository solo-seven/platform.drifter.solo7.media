package server

import (
	"context"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/dsl"
)

// DSLInterpreter implements domain.DSLInterpreter
type DSLInterpreter struct {
	interpreter *dsl.DSLInterpreter
}

// NewDSLInterpreter creates a new DSL interpreter
func NewDSLInterpreter(logger domain.Logger) *DSLInterpreter {
	return &DSLInterpreter{
		interpreter: dsl.NewDSLInterpreter(logger),
	}
}

// LoadWorld loads a complete world from DSL files
func (di *DSLInterpreter) LoadWorld(ctx context.Context, contentPath string) (*domain.World, error) {
	return di.interpreter.LoadWorld(ctx, contentPath)
}

// ValidateContent validates DSL content
func (di *DSLInterpreter) ValidateContent(ctx context.Context, contentPath string) error {
	return di.interpreter.ValidateContent(ctx, contentPath)
}

// ReloadContent reloads DSL content
func (di *DSLInterpreter) ReloadContent(ctx context.Context, contentPath string) (*domain.World, error) {
	return di.interpreter.ReloadContent(ctx, contentPath)
}
