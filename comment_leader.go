package rcs

// SetComment sets the comment leader for the RCS file.
func (f *File) SetComment(comment string) {
	f.Comment = comment
}

// GetComment returns the comment leader for the RCS file.
func (f *File) GetComment() string {
	return f.Comment
}
