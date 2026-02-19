package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/arran4/golang-rcs"
)

// CommentLeader processes the list of files and prints their comment leader.
// Command: comment leader
func CommentLeader(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("missing files")
	}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("open %s: %w", file, err)
		}
		parsed, err := rcs.ParseFile(f)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("parse %s: %w", file, err)
		}
		fmt.Printf("%s: %s\n", file, parsed.GetComment())
	}
	return nil
}

// CommentLeaderSet sets the comment leader for the specified files.
// Command: comment leader set
func CommentLeaderSet(leader string, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("missing files")
	}
	for _, filename := range files {
		// Read file
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read %s: %w", filename, err)
		}
		// Parse
		parsed, err := rcs.ParseFile(strings.NewReader(string(content)))
		if err != nil {
			return fmt.Errorf("parse %s: %w", filename, err)
		}
		// Modify
		parsed.SetComment(leader)
		// Write back
		// Assuming String() serializes correctly
		newContent := parsed.String()
		// Need to preserve permissions?
		info, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("stat %s: %w", filename, err)
		}
		err = os.WriteFile(filename, []byte(newContent), info.Mode())
		if err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
	}
	return nil
}

// CommentLeaderList prints common RCS comment leaders.
// Command: comment leader list
func CommentLeaderList() error {
	// List common leaders
	fmt.Println("Common RCS comment leaders:")
	fmt.Println("  #     (Shell, Python, Ruby, etc.)")
	fmt.Println("  /*    (C, C++, Java, etc.)")
	fmt.Println("  //    (C++, Java, Go, etc.)")
	fmt.Println("  ;     (Lisp, Assembly)")
	fmt.Println("  %     (LaTeX, PostScript)")
	fmt.Println("  --    (Haskell, Lua, SQL)")
	fmt.Println("  REM   (BASIC, Batch)")
	fmt.Println("  '     (Visual Basic)")
	return nil
}
