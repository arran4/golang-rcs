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
	Grammar(g *GrammarGenerator)
}

func (f *File) PseudoGrammar() string {
	g := NewGrammarGenerator()
	f.Grammar(g)
	return g.String()
}

func (g *GrammarGenerator) AddDefinition(name string, body string) {
	g.definitions[name] = fmt.Sprintf("%s := %s;", name, body)
}

func (f *File) Grammar(g *GrammarGenerator) {
	if g.visited["File"] {
		return
	}
	g.visited["File"] = true

	var fields []string
	fields = append(fields, "Head: string")
	fields = append(fields, "Branch: string")
	fields = append(fields, "Description: string")
	fields = append(fields, "Comment: string")
	fields = append(fields, "Access: bool")

	fields = append(fields, "Symbols: {Symbol}*")
	(&Symbol{}).Grammar(g)

	fields = append(fields, "AccessUsers: {string}*")

	fields = append(fields, "Locks: {Lock}*")
	(&Lock{}).Grammar(g)

	fields = append(fields, "Strict: bool")
	fields = append(fields, "StrictOnOwnLine: bool?")
	fields = append(fields, "DateYearPrefixTruncated: bool?")
	fields = append(fields, "Integrity: string")
	fields = append(fields, "Expand: string")
	fields = append(fields, "NewLine: string")
	fields = append(fields, "EndOfFileNewLineOffset: int?")

	fields = append(fields, "RevisionHeads: {RevisionHead}*")
	(&RevisionHead{}).Grammar(g)

	fields = append(fields, "RevisionContents: {RevisionContent}*")
	(&RevisionContent{}).Grammar(g)

	g.AddDefinition("File", "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
}

func (s *Symbol) Grammar(g *GrammarGenerator) {
	if g.visited["Symbol"] {
		return
	}
	g.visited["Symbol"] = true
	var fields []string
	fields = append(fields, "Name: string")
	fields = append(fields, "Revision: string")
	g.AddDefinition("Symbol", "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
}

func (l *Lock) Grammar(g *GrammarGenerator) {
	if g.visited["Lock"] {
		return
	}
	g.visited["Lock"] = true
	var fields []string
	fields = append(fields, "User: string")
	fields = append(fields, "Revision: string")
	g.AddDefinition("Lock", "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
}

func (r *RevisionHead) Grammar(g *GrammarGenerator) {
	if g.visited["RevisionHead"] {
		return
	}
	g.visited["RevisionHead"] = true

	var fields []string
	fields = append(fields, "Revision: Num")
	Num("").Grammar(g)

	fields = append(fields, "Date: DateTime")
	DateTime("").Grammar(g)

	fields = append(fields, "YearTruncated: bool?")
	fields = append(fields, "Author: ID")
	ID("").Grammar(g)

	fields = append(fields, "State: ID")
	// ID already visited

	fields = append(fields, "Branches: {Num}*")
	// Num already visited

	fields = append(fields, "NextRevision: Num")
	fields = append(fields, "CommitID: Sym")
	Sym("").Grammar(g)

	fields = append(fields, "Owner: {PhraseValue}*?")
	definePhraseValue(g)

	fields = append(fields, "Group: {PhraseValue}*?")
	fields = append(fields, "Permissions: {PhraseValue}*?")
	fields = append(fields, "Hardlinks: {PhraseValue}*?")
	fields = append(fields, "Deltatype: {PhraseValue}*?")
	fields = append(fields, "Kopt: {PhraseValue}*?")
	fields = append(fields, "Mergepoint: {PhraseValue}*?")
	fields = append(fields, "Filename: {PhraseValue}*?")
	fields = append(fields, "Username: {PhraseValue}*?")

	fields = append(fields, "NewPhrases: {NewPhrase}*?")
	(&NewPhrase{}).Grammar(g)

	g.AddDefinition("RevisionHead", "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
}

func (r *RevisionContent) Grammar(g *GrammarGenerator) {
	if g.visited["RevisionContent"] {
		return
	}
	g.visited["RevisionContent"] = true
	var fields []string
	fields = append(fields, "Revision: string")
	fields = append(fields, "Log: string")
	fields = append(fields, "Text: string")
	fields = append(fields, "PrecedingNewLinesOffset: int?")
	g.AddDefinition("RevisionContent", "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
}

func (n *NewPhrase) Grammar(g *GrammarGenerator) {
	if g.visited["NewPhrase"] {
		return
	}
	g.visited["NewPhrase"] = true
	var fields []string
	fields = append(fields, "Key: ID")
	ID("").Grammar(g)

	fields = append(fields, "Value: {PhraseValue}*")
	definePhraseValue(g)

	g.AddDefinition("NewPhrase", "{\n\t"+strings.Join(fields, ";\n\t")+";\n}")
}

func definePhraseValue(g *GrammarGenerator) {
	if g.visited["PhraseValue"] {
		return
	}
	g.visited["PhraseValue"] = true

	var options []string
	options = append(options, "SimpleString")
	SimpleString("").Grammar(g)

	options = append(options, "QuotedString")
	QuotedString("").Grammar(g)

	g.AddDefinition("PhraseValue", "{\n\t"+strings.Join(options, ";\n\t")+";\n}")
}

func (n Num) Grammar(g *GrammarGenerator) {
	if g.visited["Num"] {
		return
	}
	g.visited["Num"] = true
	g.AddDefinition("Num", "string")
}

func (id ID) Grammar(g *GrammarGenerator) {
	if g.visited["ID"] {
		return
	}
	g.visited["ID"] = true
	g.AddDefinition("ID", "string")
}

func (s Sym) Grammar(g *GrammarGenerator) {
	if g.visited["Sym"] {
		return
	}
	g.visited["Sym"] = true
	g.AddDefinition("Sym", "string")
}

func (d DateTime) Grammar(g *GrammarGenerator) {
	if g.visited["DateTime"] {
		return
	}
	g.visited["DateTime"] = true
	g.AddDefinition("DateTime", "string")
}

func (s SimpleString) Grammar(g *GrammarGenerator) {
	if g.visited["SimpleString"] {
		return
	}
	g.visited["SimpleString"] = true
	g.AddDefinition("SimpleString", "string")
}

func (s QuotedString) Grammar(g *GrammarGenerator) {
	if g.visited["QuotedString"] {
		return
	}
	g.visited["QuotedString"] = true
	g.AddDefinition("QuotedString", "string")
}
