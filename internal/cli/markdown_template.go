package cli

import (
	"strings"
	"text/template"
)

var markdownTemplate = template.Must(template.New("markdown").Funcs(template.FuncMap{
	"fenced": func(content string, lang string) string {
		fence := "```"
		// If content contains fence, extend it
		for strings.Contains(content, fence) {
			fence += "`"
		}
		// Fenced block usually has newline after opening fence, then content.
		// If content is not empty and doesn't end with newline, add one before closing fence.
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return fence + lang + "\n" + content + fence + "\n"
	},
}).Parse(`# RCS File

## Header

{{if .Head}}* Head: {{.Head}}
{{end}}{{if .Branch}}* Branch: {{.Branch}}
{{end}}{{if .AccessUsers}}* Access:
{{range .AccessUsers}}  * {{.}}
{{end}}{{end}}{{if .Symbols}}* Symbols:
{{range .Symbols}}  * {{.Name}}: {{.Revision}}
{{end}}{{end}}{{if .Locks}}* Locks:
{{range .Locks}}  * {{.User}}: {{.Revision}}
{{end}}{{end}}{{if .Strict}}* Strict: true
{{end}}{{if .Comment}}* Comment: {{.Comment}}
{{end}}{{if .Expand}}* Expand: {{.Expand}}
{{end}}
## Description

{{fenced .Description "text"}}
## Revisions
{{range $i, $rp := .Revisions}}
### {{$rp.Head.Revision}}

* Date: {{$rp.Head.Date}}
* Author: {{$rp.Head.Author}}
* State: {{$rp.Head.State}}
{{if $rp.Head.Branches}}* Branches: {{range $j, $b := $rp.Head.Branches}}{{if $j}} {{end}}{{$b}}{{end}}
{{end}}{{if $rp.Head.NextRevision}}* Next: {{$rp.Head.NextRevision}}
{{end}}{{if $rp.Head.CommitID}}* CommitID: {{$rp.Head.CommitID}}
{{end}}

#### Log

{{fenced $rp.Content.Log "text"}}

#### Text

{{fenced $rp.Content.Text "text"}}
{{end}}`))
