package rcs

import (
	"bufio"
	"fmt"
	"hash/fnv"
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

func hashBytes(b []byte) uint64 {
	h := fnv.New64a()
	var err error
	_, err = h.Write(b)
	if err != nil {
		panic(err)
	}
	return h.Sum64()
}

func GenerateEdDiffFromLines(from []string, to []string) (EdDiff, error) {
	m := len(from)
	n := len(to)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if from[i-1] == to[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				if lcs[i-1][j] >= lcs[i][j-1] {
					lcs[i][j] = lcs[i-1][j]
				} else {
					lcs[i][j] = lcs[i][j-1]
				}
			}
		}
	}

	type actionType int
	const (
		actMatch actionType = iota
		actAdd
		actDelete
	)

	type action struct {
		kind actionType
		text string // for add
		// for delete, we don't strictly need text but useful for debug
		// for match, we don't need anything
	}

	var actions []action
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && from[i-1] == to[j-1] {
			actions = append(actions, action{kind: actMatch})
			i--
			j--
		} else if i > 0 && (j == 0 || lcs[i-1][j] >= lcs[i][j-1]) {
			// Prefer Delete (move i)
			actions = append(actions, action{kind: actDelete})
			i--
		} else {
			// Prefer Add (move j)
			actions = append(actions, action{kind: actAdd, text: to[j-1]})
			j--
		}
	}

	// Reverse actions to get forward order
	for k := 0; k < len(actions)/2; k++ {
		actions[k], actions[len(actions)-1-k] = actions[len(actions)-1-k], actions[k]
	}

	var result EdDiff
	currentLine := 0 // 0-based index of original file processed

	for k := 0; k < len(actions); k++ {
		a := actions[k]
		switch a.kind {
		case actMatch:
			currentLine++
		case actDelete:
			// Check if we can extend previous delete
			if len(result) > 0 {
				if del, ok := result[len(result)-1].(Delete); ok {
					// Check if this delete is contiguous
					// del[0] is start line (1-based), del[1] is count
					// The range deleted is [del[0], del[0] + del[1] - 1]
					// Next line to be deleted would be del[0] + del[1]
					// currentLine is 0-based index of line being processed.
					// Since we are at `actDelete`, we are about to delete `currentLine` (0-based) => `currentLine+1` (1-based).
					// So if `del[0] + del[1] == currentLine + 1`, it's contiguous.

					if del[0]+del[1] == currentLine+1 {
						// Extend
						result[len(result)-1] = Delete{del[0], del[1] + 1}
						currentLine++
						continue
					}
				}
			}

			result = append(result, Delete{currentLine + 1, 1})
			currentLine++

		case actAdd:
			// Check if we can extend previous add
			if len(result) > 0 {
				if add, ok := result[len(result)-1].(Add); ok {
					// Check if this add is at same position
					// add.LineStart is the line number *after* which we insert.
					// If we are still at the same insertion point (currentLine), extend.
					if add.LineStart == currentLine {
						// Extend
						add.Lines = append(add.Lines, a.text)
						result[len(result)-1] = add
						continue
					}
				}
			}

			result = append(result, Add{LineStart: currentLine, Lines: []string{a.text}})
			// Do not increment currentLine
		}
	}

	return result, nil
}

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
