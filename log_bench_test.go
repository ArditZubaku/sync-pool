package main

import (
	"bytes"
	"io"
	"testing"
	"time"
)

// go test -bench=. -benchmem
// Results:
// goos: linux
// goarch: amd64
// pkg: github.com/ArditZubaku/go-sync-pool
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkLogNoPool-12            6689988               183.2 ns/op            72 B/op          2 allocs/op
// BenchmarkLogWithPool-12          7561688               155.2 ns/op             8 B/op          1 allocs/op
// PASS
// ok      github.com/ArditZubaku/go-sync-pool     2.405s

func logNoPool(w io.Writer, val string) {
	var b bytes.Buffer
	b.WriteString(time.Now().Format("15:04:05"))
	b.WriteString(" : ")
	b.WriteString(val)
	w.Write(b.Bytes())
}

var bufferPool = NewTypedPool(
	func() *bytes.Buffer {
		return new(bytes.Buffer)
	},
)

func logWithPool(w io.Writer, val string) {
	b := bufferPool.Get()
	b.Reset()

	b.WriteString(time.Now().Format("15:04:05"))
	b.WriteString(" : ")
	b.WriteString(val)
	w.Write(b.Bytes())

	bufferPool.Put(b)
}

func BenchmarkLogNoPool(b *testing.B) {
	for b.Loop() {
		logNoPool(io.Discard, "some log message")
	}
}

func BenchmarkLogWithPool(b *testing.B) {
	for b.Loop() {
		logWithPool(io.Discard, "some log message")
	}
}
