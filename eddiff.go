package rcs

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type EdDiff []EdDiffCommand

func (ed EdDiff) String() string {
	var sb strings.Builder
	for _, c := range ed {
		sb.WriteString(c.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

func ParseEdDiff(r io.Reader) (EdDiff, error) {
	scanner := bufio.NewScanner(r)
	var commands []EdDiffCommand

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var cmdType rune
		var start, count int

		n, err := fmt.Sscanf(line, "%c%d %d", &cmdType, &start, &count)
		if err != nil {
			return nil, fmt.Errorf("invalid command line %q: %v", line, err)
		}
		if n < 3 {
			return nil, fmt.Errorf("invalid command line %q: expected 3 items", line)
		}

		switch cmdType {
		case 'd':
			commands = append(commands, Delete{start, count})
		case 'a':
			var lines []string
			for i := 0; i < count; i++ {
				if !scanner.Scan() {
					return nil, fmt.Errorf("unexpected EOF reading add lines for command %s", line)
				}
				lines = append(lines, scanner.Text())
			}
			commands = append(commands, Add{Lines: lines, LineStart: start})
		default:
			return nil, fmt.Errorf("unknown command type: %c", cmdType)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

var _ fmt.Stringer = (EdDiff)(nil)

func (ed EdDiff) Apply(onto LineReader, into LineWriter) error {
	adds := make(map[int][]Add)
	dels := make(map[int]Delete)

	for _, cmd := range ed {
		switch c := cmd.(type) {
		case Add:
			adds[c.LineStart] = append(adds[c.LineStart], c)
		case Delete:
			dels[c.StartLine()] = c
		}
	}

	linesRead := 0
	for {
		// 1. Process Adds (insertions after current line)
		if cmds, ok := adds[linesRead]; ok {
			for _, cmd := range cmds {
				if _, err := cmd.Apply(onto, into); err != nil {
					return err
				}
			}
			delete(adds, linesRead)
		}

		// 2. Process Deletes (starting at next line)
		if cmd, ok := dels[linesRead+1]; ok {
			n, err := cmd.Apply(onto, into)
			if err != nil {
				return err
			}
			linesRead += n
			continue
		}

		// 3. Read/Write one line
		line, err := onto.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := into.WriteLine(line); err != nil {
			return err
		}
		linesRead++
	}

	return nil
}

// GenerateEdDiffFromLines delegates to the default registered algorithm.
// The actual implementation is now pluggable.
func GenerateEdDiffFromLines(from []string, to []string) (EdDiff, error) {
	// Circular dependency avoidance: eddiff.go (rcs) imports diff, diff imports rcs.
	// But we need to call into diff package.
	// The implementation of LCS is in diff/lcs package which imports rcs.
	// We need a way to call the registered algorithm.
	// We can't import "github.com/arran4/golang-rcs/diff" here if "diff" imports "rcs".
	// WAIT. rcs package is the base. diff package can import rcs.
	// lcs package imports rcs and diff.
	// But we need to call diff.DefaultAlgorithm() here.
	// If diff imports rcs, and rcs imports diff, we have a cycle.
	// The plan needs to be adjusted.
	// We can't have rcs -> diff -> rcs.
	// We should move EdDiff definitions to a common package or keep them in rcs, and diff package should import rcs.
	// Then rcs CANNOT import diff.
	// So GenerateEdDiffFromLines in rcs package cannot call diff.DefaultAlgorithm().
	//
	// SOLUTION:
	// Use a variable in rcs package that can be set by the diff package or main.
	// var DefaultDiffAlgorithm func([]string, []string) (EdDiff, error)
	//
	// OR:
	// Move EdDiff types to a subpackage or independent package.
	// But EdDiff is core to RCS.
	//
	// Alternative:
	// The `diff` package defines the interface and registry. It imports `rcs` to return `EdDiff`.
	// `rcs` does NOT import `diff`.
	// `rcs.GenerateEdDiffFromLines` becomes a variable or a function that panics if not set, or we implement a default simple one here.
	// Or we simply remove `GenerateEdDiffFromLines` from `rcs` package and force users to use `diff.Generate`.
	// But existing code might use `rcs.GenerateEdDiffFromLines`.
	// The user asked to "allow me to switch algorithm by name using a registry... diff/register.go".
	//
	// Let's use a variable injection for now to break the cycle if we want to keep the function signature in `rcs`.
	if DiffAlgorithm != nil {
		return DiffAlgorithm(from, to)
	}
	return nil, fmt.Errorf("no diff algorithm registered")
}

var DiffAlgorithm func(from []string, to []string) (EdDiff, error)

type EdDiffCommand interface {
	fmt.Stringer
	StartLine() int
	Apply(onto LineReader, into LineWriter) (int, error)
}

type LineReader interface {
	ReadLine() (string, error)
}

type LineWriter interface {
	WriteLine(line string) error
}

type Delete [2]int

func (d Delete) StartLine() int {
	return d[0]
}

func (d Delete) String() string {
	return fmt.Sprintf("d%d %d", d[0], d[1])
}

func (d Delete) Apply(onto LineReader, into LineWriter) (int, error) {
	for i := 0; i < d[1]; i++ {
		_, err := onto.ReadLine()
		if err != nil {
			return 0, err
		}
	}
	return d[1], nil
}

var _ EdDiffCommand = Delete{}

type Add struct {
	Lines     []string
	LineStart int
}

func (a Add) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "a%d %d\n", a.LineStart, len(a.Lines))
	for i, l := range a.Lines {
		sb.WriteString(l)
		if i < len(a.Lines)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (a Add) StartLine() int {
	return a.LineStart
}

func (a Add) Apply(onto LineReader, into LineWriter) (int, error) {
	for _, line := range a.Lines {
		err := into.WriteLine(line)
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

var _ EdDiffCommand = Add{}
