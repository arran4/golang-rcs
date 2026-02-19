package rcs

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type KeywordSubstitution int

const (
	KV KeywordSubstitution = iota
	KVL
	K
	O
	B
	V
)

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
	}
	return KV, fmt.Errorf("unknown keyword substitution mode: %s", s)
}

type KeywordData struct {
	Revision    string
	Date        time.Time
	Author      string
	State       string
	Locker      string
	Log         string
	RCSFile     string
	WorkingFile string
}

var keywordRegex = regexp.MustCompile(`\$(Author|Date|Header|Id|Locker|Log|Name|RCSfile|Revision|Source|State)(?::[^\$]*)?\$`)

func ExpandKeywords(content string, mode KeywordSubstitution, data KeywordData) string {
	if mode == K {
		return keywordRegex.ReplaceAllStringFunc(content, func(match string) string {
			parts := strings.SplitN(match[1:len(match)-1], ":", 2)
			return "$" + parts[0] + "$"
		})
	}
	if mode == O || mode == B {
		return content
	}

	return keywordRegex.ReplaceAllStringFunc(content, func(match string) string {
		inner := match[1 : len(match)-1]
		parts := strings.SplitN(inner, ":", 2)
		keyword := parts[0]

		if mode == V {
			if keyword == "Log" {
				// V mode for Log is not well defined in common docs, but usually V strips everything except value.
				// For Log, the value is tricky.
				// Assuming standard behavior: return nothing or just the message?
				// "Generate ‘5.13’ (value only)."
				// For Log, it might be weird.
				// Let's assume empty for now or best effort.
				return ""
			}
			return getKeywordValue(keyword, data, mode)
		}

		if keyword == "Log" {
			// Special handling for Log
			// It prepends the new log entry.
			// Since we only match the keyword line, we can only replace the keyword line.
			// Standard RCS behavior for Log is to insert the log message AFTER the keyword line.
			// But since we are replacing the keyword line, we can append the log message to it.

			// Extract filename from RCSFile for the keyword value
			rcsParts := strings.Split(data.RCSFile, "/")
			filename := rcsParts[len(rcsParts)-1]

			logHeader := fmt.Sprintf("Revision %s  %s  %s", data.Revision, data.Date.UTC().Format("2006/01/02 15:04:05"), data.Author)
			// Note: The format of date in log is YYYY/MM/DD HH:MM:SS

			newLog := fmt.Sprintf("$Log: %s $\n%s\n%s\n", filename, logHeader, strings.TrimSuffix(data.Log, "\n"))

			// We retain the original match if it had content?
			// No, co regenerates the keyword line.
			// But what about existing log entries?
			// The regex only matched the first line.
			// If we return newLog, we are replacing "$Log...$" with "$Log...$\nRevision...".
			// The subsequent lines of the file are untouched.
			// So we are effectively prepending.
			return newLog
		}

		val := getKeywordValue(keyword, data, mode)
		if val == "" {
			return "$" + keyword + "$"
		}
		return "$" + keyword + ": " + val + " $"
	})
}

func getKeywordValue(keyword string, data KeywordData, mode KeywordSubstitution) string {
	switch keyword {
	case "Author":
		return data.Author
	case "Date":
		return data.Date.UTC().Format("2006/01/02 15:04:05")
	case "Header":
		locker := ""
		if data.Locker != "" {
			if mode == KVL || (mode == KV) {
				locker = " " + data.Locker
			}
		}
		return fmt.Sprintf("%s %s %s %s %s%s", data.RCSFile, data.Revision, data.Date.UTC().Format("2006/01/02 15:04:05"), data.Author, data.State, locker)
	case "Id":
		parts := strings.Split(data.RCSFile, "/")
		filename := parts[len(parts)-1]
		locker := ""
		if data.Locker != "" {
			if mode == KVL || (mode == KV) {
				locker = " " + data.Locker
			}
		}
		return fmt.Sprintf("%s %s %s %s %s%s", filename, data.Revision, data.Date.UTC().Format("2006/01/02 15:04:05"), data.Author, data.State, locker)
	case "Locker":
		if data.Locker != "" {
			if mode == KVL || (mode == KV) {
				return data.Locker
			}
		}
		return ""
	case "Name":
		return "" // Not supported yet
	case "RCSfile":
		parts := strings.Split(data.RCSFile, "/")
		return parts[len(parts)-1]
	case "Revision":
		return data.Revision
	case "Source":
		return data.RCSFile
	case "State":
		return data.State
	}
	return ""
}
