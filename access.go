package rcs

// CopyAccessList copies the access list from the source file to the current file.
// It replaces the current access list with the one from the source file.
func (f *File) CopyAccessList(from *File) {
	f.Access = from.Access
	f.AccessInline = from.AccessInline
	if from.AccessUsers != nil {
		f.AccessUsers = make([]string, len(from.AccessUsers))
		copy(f.AccessUsers, from.AccessUsers)
	} else {
		f.AccessUsers = nil
	}
}
