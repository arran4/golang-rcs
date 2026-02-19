package rcs

import "fmt"

// SetState updates the state for the specified revision.
func (f *File) SetState(revision string, state string) error {
	for _, rh := range f.RevisionHeads {
		if rh.Revision.String() == revision {
			rh.State = ID(state)
			return nil
		}
	}
	return fmt.Errorf("revision %s not found", revision)
}

// GetState retrieves the state for the specified revision.
func (f *File) GetState(revision string) (string, error) {
	for _, rh := range f.RevisionHeads {
		if rh.Revision.String() == revision {
			return string(rh.State), nil
		}
	}
	return "", fmt.Errorf("revision %s not found", revision)
}

// StateEntry represents a state entry for a revision.
type StateEntry struct {
	Revision string
	State    string
}

// ListStates retrieves all states in the RCS file.
func (f *File) ListStates() []StateEntry {
	var states []StateEntry
	for _, rh := range f.RevisionHeads {
		states = append(states, StateEntry{
			Revision: rh.Revision.String(),
			State:    string(rh.State),
		})
	}
	return states
}
