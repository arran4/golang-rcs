package rcs

import (
	"fmt"
	"regexp"
	"strings"
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

type KeywordContext struct {
	Revision    string
	Date        string
	Author      string
	State       string
	Log         string
	RCSFile     string
	WorkingFile string
	Locker      string
	Strict      bool
	Comment     string // Comment leader
}

var keywordRegex = regexp.MustCompile(`\$([A-Z][a-z]+)(?::[^\$]*)?\$`)

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
		return KV, fmt.Errorf("unknown keyword substitution mode: %s", s)
	}
}

func ExpandKeywords(content string, mode KeywordSubstitution, ctx KeywordContext) string {
	if mode == O || mode == B {
		return content
	}

	return keywordRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := keywordRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		keyword := parts[1]

		// Log is special
		if keyword == "Log" {
			return expandLog(match, ctx, mode)
		}

		switch mode {
		case K:
			return fmt.Sprintf("$%s$", keyword)
		case V:
			val := getKeywordValue(keyword, ctx, mode)
			if val == "" {
				return match
			}
			return val
		case KV, KVL:
			val := getKeywordValue(keyword, ctx, mode)
			if val == "" {
				return fmt.Sprintf("$%s$", keyword)
			}
			return fmt.Sprintf("$%s: %s $", keyword, val)
		}
		return match
	})
}

func getKeywordValue(keyword string, ctx KeywordContext, mode KeywordSubstitution) string {
	switch keyword {
	case "Author":
		return ctx.Author
	case "Date":
		return formatDate(ctx.Date)
	case "Header":
		h := fmt.Sprintf("%s %s %s %s %s", ctx.RCSFile, ctx.Revision, formatDate(ctx.Date), ctx.Author, ctx.State)
		if ctx.Locker != "" {
			h += " " + ctx.Locker
		}
		return h
	case "Id":
		filename := getBasename(ctx.RCSFile)
		h := fmt.Sprintf("%s %s %s %s %s", filename, ctx.Revision, formatDate(ctx.Date), ctx.Author, ctx.State)
		if ctx.Locker != "" {
			h += " " + ctx.Locker
		}
		return h
	case "Locker":
		if mode == KVL || (mode == KV && ctx.Locker != "") {
			return ctx.Locker
		}
		return ""
	case "Name":
		return "" // TODO: Implement tag/symbol
	case "RCSfile":
		return getBasename(ctx.RCSFile)
	case "Revision":
		return ctx.Revision
	case "Source":
		return ctx.RCSFile
	case "State":
		return ctx.State
	}
	return ""
}

func getBasename(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func formatDate(rcsDate string) string {
	parts := strings.Split(rcsDate, ".")
	if len(parts) >= 6 {
		return fmt.Sprintf("%s/%s/%s %s:%s:%s", parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
	}
	return rcsDate
}

func expandLog(match string, ctx KeywordContext, mode KeywordSubstitution) string {
	filename := getBasename(ctx.RCSFile)

	// Format: Revision <rev>  <date>  <author>
	//         <message>
	//
	cleanLog := strings.TrimSpace(ctx.Log)
	newEntry := fmt.Sprintf("Revision %s  %s  %s\n%s\n", ctx.Revision, formatDate(ctx.Date), ctx.Author, cleanLog)

	if mode == V {
		// Just value: filename \n newEntry
		return fmt.Sprintf("%s\n%s", filename, newEntry)
	}

	// K, KV, KVL
	header := fmt.Sprintf("$Log: %s $", filename)
	return fmt.Sprintf("%s\n%s", header, newEntry)
}
