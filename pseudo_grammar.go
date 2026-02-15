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
	sb.WriteString("  Type := Unordered { ... } : Fields can appear in any order\n")
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
	addLexerDefinitions(g)
	f.Grammar(g)
	return g.String()
}

func addLexerDefinitions(g *GrammarGenerator) {
	g.AddDefinition("digit", "\"0\"..\"9\"")
	g.AddDefinition("idchar", "any visible graphic character except special")
	g.AddDefinition("special", "\"$\" | \",\" | \".\" | \":\" | \";\" | \"@\"")
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

	var unorderedFields []string
	unorderedFields = append(unorderedFields, "Head: Num?")
	unorderedFields = append(unorderedFields, "Branch: Num?")
	unorderedFields = append(unorderedFields, "Access: bool?")
	unorderedFields = append(unorderedFields, "AccessUsers: {ID}*")

	unorderedFields = append(unorderedFields, fmt.Sprintf("Symbols: {%s}*", (&Symbol{}).Grammar(g)))
	unorderedFields = append(unorderedFields, fmt.Sprintf("Locks: {%s}*", (&Lock{}).Grammar(g)))

	unorderedFields = append(unorderedFields, "Strict: bool?")
	unorderedFields = append(unorderedFields, "Integrity: @String@?")
	unorderedFields = append(unorderedFields, "Comment: @String@?")
	unorderedFields = append(unorderedFields, "Expand: @String@?")

	var fields []string
	fields = append(fields, "Admin: Unordered {\n\t\t"+strings.Join(unorderedFields, ";\n\t\t")+";\n\t}")

	fields = append(fields, fmt.Sprintf("RevisionHeads: {%s}*", (&RevisionHead{}).Grammar(g)))
	fields = append(fields, "Description: @String@")
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
	fields = append(fields, "Name: Sym")
	Sym("").Grammar(g)
	fields = append(fields, "Revision: Num")
	Num("").Grammar(g)
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
	fields = append(fields, "User: ID")
	ID("").Grammar(g)
	fields = append(fields, "Revision: Num")
	Num("").Grammar(g)
	g.AddDefinition(name, "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
	return name
}

func (r *RevisionHead) Grammar(g *GrammarGenerator) string {
	name := "RevisionHead"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true

	var parts []string
	parts = append(parts, fmt.Sprintf("Revision: %s", Num("").Grammar(g)))

	var unordered []string
	unordered = append(unordered, fmt.Sprintf("Date: %s", DateTime("").Grammar(g)))
	unordered = append(unordered, fmt.Sprintf("Author: %s", ID("").Grammar(g)))
	unordered = append(unordered, fmt.Sprintf("State: %s", ID("").Grammar(g)))
	unordered = append(unordered, fmt.Sprintf("Branches: {%s}*", Num("").Grammar(g)))
	unordered = append(unordered, fmt.Sprintf("NextRevision: %s", Num("").Grammar(g)))
	unordered = append(unordered, fmt.Sprintf("CommitID: %s", Sym("").Grammar(g)))

	pvName := definePhraseValue(g)
	unordered = append(unordered, fmt.Sprintf("Owner: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Group: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Permissions: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Hardlinks: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Deltatype: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Kopt: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Mergepoint: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Filename: {%s}*?", pvName))
	unordered = append(unordered, fmt.Sprintf("Username: {%s}*?", pvName))

	unordered = append(unordered, fmt.Sprintf("NewPhrases: {%s}*?", (&NewPhrase{}).Grammar(g)))

	parts = append(parts, "Meta: Unordered {\n\t\t"+strings.Join(unordered, ";\n\t\t")+";\n\t}")

	g.AddDefinition(name, "{\n\t"+strings.Join(parts, ";\n\t")+";\n}")
	return name
}

func (r *RevisionContent) Grammar(g *GrammarGenerator) string {
	name := "RevisionContent"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true

	var parts []string
	parts = append(parts, fmt.Sprintf("Revision: %s", Num("").Grammar(g)))

	var unordered []string
	unordered = append(unordered, "Log: @String@")
	unordered = append(unordered, "Text: @String@")

	parts = append(parts, "Data: Unordered {\n\t\t"+strings.Join(unordered, ";\n\t\t")+";\n\t}")

	g.AddDefinition(name, "{\n\t"+strings.Join(parts, ";\n\t")+";\n}")
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
	g.AddDefinition(name, "{digit | .}+")
	return name
}

func (id ID) Grammar(g *GrammarGenerator) string {
	name := "ID"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "{idchar | .}+")
	return name
}

func (s Sym) Grammar(g *GrammarGenerator) string {
	name := "Sym"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "{idchar}+")
	return name
}

func (d DateTime) Grammar(g *GrammarGenerator) string {
	name := "DateTime"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "{digit | .}+")
	return name
}

func (s SimpleString) Grammar(g *GrammarGenerator) string {
	name := "SimpleString"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "{idchar | .}+")
	return name
}

func (s QuotedString) Grammar(g *GrammarGenerator) string {
	name := "QuotedString"
	if g.visited[name] {
		return name
	}
	g.visited[name] = true
	g.AddDefinition(name, "\"@\" {any_char}* \"@\" (doubled @)")
	return name
}
