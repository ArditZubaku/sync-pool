# [sync.Pool](https://pkg.go.dev/sync#Pool)

How to use sync.Pool in Go to manage resources efficiently - by reusing expensive-to-create objects like buffers and slices, we can minimize memory allocations and reduce pressure on the garbage collector.

## Typed Wrapper for `sync.Pool`

The standard library's `sync.Pool` predates Go generics, so it only works with `any` values. Because of that, every `Get()` call requires a manual type assertion to recover the correct type. I added a generic wrapper around it to provide compile-time type safety and remove the need for explicit assertions.
