package rcs

import (
	"strings"
	"testing"
)

var sink string

func BenchmarkAtQuote(b *testing.B) {
	s := "test@string@with@ats"
	for i := 0; i < b.N; i++ {
		sink = AtQuote(s)
	}
}

func BenchmarkAtQuote_Long(b *testing.B) {
	s := strings.Repeat("test@string@with@ats", 100)
	for i := 0; i < b.N; i++ {
		sink = AtQuote(s)
	}
}

func BenchmarkWriteAtQuote_Stream(b *testing.B) {
	s := "test@string@with@ats"
	var sb strings.Builder
	for i := 0; i < b.N; i++ {
		sb.Reset()
		_, _ = WriteAtQuote(&sb, s)
	}
}

func BenchmarkWriteAtQuote_Stream_Long(b *testing.B) {
	s := strings.Repeat("test@string@with@ats", 100)
	var sb strings.Builder
	for i := 0; i < b.N; i++ {
		sb.Reset()
		_, _ = WriteAtQuote(&sb, s)
	}
}

// Simulating usage in String() methods
func BenchmarkStringer_Usage_AtQuote(b *testing.B) {
	s := "test@string@with@ats"
	var sb strings.Builder
	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.WriteString("prefix")
		sb.WriteString(AtQuote(s))
		sb.WriteString("suffix")
	}
}

func BenchmarkStringer_Usage_WriteAtQuote(b *testing.B) {
	s := "test@string@with@ats"
	var sb strings.Builder
	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.WriteString("prefix")
		_, _ = WriteAtQuote(&sb, s)
		sb.WriteString("suffix")
	}
}

func BenchmarkRevisionHeadStringWithNewLine(b *testing.B) {
	rh := &RevisionHead{
		Revision:     "1.1",
		Date:         "2023.10.26.12.00.00",
		Author:       "user",
		State:        "Exp",
		Branches:     []Num{"1.1.1.1"},
		NextRevision: "1.2",
	}
	nl := "\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rh.StringWithNewLine(nl)
	}
}

func BenchmarkRevisionHeadStringWithNewLine_Spaces(b *testing.B) {
	rh := &RevisionHead{
		RevisionHeadFormattingOptions: RevisionHeadFormattingOptions{
			DateSeparatorSpaces:      2,
			DateAuthorSpacingSpaces:  3,
			AuthorStateSpacingSpaces: 4,
		},
		Revision:     "1.1",
		Date:         "2023.10.26.12.00.00",
		Author:       "user",
		State:        "Exp",
		Branches:     []Num{"1.1.1.1"},
		NextRevision: "1.2",
	}
	nl := "\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rh.StringWithNewLine(nl)
	}
}
