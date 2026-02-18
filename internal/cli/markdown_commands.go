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

	outString := rcsFileToMarkdown(r)
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

func rcsFileToMarkdown(f *rcs.File) string {
	var sb strings.Builder
	sb.WriteString("# RCS File\n\n")

	sb.WriteString("## Header\n\n")
	if f.Head != "" {
		sb.WriteString(fmt.Sprintf("* Head: %s\n", f.Head))
	}
	if f.Branch != "" {
		sb.WriteString(fmt.Sprintf("* Branch: %s\n", f.Branch))
	}
	if len(f.AccessUsers) > 0 {
		sb.WriteString("* Access:\n")
		for _, u := range f.AccessUsers {
			sb.WriteString(fmt.Sprintf("  * %s\n", u))
		}
	}
	if len(f.Symbols) > 0 {
		sb.WriteString("* Symbols:\n")
		for _, s := range f.Symbols {
			sb.WriteString(fmt.Sprintf("  * %s: %s\n", s.Name, s.Revision))
		}
	}
	if len(f.Locks) > 0 {
		sb.WriteString("* Locks:\n")
		for _, l := range f.Locks {
			sb.WriteString(fmt.Sprintf("  * %s: %s\n", l.User, l.Revision))
		}
	}
	if f.Strict {
		sb.WriteString("* Strict: true\n")
	}
	if f.Comment != "" {
		sb.WriteString(fmt.Sprintf("* Comment: %s\n", f.Comment))
	}
	if f.Expand != "" {
		sb.WriteString(fmt.Sprintf("* Expand: %s\n", f.Expand))
	}
	sb.WriteString("\n")

	sb.WriteString("## Description\n\n")
	writeFencedBlock(&sb, "text", f.Description)
	sb.WriteString("\n")

	sb.WriteString("## Revisions\n\n")

	// Map RevisionContents by Revision string for easy lookup
	contentsMap := make(map[string]*rcs.RevisionContent)
	for _, rc := range f.RevisionContents {
		contentsMap[rc.Revision] = rc
	}

	for _, rh := range f.RevisionHeads {
		rev := rh.Revision.String()
		sb.WriteString(fmt.Sprintf("### %s\n\n", rev))
		sb.WriteString(fmt.Sprintf("* Date: %s\n", rh.Date))
		sb.WriteString(fmt.Sprintf("* Author: %s\n", rh.Author))
		sb.WriteString(fmt.Sprintf("* State: %s\n", rh.State))
		if len(rh.Branches) > 0 {
			branches := make([]string, len(rh.Branches))
			for i, b := range rh.Branches {
				branches[i] = b.String()
			}
			sb.WriteString(fmt.Sprintf("* Branches: %s\n", strings.Join(branches, " ")))
		}
		if rh.NextRevision != "" {
			sb.WriteString(fmt.Sprintf("* Next: %s\n", rh.NextRevision))
		}
		if rh.CommitID != "" {
			sb.WriteString(fmt.Sprintf("* CommitID: %s\n", rh.CommitID))
		}
		sb.WriteString("\n")

		if rc, ok := contentsMap[rev]; ok {
			sb.WriteString("#### Log\n\n")
			writeFencedBlock(&sb, "text", rc.Log)
			sb.WriteString("\n")

			sb.WriteString("#### Text\n\n")
			writeFencedBlock(&sb, "text", rc.Text)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func writeFencedBlock(sb *strings.Builder, lang, content string) {
	fence := "```"
	// Check if content contains fence, if so, extend fence
	for strings.Contains(content, fence) {
		fence += "`"
	}
	sb.WriteString(fence + lang + "\n")
	if content != "" {
		// Ensure content ends with newline if not empty
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString(content + "\n")
		} else {
			sb.WriteString(content)
		}
	} else {
		// Empty content
	}
	sb.WriteString(fence + "\n")
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

type parseState int

const (
	stateStart parseState = iota
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
	var state parseState = stateStart
	var subState string // "Access", "Symbols", "Locks"

	var currentRevision *rcs.RevisionHead
	var currentContent *rcs.RevisionContent

	var multilineBuilder strings.Builder
	var fence string
	inFence := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if inFence {
			if strings.HasPrefix(line, fence) && strings.TrimSpace(line) == fence {
				inFence = false
				content := multilineBuilder.String()
				// Remove the last newline which is usually added by the builder loop
				// if it matches what writeFencedBlock does.
				// writeFencedBlock ensures content ends with \n inside the block.
				// Here we append \n for each line. So the last line has \n.
				// If the original content had no trailing newline, writeFencedBlock added one.
				// So we should probably keep it?
				// But strings.Join(lines, "\n") would be better.
				// Current approach: we append line + "\n".
				// So "line1\nline2\n".
				// This is correct for preserving lines.
				// However, if the source had "line1", we write "line1\n".
				// So we read "line1\n".
				// RCS generally expects newlines.

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
				}
				multilineBuilder.Reset()
				continue
			}
			multilineBuilder.WriteString(line + "\n")
			continue
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
		if strings.HasPrefix(line, "### ") && (state == stateRevisions || state == stateRevision) {
			rev := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			currentRevision = &rcs.RevisionHead{
				Revision: rcs.Num(rev),
			}
			f.RevisionHeads = append(f.RevisionHeads, currentRevision)
			currentContent = &rcs.RevisionContent{
				Revision: rev,
			}
			f.RevisionContents = append(f.RevisionContents, currentContent)
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

		// Fence start
		if strings.HasPrefix(line, "```") && (state == stateDescription || state == stateLog || state == stateText) {
			fence = strings.TrimSpace(line)
			// Remove language hint
			// Fence is just the backticks part of the line start?
			// But `strings.TrimSpace` might have removed indentation?
			// We check `line` for prefix.
			// Usually fenced blocks start at start of line in our output.

			// Find how many backticks
			ticks := 0
			trimmedFence := strings.TrimSpace(fence)
			for _, r := range trimmedFence {
				if r == '`' {
					ticks++
				} else {
					break
				}
			}
			if ticks < 3 {
				// Not a valid fence, probably text?
				// But we are outside fence context, looking for start.
				// If it's `Key: Value`, it won't start with ```.
				// Only Description/Log/Text start with fence.
			} else {
				fence = trimmedFence[:ticks]
				inFence = true
			}
			continue
		}

		// Key-Value
		if strings.HasPrefix(trimmed, "* ") {
			parts := strings.SplitN(trimmed, ":", 2)
			key := strings.TrimSpace(strings.TrimPrefix(parts[0], "* "))
			val := ""
			if len(parts) > 1 {
				val = strings.TrimSpace(parts[1])
			}

			if state == stateHeader {
				// Check indentation to see if it is a sub-item
				indent := 0
				for _, r := range line {
					if r == ' ' {
						indent++
					} else {
						break
					}
				}

				if indent > 0 && subState != "" {
					// Sub-item
					// Remove `* ` from trimmed (already done in parts[0])
					// But `parts` comes from `trimmed`.
					// `trimmed` starts with `* `.
					// `key` is `tag` in `* tag: rev`

					switch subState {
					case "Access":
						// Format: `  * user`
						// key is `user`
						f.AccessUsers = append(f.AccessUsers, key)
						f.Access = true
					case "Symbols":
						// Format: `  * tag: rev`
						// key is `tag`, val is `rev`
						f.Symbols = append(f.Symbols, &rcs.Symbol{Name: key, Revision: val})
					case "Locks":
						// Format: `  * user: rev`
						// key is `user`, val is `rev`
						f.Locks = append(f.Locks, &rcs.Lock{User: key, Revision: val})
					}
				} else {
					// Top level item
					subState = ""
					switch key {
					case "Head":
						f.Head = val
					case "Branch":
						f.Branch = val
					case "Strict":
						if val == "true" {
							f.Strict = true
						}
					case "Comment":
						f.Comment = val
					case "Expand":
						f.Expand = val
					case "Access":
						subState = "Access"
					case "Symbols":
						subState = "Symbols"
					case "Locks":
						subState = "Locks"
					}
				}
			} else if state == stateRevision {
				if currentRevision != nil {
					switch key {
					case "Date":
						currentRevision.Date = rcs.DateTime(val)
					case "Author":
						currentRevision.Author = rcs.ID(val)
					case "State":
						currentRevision.State = rcs.ID(val)
					case "Branches":
						if val != "" {
							for _, b := range strings.Fields(val) {
								currentRevision.Branches = append(currentRevision.Branches, rcs.Num(b))
							}
						}
					case "Next":
						currentRevision.NextRevision = rcs.Num(val)
					case "CommitID":
						currentRevision.CommitID = rcs.Sym(val)
					}
				}
			}
			continue
		}
	}

	return f, nil
}
