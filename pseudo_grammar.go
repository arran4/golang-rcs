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
	rootPkgPath := t.Elem().PkgPath()

	// Seed interfaces implementations
	// This is manual because we can't find implementations dynamically easily
	// If you add more interface implementations, please add them here.
	interfaceImplementations := map[reflect.Type][]reflect.Type{
		reflect.TypeOf((*PhraseValue)(nil)).Elem(): {
			reflect.TypeOf(SimpleString("")),
			reflect.TypeOf(QuotedString("")),
		},
	}

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

		if base.Kind() == reflect.Struct {
			if base.PkgPath() != rootPkgPath && base.PkgPath() != "" {
				continue
			}

			var fields []string
			for i := 0; i < base.NumField(); i++ {
				field := base.Field(i)
				if field.PkgPath != "" {
					continue
				}

				fieldType := field.Type
				displayType := formatType(fieldType)
				collectDependencies(fieldType, &queue, rootPkgPath, interfaceImplementations)

				tag := field.Tag.Get("json")
				if strings.Contains(tag, "omitempty") {
					displayType += "?"
				}

				fields = append(fields, fmt.Sprintf("%s: %s;", field.Name, displayType))
			}
			definitions[base.Name()] = fmt.Sprintf("%s := {\n\t%s\n};", base.Name(), strings.Join(fields, "\n\t"))
		} else if base.Kind() == reflect.Interface {
			if impls, ok := interfaceImplementations[base]; ok {
				var implNames []string
				for _, impl := range impls {
					implNames = append(implNames, fmt.Sprintf("%s;", impl.Name()))
					collectDependencies(impl, &queue, rootPkgPath, interfaceImplementations)
				}
				definitions[base.Name()] = fmt.Sprintf("%s := {\n\t%s\n};", base.Name(), strings.Join(implNames, "\n\t"))
			} else {
				definitions[base.Name()] = fmt.Sprintf("%s := interface;", base.Name())
			}
		} else if base.Name() != "" && (base.PkgPath() == rootPkgPath) {
			// Named basic types (e.g. ID, Num, SimpleString)
			// Only if defined in this package
			if base.Kind() != reflect.Struct && base.Kind() != reflect.Interface {
				definitions[base.Name()] = fmt.Sprintf("%s := %s;", base.Name(), base.Kind().String())
			}
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

func collectDependencies(t reflect.Type, queue *[]reflect.Type, rootPkgPath string, impls map[reflect.Type][]reflect.Type) {
	base := t
	for base.Kind() == reflect.Ptr || base.Kind() == reflect.Slice {
		base = base.Elem()
	}

	// Add implementation dependencies if it's an interface
	if base.Kind() == reflect.Interface {
		if implementations, ok := impls[base]; ok {
			*queue = append(*queue, implementations...)
		}
	}

	if base.PkgPath() == rootPkgPath || base.PkgPath() == "" {
		*queue = append(*queue, base)
	}
}
