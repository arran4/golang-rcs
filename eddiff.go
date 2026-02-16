package rcs

import (
	"github.com/arran4/golang-rcs/diff"
)

// Type aliases for backward compatibility if needed, or just to expose them in rcs package.
// Ideally, we should switch the codebase to use diff package types, but rcs package is likely the API surface.
// Let's alias them.

type EdDiff = diff.EdDiff
type EdDiffCommand = diff.EdDiffCommand
type Delete = diff.Delete
type Add = diff.Add
type LineReader = diff.LineReader
type LineWriter = diff.LineWriter

// GenerateEdDiffFromLines delegates to the diff package.
func GenerateEdDiffFromLines(from []string, to []string) (EdDiff, error) {
	return diff.Generate(from, to)
}

// ParseEdDiff delegates to the diff package.
var ParseEdDiff = diff.ParseEdDiff
