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
	// Use the template defined in markdown_template.go
	// However, template funcs "fenced" need to be defined.
	// We defined them in markdown_template.go
	// We just need to execute it.

	// We also need to prepare data if necessary.
	// But `rcs.File` matches the template structure.
	// Except `RevisionContents` order must match `RevisionHeads`.
	// The RCS parser guarantees they are parsed in order?
	// `ParseFile` populates `RevisionHeads` and `RevisionContents`.
	// They correspond by index if `ParseRevisionContents` parses them in same order as headers.
	// Let's verify `parser.go`.

	// In `parser.go`, `ParseRevisionHeaders` reads headers.
	// Then `ParseRevisionContents` reads contents.
	// RCS file format: headers first, then description, then contents.
	// Usually contents appear in same order as headers?
	// Wait, RCS file has `desc` then `deltatext`.
	// `deltatext` blocks are keyed by revision number.
	// `parser.go` `ParseRevisionContents` reads them sequentially.
	// Does it guarantee order?
	// `ParseRevisionContents` loop reads until EOF.
	// It appends to `rcs`.

	// If the order in file differs from headers, index access `index $.RevisionContents $i` will be WRONG.
	// We must map them by revision.

	// Let's create a struct wrapper for template execution.

	contentsMap := make(map[string]*rcs.RevisionContent)
	for _, rc := range f.RevisionContents {
		contentsMap[rc.Revision] = rc
	}

	// We need to pass this map to template.
	// But template iterates `RevisionHeads`.
	// We can use a custom function `get_content`.

	// Re-parse template with extra func?
	// `markdownTemplate` is global. We can't change funcs easily per execution without cloning (which is fine).
	// Or we can just attach the map to the data and use `index`.
	// But `index` on map works in Go templates.

	// So we need to pass a struct that has the map.
	// And update template to use `index .RevisionContentsMap $rh.Revision`.
	// But `$rh.Revision` is `rcs.Num` (string alias).
	// Map key is `string`. `rcs.Num` should work as key? No, types must match.
	// `rcs.Num` is `string`.

	// Let's redefine the template to support this lookup safely.

	// Since `markdownTemplate` is in another file, we should update IT to support map lookup.
	// But `markdownTemplate` is initialized in `init` (var block).
	// We can clone it and add func?

	t, err := markdownTemplate.Clone()
	if err != nil {
		return "", fmt.Errorf("failed to clone template: %w", err)
	}

	// We can add a function `content` that takes a revision string and returns content.
	// We can't use template.FuncMap because template package is not imported.
	// But wait, we used template.FuncMap in `markdown_template.go`.
	// We need to import "text/template" here if we want to use it.
	// Or we can rely on `t.Funcs(map[string]interface{}{...})` if it accepted that.
	// But it requires template.FuncMap which is map[string]interface{}.
	// So we must import text/template.

	// But wait, the template string is already parsed. We can't change the text of the template easily.
	// We defined the template text in `markdown_template.go`.
	// If we want to change logic, we should update `markdown_template.go` to use a function we provide, or use a map we provide.

	// Let's update `markdown_template.go` to use `call .Content rev`.
	// Or just `{{with (call $.ContentLookup .Revision)}}`.

	// Better: Prepare a slice of structs that has Head and Content paired.
	// This is cleaner for template.

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

	// We need to update the template to iterate `.Revisions`.
	// Update `markdown_template.go` first.

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
