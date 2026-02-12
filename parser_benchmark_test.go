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
