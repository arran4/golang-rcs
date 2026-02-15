package diff

import (
	"fmt"
)

// Generate generates an EdDiff using the registered default algorithm.
func Generate(from []string, to []string) (EdDiff, error) {
	algo, err := DefaultAlgorithm()
	if err != nil {
		return nil, fmt.Errorf("getting default diff algorithm: %w", err)
	}
	return algo(from, to)
}
