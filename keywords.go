package rcs

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// KeywordSubstitution represents the mode for keyword expansion.
type KeywordSubstitution int

const (
	KV  KeywordSubstitution = iota // $Keyword: value $
	KVL                            // $Keyword: value $ (locker always inserted if locked)
	K                              // $Keyword$
	O                              // Old value (no substitution)
	B                              // Binary (no substitution, like O)
	V                              // Value only (value)
)

// ParseKeywordSubstitution parses the substitution mode string (e.g., "kv", "kvl", "k").
func ParseKeywordSubstitution(s string) (KeywordSubstitution, error) {
	switch s {
	case "kv":
		return KV, nil
	case "kvl":
		return KVL, nil
	case "k":
		return K, nil
	case "o":
		return O, nil
	case "b":
		return B, nil
	case "v":
		return V, nil
	default:
		return KV, fmt.Errorf("unknown substitution mode: %s", s)
	}
}

// KeywordData contains metadata for keyword expansion.
type KeywordData struct {
	Revision string
	Date     time.Time
	Author   string
	State    string
	Locker   string // Name of the locker if applicable
	Log      string // Log message
	RCSFile  string // Name of the RCS file (without path)
	Source   string // Full path to the RCS file
}

var keywordRegex = regexp.MustCompile(`\$(Author|Date|Header|Id|Locker|Log|Name|RCSfile|Revision|Source|State)(:[^$]*)?\$`)

// ExpandKeywords replaces RCS keywords in the content based on the mode and data.
func ExpandKeywords(content string, data KeywordData, mode KeywordSubstitution) string {
	if mode == O || mode == B {
		return content
	}

	return keywordRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract keyword name
		parts := strings.SplitN(match[1:len(match)-1], ":", 2)
		keyword := parts[0]

		if mode == K {
			return fmt.Sprintf("$%s$", keyword)
		}

		var value string
		switch keyword {
		case "Author":
			value = data.Author
		case "Date":
			// RCS date format: YYYY/MM/DD HH:mm:ss
			value = data.Date.Format("2006/01/02 15:04:05")
		case "Header":
			// $Header: source revision date author state locker $
			value = fmt.Sprintf("%s %s %s %s %s", data.Source, data.Revision, data.Date.Format("2006/01/02 15:04:05"), data.Author, data.State)
			if data.Locker != "" {
				value += " " + data.Locker
			}
		case "Id":
			// $Id: filename revision date author state locker $
			value = fmt.Sprintf("%s %s %s %s %s", data.RCSFile, data.Revision, data.Date.Format("2006/01/02 15:04:05"), data.Author, data.State)
			if data.Locker != "" {
				value += " " + data.Locker
			}
		case "Locker":
			value = data.Locker
		case "Log":
			value = data.RCSFile
		case "Name":
			// Name is symbolic name used to check out. We don't have it in KeywordData yet, assume empty or passed in data?
			// The current call sites might not provide it.
			value = ""
		case "RCSfile":
			value = data.RCSFile
		case "Revision":
			value = data.Revision
		case "Source":
			value = data.Source
		case "State":
			value = data.State
		}

		if mode == V {
			return value
		}

		// Mode is KV or KVL

		// Handle Locker logic for KV vs KVL is done by caller populating data.Locker?
		// No, caller provides locker name if locked.
		// KV: locker inserted only if being locked (ci -l, co -l).
		// KVL: locker inserted if locked.
		// We assume data.Locker is correctly populated by the caller based on the mode.

		expanded := fmt.Sprintf("$%s: %s $", keyword, value)

		if keyword == "Log" {
			// For Log, we append the log message.
			// Format:
			// $Log: filename $
			// Revision 1.1  2020/01/01 00:00:00  tester
			// log message
			//

			// We need a way to indent or format the log message?
			// The example shows:
			// Revision 1.1  2020/01/01 00:00:00  tester
			// r1
			//
			// (with a trailing newline and maybe blank line?)

			logEntry := fmt.Sprintf("Revision %s  %s  %s\n%s",
				data.Revision,
				data.Date.Format("2006/01/02 15:04:05"),
				data.Author,
				data.Log,
			)
			// RCS usually uses the comment leader of the file.
			// The caller (Checkout) has access to comment leader, but KeywordData doesn't.
			// We should add CommentLeader to KeywordData?
			// But for now, let's assume no comment leader handling or just append it.
			// Wait, the test input has `comment @# @;`.
			// The output doesn't show comment leaders in the example expected output?
			// Wait, expected output:
			// Revision 1.1  2020/01/01 00:00:00  tester
			// r1
			//
			// It doesn't look like it uses comment leader '# '.
			// Ah, the test input file content `text` block starts with `@`.
			// The `comment` header is `@# @`.
			// If `co` inserts log, it should prefix lines with comment leader?
			// The example expected text:
			// $Log: file.txt,v $
			// Revision 1.1  2020/01/01 00:00:00  tester
			// r1
			//
			// There is no `#` prefix. Maybe because the surrounding text doesn't have it?
			// Or maybe the test expectation is simplified?
			// Or maybe I misread the test.
			// The input file has `comment @# @;`.
			// The expected output for `$Log$` is just the text.
			// If I look at `TODO-co-k-kv.txtar`:
			// $Log: file.txt,v $
			// Revision 1.1  2020/01/01 00:00:00  tester
			// r1
			//
			// It seems it just appends plain text.
			// I'll stick to appending plain text for now.

			return expanded + "\n" + logEntry
		}

		return expanded
	})
}
