package testing

import (
	"io"
	"strings"
)

type StringLineReader struct {
	lines []string
	pos   int
}

func NewStringLineReader(content interface{}) *StringLineReader {
	var lines []string
	switch c := content.(type) {
	case string:
		lines = strings.Split(c, "\n")
		if c == "" {
			lines = []string{}
		} else {
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
		}
	case []string:
		lines = c
	}
	return &StringLineReader{lines: lines}
}

func (r *StringLineReader) ReadLine() (string, error) {
	if r.pos >= len(r.lines) {
		return "", io.EOF
	}
	line := r.lines[r.pos]
	r.pos++
	return line, nil
}

type StringLineWriter struct {
	lines []string
}

func (w *StringLineWriter) WriteLine(line string) error {
	w.lines = append(w.lines, line)
	return nil
}

func (w *StringLineWriter) Lines() []string {
	return w.lines
}

func (w *StringLineWriter) String() string {
	if len(w.lines) == 0 {
		return ""
	}
	return strings.Join(w.lines, "\n") + "\n"
}
