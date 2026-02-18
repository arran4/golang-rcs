package rcs

// CopyAccessList copies the access list from the source file to the receiver file.
func (f *File) CopyAccessList(from *File) {
	f.Access = from.Access
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
	if from.AccessUsers == nil {
		return
	}
	f.Access = true // Ensure access list exists if we are appending
	if f.AccessUsers == nil {
		f.AccessUsers = make([]string, 0, len(from.AccessUsers))
	}

	existing := make(map[string]bool)
	for _, user := range f.AccessUsers {
		existing[user] = true
	}

	for _, user := range from.AccessUsers {
		if !existing[user] {
			f.AccessUsers = append(f.AccessUsers, user)
			existing[user] = true
		}
	}
}
