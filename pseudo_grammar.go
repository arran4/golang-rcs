package rcs

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func (f *File) PseudoGrammar() string {
	g := NewGrammarGenerator()
	g.Consume(reflect.TypeOf(f))
	return g.String()
}

type GrammarGenerator struct {
	definitions              map[string]string
	visited                  map[reflect.Type]bool
	interfaceImplementations map[reflect.Type][]reflect.Type
	rootPkgPath              string
}

func NewGrammarGenerator() *GrammarGenerator {
	return &GrammarGenerator{
		definitions: make(map[string]string),
		visited:     make(map[reflect.Type]bool),
		interfaceImplementations: map[reflect.Type][]reflect.Type{
			reflect.TypeOf((*PhraseValue)(nil)).Elem(): {
				reflect.TypeOf(SimpleString("")),
				reflect.TypeOf(QuotedString("")),
			},
		},
	}
}

func (g *GrammarGenerator) Consume(t reflect.Type) {
	if g.rootPkgPath == "" {
		g.rootPkgPath = t.Elem().PkgPath()
	}
	g.walk(t)
}

func (g *GrammarGenerator) walk(t reflect.Type) {
	base := g.unwrap(t)

	if g.visited[base] {
		return
	}
	g.visited[base] = true

	if base.Kind() == reflect.Struct {
		g.processStruct(base)
	} else if base.Kind() == reflect.Interface {
		g.processInterface(base)
	} else if base.Name() != "" && (base.PkgPath() == g.rootPkgPath) {
		g.processBasic(base)
	}
}

func (g *GrammarGenerator) unwrap(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return t
}

func (g *GrammarGenerator) processStruct(t reflect.Type) {
	if t.PkgPath() != g.rootPkgPath && t.PkgPath() != "" {
		return
	}

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldType := field.Type
		displayType := g.formatType(fieldType)

		// Recursively walk dependencies
		g.collectDependencies(fieldType)

		tag := field.Tag.Get("json")
		if strings.Contains(tag, "omitempty") {
			displayType += "?"
		}

		fields = append(fields, fmt.Sprintf("%s: %s;", field.Name, displayType))
	}
	g.definitions[t.Name()] = fmt.Sprintf("%s := {\n\t%s\n};", t.Name(), strings.Join(fields, "\n\t"))
}

func (g *GrammarGenerator) processInterface(t reflect.Type) {
	if impls, ok := g.interfaceImplementations[t]; ok {
		var implNames []string
		for _, impl := range impls {
			implNames = append(implNames, fmt.Sprintf("%s;", impl.Name()))
			g.collectDependencies(impl)
		}
		g.definitions[t.Name()] = fmt.Sprintf("%s := {\n\t%s\n};", t.Name(), strings.Join(implNames, "\n\t"))
	} else {
		g.definitions[t.Name()] = fmt.Sprintf("%s := interface;", t.Name())
	}
}

func (g *GrammarGenerator) processBasic(t reflect.Type) {
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Interface {
		g.definitions[t.Name()] = fmt.Sprintf("%s := %s;", t.Name(), t.Kind().String())
	}
}

func (g *GrammarGenerator) collectDependencies(t reflect.Type) {
	base := g.unwrap(t)

	if base.Kind() == reflect.Interface {
		if implementations, ok := g.interfaceImplementations[base]; ok {
			for _, impl := range implementations {
				g.walk(impl)
			}
		}
	}

	if base.PkgPath() == g.rootPkgPath || base.PkgPath() == "" {
		g.walk(base)
	}
}

func (g *GrammarGenerator) formatType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return g.formatType(t.Elem())
	}
	if t.Kind() == reflect.Slice {
		return "{" + g.formatType(t.Elem()) + "}*"
	}
	return t.Name()
}

func (g *GrammarGenerator) String() string {
	var keys []string
	for k := range g.definitions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder

	// Legend
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
