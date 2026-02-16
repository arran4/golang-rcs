package hashline

import (
	"math/rand"
	"strings"
	"testing"
)

func BenchmarkGenerateEdDiffFromLines(b *testing.B) {
	from := make([]string, 1000)
	to := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		from[i] = randString(10)
		to[i] = randString(10)
		if i%2 == 0 {
			to[i] = from[i]
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateEdDiffFromLines(from, to)
	}
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(letterBytes[rand.Intn(len(letterBytes))])
	}
	return sb.String()
}
