package dsl

import (
	"context"
	"strings"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// MarkdownParser implements the domain.MarkdownParser interface
type MarkdownParser struct {
	logger domain.Logger
}

// NewMarkdownParser creates a new MarkdownParser
func NewMarkdownParser(logger domain.Logger) *MarkdownParser {
	return &MarkdownParser{logger: logger}
}

// ParseMarkdown parses a Markdown string into an AST
func (mp *MarkdownParser) ParseMarkdown(ctx context.Context, markdown string) (domain.MarkdownAST, error) {
	// Simple implementation that just returns the markdown as-is
	return &simpleMarkdownAST{content: markdown}, nil
}

// simpleMarkdownAST is a simple implementation of MarkdownAST
type simpleMarkdownAST struct {
	content string
}

// RenderToText renders the AST to plain text
func (sma *simpleMarkdownAST) RenderToText(ctx context.Context) (string, error) {
	// Simple implementation that strips markdown formatting
	content := sma.content
	// Remove markdown headers
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		// Remove # headers
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		}
		// Remove ** bold markers
		line = strings.ReplaceAll(line, "**", "")
		// Remove * italic markers
		line = strings.ReplaceAll(line, "*", "")
		result = append(result, line)
	}
	return strings.Join(result, "\n"), nil
}
