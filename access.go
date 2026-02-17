package rcs

func (f *File) CopyAccessList(from *File) {
	f.Access = from.Access
	if from.AccessUsers != nil {
		f.AccessUsers = make([]string, len(from.AccessUsers))
		copy(f.AccessUsers, from.AccessUsers)
	} else {
		f.AccessUsers = nil
	}
}
