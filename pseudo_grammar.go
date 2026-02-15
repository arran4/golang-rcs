package rcs

import (
	"fmt"
	"sort"
	"strings"
)

type GrammarGenerator struct {
	visited     map[string]bool
	definitions map[string]string
}

func NewGrammarGenerator() *GrammarGenerator {
	return &GrammarGenerator{
		visited:     make(map[string]bool),
		definitions: make(map[string]string),
	}
}

func (g *GrammarGenerator) String() string {
	var keys []string
	for k := range g.definitions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("Legend:\n")
	sb.WriteString("  Type := { ... }  : Sequential structure (all fields required unless marked)\n")
	sb.WriteString("  Type := { A; B; }: Choice/Interface implementation (A or B)\n")
	sb.WriteString("  {Type}*          : Sequence of zero or more Type\n")
	sb.WriteString("  Type?            : Optional field\n")
	sb.WriteString("\n")

	for i, k := range keys {
		sb.WriteString(g.definitions[k])
		if i < len(keys)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

type Grammerer interface {
	Grammar(g *GrammarGenerator) string
}

func (f *File) PseudoGrammar() string {
	g := NewGrammarGenerator()
	f.Grammar(g)
	return g.String()
}

func (g *GrammarGenerator) AddDefinition(name string, body string) {
	g.definitions[name] = fmt.Sprintf("%s := %s;", name, body)
}

func (f *File) Grammar(g *GrammarGenerator) string {
	name := "File"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true

	var fields []string
	fields = append(fields, "Head: string")
	fields = append(fields, "Branch: string")
	fields = append(fields, "Access: bool")
	fields = append(fields, "AccessUsers: {string}*")

	fields = append(fields, fmt.Sprintf("Symbols: {%s}*", (&Symbol{}).Grammar(g)))
	fields = append(fields, fmt.Sprintf("Locks: {%s}*", (&Lock{}).Grammar(g)))

	fields = append(fields, "Strict: bool")
	fields = append(fields, "Integrity: string")
	fields = append(fields, "Comment: string")
	fields = append(fields, "Expand: string")

	fields = append(fields, fmt.Sprintf("RevisionHeads: {%s}*", (&RevisionHead{}).Grammar(g)))
	fields = append(fields, "Description: string")
	fields = append(fields, fmt.Sprintf("RevisionContents: {%s}*", (&RevisionContent{}).Grammar(g)))

	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func (s *Symbol) Grammar(g *GrammarGenerator) string {
	name := "Symbol"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	var fields []string
	fields = append(fields, "Name: string")
	fields = append(fields, "Revision: string")
	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func (l *Lock) Grammar(g *GrammarGenerator) string {
	name := "Lock"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	var fields []string
	fields = append(fields, "User: string")
	fields = append(fields, "Revision: string")
	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func (r *RevisionHead) Grammar(g *GrammarGenerator) string {
	name := "RevisionHead"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true

	var fields []string
	fields = append(fields, fmt.Sprintf("Revision: %s", Num("").Grammar(g)))
	fields = append(fields, fmt.Sprintf("Date: %s", DateTime("").Grammar(g)))
	fields = append(fields, fmt.Sprintf("Author: %s", ID("").Grammar(g)))
	fields = append(fields, fmt.Sprintf("State: %s", ID("").Grammar(g)))

	fields = append(fields, fmt.Sprintf("Branches: {%s}*", Num("").Grammar(g)))
	fields = append(fields, fmt.Sprintf("NextRevision: %s", Num("").Grammar(g)))
	fields = append(fields, fmt.Sprintf("CommitID: %s", Sym("").Grammar(g)))

	pvName := definePhraseValue(g)
	fields = append(fields, fmt.Sprintf("Owner: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Group: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Permissions: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Hardlinks: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Deltatype: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Kopt: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Mergepoint: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Filename: {%s}*?", pvName))
	fields = append(fields, fmt.Sprintf("Username: {%s}*?", pvName))

	fields = append(fields, fmt.Sprintf("NewPhrases: {%s}*?", (&NewPhrase{}).Grammar(g)))

	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func (r *RevisionContent) Grammar(g *GrammarGenerator) string {
	name := "RevisionContent"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	var fields []string
	fields = append(fields, "Revision: string")
	fields = append(fields, "Log: string")
	fields = append(fields, "Text: string")
	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func (n *NewPhrase) Grammar(g *GrammarGenerator) string {
	name := "NewPhrase"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	var fields []string
	fields = append(fields, fmt.Sprintf("Key: %s", ID("").Grammar(g)))
	fields = append(fields, fmt.Sprintf("Value: {%s}*", definePhraseValue(g)))

	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func definePhraseValue(g *GrammarGenerator) string {
	name := "PhraseValue"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true

	var options []string
	options = append(options, SimpleString("").Grammar(g))
	options = append(options, QuotedString("").Grammar(g))

	g.AddDefinition(name, "{\n\t"+strings.Join(options, ";\n\t")+";\n}")
	return name
}

func (n Num) Grammar(g *GrammarGenerator) string {
	name := "Num"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "string")
	return name
}

func (id ID) Grammar(g *GrammarGenerator) string {
	name := "ID"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "string")
	return name
}

func (s Sym) Grammar(g *GrammarGenerator) string {
	name := "Sym"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "string")
	return name
}

func (d DateTime) Grammar(g *GrammarGenerator) string {
	name := "DateTime"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "string")
	return name
}

func (s SimpleString) Grammar(g *GrammarGenerator) string {
	name := "SimpleString"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "string")
	return name
}

func (s QuotedString) Grammar(g *GrammarGenerator) string {
	name := "QuotedString"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "string")
	return name
}
