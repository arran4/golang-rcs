package rcs

import (
	"strings"
	"testing"
)

func BenchmarkScanStrings(b *testing.B) {
	token := "token"
	repeatCount := 10000
	input := strings.Repeat(token, repeatCount)

	r := strings.NewReader(input)
	s := NewScanner(r)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := ScanStrings(s, token)
		if err != nil {
			// Reset scanner if EOF or error
			r.Reset(input)
			s = NewScanner(r)
			err = ScanStrings(s, token)
			if err != nil {
				b.Fatalf("ScanStrings failed after reset: %v", err)
			}
		}
	}
}
