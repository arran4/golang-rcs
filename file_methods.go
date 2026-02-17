package rcs

// SetLock sets a lock for the user on the given revision.
// Returns true if the lock was set or updated, false if it already existed identically.
func (file *File) SetLock(user, revision string) bool {
	for _, l := range file.Locks {
		if l.User == user {
			if l.Revision == revision {
				return false
			}
			l.Revision = revision
			return true
		}
	}
	file.Locks = append(file.Locks, &Lock{User: user, Revision: revision})
	return true
}

// ClearLock removes the lock for the user on the given revision.
// Returns true if a lock was removed.
func (file *File) ClearLock(user, revision string) bool {
	out := file.Locks[:0]
	changed := false
	for _, l := range file.Locks {
		if l.User == user && l.Revision == revision {
			changed = true
			continue
		}
		out = append(out, l)
	}
	file.Locks = out
	return changed
}

// AddAccess adds users to the access list.
func (file *File) AddAccess(users []string) {
	if len(users) == 0 {
		return
	}
	file.Access = true
	existing := make(map[string]bool)
	for _, u := range file.AccessUsers {
		existing[u] = true
	}
	for _, u := range users {
		if !existing[u] {
			file.AccessUsers = append(file.AccessUsers, u)
			existing[u] = true
		}
	}
}

// RemoveAccess removes users from the access list.
func (file *File) RemoveAccess(users []string) {
	if len(users) == 0 {
		return
	}
	toRemove := make(map[string]bool)
	for _, u := range users {
		toRemove[u] = true
	}
	out := file.AccessUsers[:0]
	for _, u := range file.AccessUsers {
		if !toRemove[u] {
			out = append(out, u)
		}
	}
	file.AccessUsers = out
}

// RemoveAllAccess removes all users from the access list.
func (file *File) RemoveAllAccess() {
	file.AccessUsers = nil
	file.Access = true
}

// DeleteRevision deletes a revision or range of revisions.
// Currently only supports simple deletion logging or stub.
// Implementing full revision deletion is complex as it involves fixing up the delta tree.
func (file *File) DeleteRevision(revision string) error {
	// TODO: Implement revision deletion logic.
	// For now, this is a placeholder to allow verifying argument parsing.
	return nil
}
