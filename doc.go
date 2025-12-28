// Package rcs provides functionality for parsing and processing RCS (Revision Control System) files.
//
// The RCS file format is used by the RCS version control system to store the history of a file.
// This package allows you to parse these files into Go structs, inspect revision history,
// and access the content of specific revisions.
//
// The main entry point is the ParseFile function, which reads an RCS file from an io.Reader
// and returns a File struct containing the parsed data.
//
// Example usage:
//
//	f, err := os.Open("file.go,v")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer f.Close()
//
//	rcsFile, err := rcs.ParseFile(f)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Head Revision: %s\n", rcsFile.Head)
//	for _, rev := range rcsFile.RevisionHeads {
//		fmt.Printf("Revision: %s, Date: %s, Author: %s\n", rev.Revision, rev.Date, rev.Author)
//	}
package rcs
