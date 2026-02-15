package rcs

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func (f *File) PseudoGrammar() string {
	return GeneratePseudoGrammar(reflect.TypeOf(f))
}

func GeneratePseudoGrammar(t reflect.Type) string {
	definitions := make(map[string]string)
	visited := make(map[reflect.Type]bool)
	queue := []reflect.Type{t}

	// We want to process types defined in this package.
	// We can use the PkgPath of the root type to filter others.
	rootPkgPath := t.Elem().PkgPath()

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// Unwrap ptr/slice to get base type
		base := curr
		for base.Kind() == reflect.Ptr || base.Kind() == reflect.Slice {
			base = base.Elem()
		}

		if visited[base] {
			continue
		}
		visited[base] = true

		// Only generate definitions for structs in the same package (or if PkgPath is empty and it's not basic)
		// Basic types have empty PkgPath but we check Kind.
		if base.Kind() == reflect.Struct {
			if base.PkgPath() != rootPkgPath && base.PkgPath() != "" {
				continue
			}

			var fields []string
			for i := 0; i < base.NumField(); i++ {
				field := base.Field(i)
				if field.PkgPath != "" {
					// Skip unexported fields?
					// RCS parser uses exported fields.
					// But if there are unexported fields, they might be implementation details.
					// Let's include all fields for now or check PkgPath (empty for exported).
					// Wait, reflect.StructField.PkgPath is empty for exported fields.
					continue
				}

				fieldType := field.Type
				displayType := formatType(fieldType)
				collectDependencies(fieldType, &queue, rootPkgPath)

				tag := field.Tag.Get("json")
				if strings.Contains(tag, "omitempty") {
					displayType += "?"
				}

				fields = append(fields, fmt.Sprintf("%s: %s;", field.Name, displayType))
			}
			definitions[base.Name()] = fmt.Sprintf("%s := {\n\t%s\n};", base.Name(), strings.Join(fields, "\n\t"))
		}
	}

	var keys []string
	for k := range definitions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		sb.WriteString(definitions[k])
		if i < len(keys)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func formatType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return formatType(t.Elem())
	}
	if t.Kind() == reflect.Slice {
		return "{" + formatType(t.Elem()) + "}*"
	}
	return t.Name()
}

func collectDependencies(t reflect.Type, queue *[]reflect.Type, rootPkgPath string) {
	base := t
	for base.Kind() == reflect.Ptr || base.Kind() == reflect.Slice {
		base = base.Elem()
	}
	if base.Kind() == reflect.Struct {
		if base.PkgPath() == rootPkgPath || base.PkgPath() == "" {
			*queue = append(*queue, base)
		}
	}
}
