package rcs

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// PrintRLog prints the RCS file information in rlog format.
func PrintRLog(w io.Writer, f *File, filename string, workingFilename string, filter Filter) error {
	var err error
	printf := func(format string, a ...interface{}) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, format, a...)
	}

	printf("RCS file: %s\n", filename)
	printf("Working file: %s\n", workingFilename)
	printf("head: %s\n", f.Head)
	printf("branch: %s\n", f.Branch)

	locksStr := ""
	if f.Strict {
		locksStr = "strict"
	}
	printf("locks: %s", locksStr)
	for _, l := range f.Locks {
		printf("\n\t%s: %s", l.User, l.Revision)
	}
	printf("\n")

	printf("access list:\n")
	for _, user := range f.AccessUsers {
		printf("\t%s\n", user)
	}

	printf("symbolic names:\n")
	for _, sym := range f.Symbols {
		printf("\t%s: %s\n", sym.Name, sym.Revision)
	}

	expand := f.Expand
	if expand == "" {
		expand = "kv"
	}
	printf("keyword substitution: %s\n", expand)

	var revisionsToPrint []*RevisionHead
	for _, rh := range f.RevisionHeads {
		if filter == nil || filter.Match(rh) {
			revisionsToPrint = append(revisionsToPrint, rh)
		}
	}

	printf("total revisions: %d;\tselected revisions: %d\n", len(f.RevisionHeads), len(revisionsToPrint))
	printf("description:\n%s", f.Description)

	for _, rh := range revisionsToPrint {
		printf("----------------------------\n")
		printf("revision %s\n", rh.Revision)

		dateStr := string(rh.Date)
		t, e := ParseDate(dateStr, time.Time{}, nil)
		if e == nil {
			dateStr = t.Format("2006/01/02 15:04:05")
		}

		printf("date: %s;  author: %s;  state: %s;", dateStr, rh.Author, rh.State)

		// Calculate lines stats
		linesStats, e := getLinesStats(f, rh)
		if e == nil && linesStats != "" {
			printf("%s", linesStats)
		}

		if len(rh.Branches) > 0 {
			printf("  branches:")
			for _, b := range rh.Branches {
				printf(" %s;", b)
			}
		}
		// next field is typically not shown in default rlog output unless verbose/debug

		printf("\n")

		logMsg, e := f.GetLogMessage(string(rh.Revision))
		if e == nil {
			printf("%s\n", logMsg)
		}
	}
	printf("=============================================================================\n")

	return err
}

func getLinesStats(f *File, rh *RevisionHead) (string, error) {
	// Determine if trunk or branch based on number of parts in revision
	// Simplistic check: X.Y is trunk, X.Y.Z.W is branch
	// This might fail for Vendor branches (1.1.1.1) which are branches but treated specially?
	// Standard RCS logic: if revision has odd number of dots (even number of fields), it's trunk or branch tip.
	// But structure is recursive.
	// 1.2 is trunk. 1.2.1.1 is branch.
	// Trunk revisions have 2 components.
	parts := strings.Split(string(rh.Revision), ".")
	isTrunk := len(parts) == 2

	var text string
	var isReverse bool

	if isTrunk {
		// Reverse delta: stored in NextRevision
		nextRev := string(rh.NextRevision)
		if nextRev == "" {
			return "", nil // No next revision (e.g. 1.1), so no delta to compare against
		}
		// Find next revision content
		found := false
		for _, rc := range f.RevisionContents {
			if rc.Revision == nextRev {
				text = rc.Text
				found = true
				break
			}
		}
		if !found {
			// Next revision declared but content missing?
			return "", nil
		}
		isReverse = true
	} else {
		// Forward delta: stored in current revision
		// Find own content
		found := false
		for _, rc := range f.RevisionContents {
			if rc.Revision == string(rh.Revision) {
				text = rc.Text
				found = true
				break
			}
		}
		if !found {
			return "", nil
		}
		isReverse = false
	}

	dCount, aCount := parseDeltaStats(text)

	if isReverse {
		// Reverse delta:
		// d N M means delete M lines at N. These lines are in current revision but not in next.
		// So they are "added" in current relative to next. -> +dCount
		// a N M means add M lines at N. These lines are in next but not in current.
		// So they are "removed" in current relative to next. -> -aCount
		return fmt.Sprintf("  lines: +%d -%d", dCount, aCount), nil
	} else {
		// Forward delta:
		// d N M means delete M lines. Removed from previous. -> -dCount
		// a N M means add M lines. Added to previous. -> +aCount
		return fmt.Sprintf("  lines: +%d -%d", aCount, dCount), nil
	}
}

func parseDeltaStats(delta string) (dCount, aCount int) {
	scanner := bufio.NewScanner(strings.NewReader(delta))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		cmd := line[0]
		if cmd == 'a' || cmd == 'd' {
			// Format: d<start> <count> or a<start> <count>
			// But wait, is there space after command char?
			// RCS format says: "dla count" or "ala count". "la" is line number.
			// Examples: "d1 1", "a1 1". No space between command char and line number?
			// Let's check parser code or examples.
			// Example: "d1 1". 'd' at 0. '1' at 1.
			// So no space.
			// We need to parse "d<num> <num>".

			// Check if line matches pattern
			parts := strings.Fields(line[1:])
			// parts[0] is part of line number if space existed?
			// If "d1 1", line[1:] is "1 1". Fields: ["1", "1"].
			// If "d12 3", line[1:] is "12 3". Fields: ["12", "3"].
			// If "d 1 1" (invalid?).

			if len(parts) >= 2 {
				count, err := strconv.Atoi(parts[1])
				if err != nil {
					continue // Not a command line
				}
				// Also check if parts[0] is number
				if _, err := strconv.Atoi(parts[0]); err != nil {
					continue
				}

				if cmd == 'a' {
					aCount += count
					// Skip 'count' lines of data
					for i := 0; i < count; i++ {
						if !scanner.Scan() {
							break
						}
					}
				} else {
					dCount += count
				}
			}
		}
	}
	return
}
