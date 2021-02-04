package convert

import "testing"

func BenchmarkStr2Bytes(b *testing.B) {
	b.Run("Default", func(b *testing.B) {
		b.ReportAllocs()
		s := "hello world"
		for i := 0; i < b.N; i++ {
			_ = []byte(s)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		s := "hello world"
		for i := 0; i < b.N; i++ {
			_ = StrToBytes(s)
		}
	})
}

func BenchmarkBytesToStr(b *testing.B) {
	b.Run("Default", func(b *testing.B) {
		b.ReportAllocs()
		bs := []byte("hello world")
		for i := 0; i < b.N; i++ {
			_ = string(bs)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		bs := []byte("hello world")
		for i := 0; i < b.N; i++ {
			_ = BytesToStr(bs)
		}
	})
}
