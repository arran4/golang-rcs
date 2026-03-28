package diff

import (
	"fmt"
)

// Generate generates an EdDiff using the registered default algorithm.
func Generate(from []string, to []string) (EdDiff, error) {
	if defaultAlgo != "" {
		algo, err := DefaultAlgorithm()
		if err != nil {
			return nil, fmt.Errorf("getting default diff algorithm: %w", err)
		}
		return algo(from, to)
	}

	// Fallback to explicitly fetching "znkr" as the new default algorithm,
	// otherwise fall back to whatever is first available if znkr isn't registered for some reason.
	algo, err := GetAlgorithm("znkr")
	if err == nil {
		return algo(from, to)
	}

	return GenerateEdDiffFromLines(from, to)
}
