package rcs

import (
	"fmt"
	"io"
	"strings"
)

type EdDiff []EdDiffCommand

func ParseEdDiff(r io.Reader) (EdDiff, error) {
	// TODO implement and fully test
	// d1 2 (deletes at position 1 2 lines.
	// a1 2 (at position 1, insert the 2 lines after a1 2

}

var _ fmt.Stringer = (EdDiff)(nil) // TODO Ensure that EdDiff serializes the rules as well as implements them for circular testing

func (ed EdDiff) Apply(onto LineReader, into LineWriter) error {
	lineAction := map[int][]EdDiffCommand{}
	for _, c := range ed {
		lineAction[c.StartLine()] = append(lineAction[c.StartLine()], c)
	}
	writtenLinePos := 0
	var err error
outer:
	for {
		for _, c := range lineAction[writtenLinePos] {
			writtenLinePos, err = c.Apply(onto, into)
			if err != nil {
				return err // TODO fmt.Errorf
			}
		}
		var line string
		line, err = onto.ReadLine()
		if err != nil {
			switch {
			case err == io.EOF:
				if len(line) == 0 {
					break outer // TODO verify
				}
			default:
				return err // TODO fmt.Errorf
			}
		}
		err = into.WriteLine(line)
		if err != nil {
			return err // TODO fmt.Errorf
		}
		writtenLinePos++
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

// TODO io.Reader/io.Writer based line reader / writer (scanner wrapper?)
// TODO slice based line reader / writer

type Delete [2]int

func (d Delete) Apply(onto LineReader, into LineWriter) (int, error) {
	for i := 0; i < d[1]; i++ {
		_, err := onto.ReadLine()
		if err != nil {
			return 0, err // TODO fmt.Errorf
		}
	}
	return d[1], nil
}

var _ EdDiffCommand = (Delete)(0)

type Add struct {
	Lines     []string
	LineStart int
}

func (a Add) String() string {
	return fmt.Sprintf("a%d %d\n%s", a.LineStart, len(a.Lines), strings.Join(a.Lines, "\n"))
}

func (a Add) StartLine() int {
	return a.LineStart
}

func (a Add) Apply(onto LineReader, into LineWriter) (int, error) {
	for _, line := range a.Lines {
		err := into.WriteLine(line)
		if err != nil {
			return 0, err // TODO fmt.Errorf
		}
	}
	return len(a.Lines), nil
}

var _ EdDiffCommand = (Add)(nil)
