package main

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

var buffPool = NewTypedPool(
	func() *bytes.Buffer {
		fmt.Println("New buffer is created")
		return new(bytes.Buffer)
	},
)

func log(w io.Writer, val string) {
	// var b bytes.Buffer
	b := buffPool.Get()
	b.Reset()

	b.WriteString(time.Now().Format("15:04:05"))
	b.WriteString(" : ")
	b.WriteString(val)
	w.Write(b.Bytes())

	buffPool.Put(b)
}
