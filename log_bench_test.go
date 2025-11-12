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
// BenchmarkLogNoPool-14           12760836                93.24 ns/op           72 B/op          2 allocs/op
// BenchmarkLogWithPool-14         16930678                70.34 ns/op            0 B/op          0 allocs/op
// PASS
// ok      github.com/ArditZubaku/go-sync-pool     2.575s

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

	b.Write(time.Now().AppendFormat(b.AvailableBuffer(), "15:04:05"))
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
