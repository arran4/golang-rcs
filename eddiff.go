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
