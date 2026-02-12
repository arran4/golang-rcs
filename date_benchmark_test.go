package rcs

import (
	"testing"
	"time"
)

func BenchmarkParseYearDOY(b *testing.B) {
	input := "2018-110"
	now := time.Now()
	for i := 0; i < b.N; i++ {
		_, _ = ParseDate(input, now, time.UTC)
	}
}

func BenchmarkParseYearWeekDow(b *testing.B) {
	input := "2018-w16-5"
	now := time.Now()
	for i := 0; i < b.N; i++ {
		_, _ = ParseDate(input, now, time.UTC)
	}
}
