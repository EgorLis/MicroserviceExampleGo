package benchstr

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func makeParts(n int) []string {
	parts := make([]string, 0, n*2)
	for i := 0; i < n; i++ {
		parts = append(parts, "k")
		parts = append(parts, "v")
	}
	return parts
}

func concatPlus(parts []string) string {
	s := ""
	for i := 0; i < len(parts); i += 2 {
		if i > 0 {
			s += "&"
		}
		s += parts[i] + "=" + parts[i+1]
	}
	return s
}

func concatSprintf(parts []string) string {
	s := ""
	for i := 0; i < len(parts); i += 2 {
		if i > 0 {
			s = fmt.Sprintf("%s&%s=%s", s, parts[i], parts[i+1])
		} else {
			s = fmt.Sprintf("%s=%s", parts[i], parts[i+1])
		}
	}
	return s
}

func concatBuilder(parts []string) string {
	var b strings.Builder
	for i := 0; i < len(parts); i += 2 {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(parts[i])
		b.WriteString("=")
		b.WriteString(parts[i+1])
	}
	return b.String()
}

func concatBuffer(parts []string) string {
	var buf bytes.Buffer
	for i := 0; i < len(parts); i += 2 {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(parts[i])
		buf.WriteString("=")
		buf.WriteString(parts[i+1])
	}
	return buf.String()
}

func BenchmarkPlus_4pairs(b *testing.B) {
	parts := makeParts(4)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatPlus(parts)
	}
}

func BenchmarkSprintf_4pairs(b *testing.B) {
	parts := makeParts(4)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatSprintf(parts)
	}
}

func BenchmarkBuilder_4pairs(b *testing.B) {
	parts := makeParts(4)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatBuilder(parts)
	}
}

func BenchmarkBuffer_4pairs(b *testing.B) {
	parts := makeParts(4)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatBuffer(parts)
	}
}

func BenchmarkPlus_50pairs(b *testing.B) {
	parts := makeParts(50)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatPlus(parts)
	}
}

func BenchmarkSprintf_50pairs(b *testing.B) {
	parts := makeParts(50)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatSprintf(parts)
	}
}

func BenchmarkBuilder_50pairs(b *testing.B) {
	parts := makeParts(50)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatBuilder(parts)
	}
}

func BenchmarkBuffer_50pairs(b *testing.B) {
	parts := makeParts(50)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatBuffer(parts)
	}
}
