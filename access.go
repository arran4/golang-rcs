package rcs

// CopyAccessList copies the access list from the source file to the receiver file.
func (f *File) CopyAccessList(from *File) {
	f.Access = from.Access
	f.AccessMultiline = from.AccessMultiline
	if from.AccessUsers != nil {
		f.AccessUsers = make([]string, len(from.AccessUsers))
		copy(f.AccessUsers, from.AccessUsers)
	} else {
		f.AccessUsers = nil
	}
}

// AppendAccessList appends the access list from the source file to the receiver file,
// avoiding duplicates.
func (f *File) AppendAccessList(from *File) {
	if !from.Access {
		return
	}
	f.Access = true
	// If the source has multiline access, we switch to multiline to accommodate potentially large lists.
	// Or should we? The requirement is to support both.
	// Usually appending implies growth, so preserving the "wider" format (multiline) seems safer if either is multiline.
	if from.AccessMultiline {
		f.AccessMultiline = true
	}

	seen := make(map[string]bool)
	for _, user := range f.AccessUsers {
		seen[user] = true
	}

	for _, user := range from.AccessUsers {
		if !seen[user] {
			f.AccessUsers = append(f.AccessUsers, user)
			seen[user] = true
		}
	}
}
