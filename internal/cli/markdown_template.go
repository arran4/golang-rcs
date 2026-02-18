package cli

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed markdown.tmpl
var markdownTmpl string

var markdownTemplate = template.Must(template.New("markdown").Funcs(template.FuncMap{
	"quote": func(content string) string {
		if content == "" {
			return "> \n"
		}
		var sb strings.Builder
		lines := strings.Split(content, "\n")
		// If the last line is empty (due to trailing newline in Split), ignore it loop?
		// strings.Split("a\n", "\n") -> ["a", ""]
		// We want "> a\n"

		for i, line := range lines {
			if i == len(lines)-1 && line == "" {
				continue
			}
			sb.WriteString("> " + line + "\n")
		}
		return sb.String()
	},
}).Parse(markdownTmpl))
