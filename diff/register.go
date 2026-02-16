package diff

import (
	"fmt"
)

// DiffAlgorithm generates an EdDiff from two line slices.
type DiffAlgorithm func(from []string, to []string) (EdDiff, error)

var (
	registry    = make(map[string]DiffAlgorithm)
	defaultAlgo string
)

// Register registers a diff algorithm with a name.
func Register(name string, algo DiffAlgorithm) {
	registry[name] = algo
}

// GetAlgorithm returns the registered algorithm by name.
func GetAlgorithm(name string) (DiffAlgorithm, error) {
	if algo, ok := registry[name]; ok {
		return algo, nil
	}
	return nil, fmt.Errorf("diff algorithm %q not found", name)
}

// DefaultAlgorithm returns the default algorithm (the first registered one).
func DefaultAlgorithm() (DiffAlgorithm, error) {
	if defaultAlgo == "" {
		return nil, fmt.Errorf("no diff algorithms registered")
	}
	return GetAlgorithm(defaultAlgo)
}

// SetDefaultAlgorithm sets the default algorithm by name.
func SetDefaultAlgorithm(name string) error {
	if _, ok := registry[name]; !ok {
		return fmt.Errorf("diff algorithm %q not found", name)
	}
	defaultAlgo = name
	return nil
}

// ListAlgorithms returns a list of registered algorithm names.
func ListAlgorithms() []string {
	var keys []string
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}
