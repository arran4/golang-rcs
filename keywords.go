package rcs

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
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

var (
	keywordRegex     *regexp.Regexp
	keywordRegexOnce sync.Once
)

// ExpandKeywords replaces RCS keywords in the content based on the mode and data.
func ExpandKeywords(content string, data KeywordData, mode KeywordSubstitution) string {
	if mode == O || mode == B {
		return content
	}

	keywordRegexOnce.Do(func() {
		keywordRegex = regexp.MustCompile(`\$(Author|Date|Header|Id|Locker|Log|Name|RCSfile|Revision|Source|State)(:[^$]*)?\$`)
	})

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

		expanded := fmt.Sprintf("$%s: %s $", keyword, value)

		if keyword == "Log" {
			logEntry := fmt.Sprintf("Revision %s  %s  %s\n%s",
				data.Revision,
				data.Date.Format("2006/01/02 15:04:05"),
				data.Author,
				data.Log,
			)
			// TODO: Add support for comment leader insertion
			return expanded + "\n" + logEntry
		}

		return expanded
	})
}
