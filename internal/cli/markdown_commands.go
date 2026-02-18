package cli

import (
	"bufio"
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io"
	"strings"
)

// ToMarkdown is a subcommand `gorcs to-markdown`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	files: ... List of files to process, or - for stdin
func ToMarkdown(output string, force bool, files ...string) error {
	var err error
	if files, err = ensureFiles(files); err != nil {
		return err
	}
	if output != "" && output != "-" && len(files) > 1 {
		return fmt.Errorf("cannot specify output file with multiple input files")
	}
	for _, fn := range files {
		if err := processFileToMarkdown(fn, output, force); err != nil {
			return err
		}
	}
	return nil
}

func processFileToMarkdown(fn string, output string, force bool) error {
	f, err := OpenFile(fn, false)
	if err != nil {
		return fmt.Errorf("error with file %s: %w", fn, err)
	}
	defer func() {
		_ = f.Close()
	}()
	r, err := rcs.ParseFile(f)
	if err != nil {
		return fmt.Errorf("error parsing %s: %w", fn, err)
	}

	outString, err := rcsFileToMarkdown(r)
	if err != nil {
		return fmt.Errorf("error converting %s to markdown: %w", fn, err)
	}
	b := []byte(outString)

	if output == "-" {
		fmt.Printf("%s", b)
	} else if output != "" {
		if err := writeOutput(output, b, force); err != nil {
			return err
		}
	} else if fn == "-" {
		// When reading from stdin and no output file specified, write to stdout
		fmt.Printf("%s", b)
	} else {
		// Default output: filename + .md
		outPath := fn + ".md"
		if err := writeOutput(outPath, b, force); err != nil {
			return err
		}
	}
	return nil
}

func rcsFileToMarkdown(f *rcs.File) (string, error) {
	var sb strings.Builder

	contentsMap := make(map[string]*rcs.RevisionContent)
	for _, rc := range f.RevisionContents {
		contentsMap[rc.Revision] = rc
	}

	t, err := markdownTemplate.Clone()
	if err != nil {
		return "", fmt.Errorf("failed to clone template: %w", err)
	}

	type RevisionPair struct {
		Head    *rcs.RevisionHead
		Content *rcs.RevisionContent
	}

	var revisions []RevisionPair
	for _, rh := range f.RevisionHeads {
		rev := rh.Revision.String()
		rc := contentsMap[rev]
		if rc == nil {
			rc = &rcs.RevisionContent{}
		}
		revisions = append(revisions, RevisionPair{Head: rh, Content: rc})
	}

	data := struct {
		*rcs.File
		Revisions []RevisionPair
	}{
		File:      f,
		Revisions: revisions,
	}

	if err := t.Execute(&sb, data); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// FromMarkdown is a subcommand `gorcs from-markdown`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	mmap: -m --mmap Use mmap to read file
//	files: ... List of files to process, or - for stdin
func FromMarkdown(output string, force, useMmap bool, files ...string) error {
	var err error
	if files, err = ensureFiles(files); err != nil {
		return err
	}
	if output != "" && output != "-" && len(files) > 1 {
		return fmt.Errorf("cannot specify output file with multiple input files")
	}
	for _, fn := range files {
		if err := processFileFromMarkdown(fn, output, force, useMmap); err != nil {
			return err
		}
	}
	return nil
}

func processFileFromMarkdown(fn string, output string, force, useMmap bool) error {
	f, err := OpenFile(fn, useMmap)
	if err != nil {
		return fmt.Errorf("error with file %s: %w", fn, err)
	}
	defer func() {
		_ = f.Close()
	}()

	r, err := parseMarkdownFile(f)
	if err != nil {
		return fmt.Errorf("error parsing markdown %s: %w", fn, err)
	}

	outBytes := []byte(r.String())

	if output == "-" {
		fmt.Print(string(outBytes))
	} else if output != "" {
		if err := writeOutput(output, outBytes, force); err != nil {
			return err
		}
	} else if fn == "-" {
		fmt.Print(string(outBytes))
	} else {
		// Default output: remove .md suffix, append ,v if not present
		outPath := fn
		if strings.HasSuffix(fn, ".md") {
			outPath = strings.TrimSuffix(fn, ".md")
		}
		if !strings.HasSuffix(outPath, ",v") {
			outPath += ",v"
		}
		if err := writeOutput(outPath, outBytes, force); err != nil {
			return err
		}
	}
	return nil
}

const (
	stateStart = iota
	stateHeader
	stateDescription
	stateRevisions
	stateRevision
	stateLog
	stateText
)

func parseMarkdownFile(r io.Reader) (*rcs.File, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024*10)

	f := rcs.NewFile()
	var state = stateStart
	var subState string // "Access", "Symbols", "Locks", "Comment"

	var currentContent *rcs.RevisionContent

	var multilineBuilder strings.Builder
	inQuote := false

	commitQuote := func() {
		if !inQuote {
			return
		}
		inQuote = false
		content := multilineBuilder.String()
		if len(content) > 0 && content[len(content)-1] == '\n' {
			content = content[:len(content)-1]
		}

		switch state {
		case stateDescription:
			f.Description = content
			state = stateStart
		case stateLog:
			if currentContent != nil {
				currentContent.Log = content
			}
			state = stateRevision
		case stateText:
			if currentContent != nil {
				currentContent.Text = content
			}
			state = stateRevision
		case stateHeader:
			if subState == "Comment" {
				f.Comment = content
			}
		}
		multilineBuilder.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		isQuoteLine := strings.HasPrefix(trimmed, ">")

		// Check if we expect a quote
		expectingQuote := state == stateDescription || state == stateLog || state == stateText || (state == stateHeader && subState == "Comment")

		if isQuoteLine && expectingQuote {
			inQuote = true
			content := strings.TrimPrefix(trimmed, ">")
			if len(content) > 0 && content[0] == ' ' {
				content = content[1:]
			}
			multilineBuilder.WriteString(content + "\n")
			continue
		}

		// If we were quoting, but this line is not a quote, or we are not expecting one anymore
		if inQuote {
			commitQuote()
		}

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(line, "# RCS File") {
			continue
		}
		if strings.HasPrefix(line, "## Header") {
			state = stateHeader
			subState = ""
			continue
		}
		if strings.HasPrefix(line, "## Description") {
			state = stateDescription
			subState = ""
			continue
		}
		if strings.HasPrefix(line, "## Revisions") {
			state = stateRevisions
			subState = ""
			continue
		}

		if strings.HasPrefix(line, "### ") {
			revStr := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			revStr = strings.Trim(revStr, "`")

			found := false
			for _, rc := range f.RevisionContents {
				if rc.Revision == revStr {
					currentContent = rc
					found = true
					break
				}
			}
			if !found {
				currentContent = &rcs.RevisionContent{Revision: revStr}
				f.RevisionContents = append(f.RevisionContents, currentContent)
			}
			state = stateRevision
			subState = ""
			continue
		}

		if strings.HasPrefix(line, "#### Log") && state == stateRevision {
			state = stateLog
			continue
		}
		if strings.HasPrefix(line, "#### Text") && state == stateRevision {
			state = stateText
			continue
		}

		if state == stateHeader && strings.HasPrefix(trimmed, ":") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, ":"))
			unquote := func(s string) string {
				return strings.Trim(s, "`")
			}

			switch subState {
			case "Head":
				f.Head = unquote(val)
			case "Branch":
				f.Branch = unquote(val)
			case "Strict":
				if val == "true" {
					f.Strict = true
				}
			case "Expand":
				f.Expand = unquote(val)
			case "Access":
				f.AccessUsers = append(f.AccessUsers, unquote(val))
				f.Access = true
			case "Symbols":
				parts := strings.SplitN(val, ":", 2)
				if len(parts) == 2 {
					f.Symbols = append(f.Symbols, &rcs.Symbol{Name: unquote(strings.TrimSpace(parts[0])), Revision: unquote(strings.TrimSpace(parts[1]))})
				}
			case "Locks":
				parts := strings.SplitN(val, ":", 2)
				if len(parts) == 2 {
					f.Locks = append(f.Locks, &rcs.Lock{User: unquote(strings.TrimSpace(parts[0])), Revision: unquote(strings.TrimSpace(parts[1]))})
				}
			}
			continue
		}

		if state == stateHeader && trimmed != "" && !strings.HasPrefix(trimmed, ":") {
			// If it's a quote for Comment, it should have been handled by isQuoteLine check above if subState is Comment.
			// If we are here, it's a key.
			subState = strings.TrimSpace(trimmed)
			continue
		}

		if state == stateRevisions && strings.HasPrefix(trimmed, "|") {
			parts := strings.Split(trimmed, "|")
			if len(parts) < 8 {
				continue
			}

			cols := make([]string, 0, len(parts))
			for _, p := range parts {
				cols = append(cols, strings.TrimSpace(p))
			}

			rev := strings.Trim(cols[1], "`")
			if rev == "Revision" || strings.HasPrefix(rev, ":") || rev == "" {
				continue
			}

			rh := &rcs.RevisionHead{
				Revision: rcs.Num(rev),
				Date:     rcs.DateTime(cols[2]),
				Author:   rcs.ID(cols[3]),
				State:    rcs.ID(cols[4]),
			}

			branchesStr := cols[5]
			if branchesStr != "" {
				bs := strings.Fields(branchesStr)
				for _, b := range bs {
					rh.Branches = append(rh.Branches, rcs.Num(strings.Trim(b, "`")))
				}
			}

			rh.NextRevision = rcs.Num(strings.Trim(cols[6], "`"))
			rh.CommitID = rcs.Sym(strings.Trim(cols[7], "`"))

			f.RevisionHeads = append(f.RevisionHeads, rh)

			exists := false
			for _, rc := range f.RevisionContents {
				if rc.Revision == rev {
					exists = true
					break
				}
			}
			if !exists {
				f.RevisionContents = append(f.RevisionContents, &rcs.RevisionContent{Revision: rev})
			}

			continue
		}
	}

	if inQuote {
		commitQuote()
	}

	return f, nil
}
