# [sync.Pool](https://pkg.go.dev/sync#Pool)

How to use sync.Pool in Go to manage resources efficiently - by reusing expensive-to-create objects like buffers and slices, we can minimize memory allocations and reduce pressure on the garbage collector.

## Typed Wrapper for `sync.Pool`

The standard library's `sync.Pool` predates Go generics, so it only works with `any` values. Because of that, every `Get()` call requires a manual type assertion to recover the correct type. I added a generic wrapper around it to provide compile-time type safety and remove the need for explicit assertions.

## Benchmark Results

### Performance Comparison

Running basic performance benchmarks:

```bash
go test -bench=. -benchmem
```

Results show significant improvements when using sync.Pool:

```
goos: linux
goarch: amd64
pkg: github.com/ArditZubaku/go-sync-pool
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkLogNoPool-12            6689988               183.2 ns/op            72 B/op          2 allocs/op
BenchmarkLogWithPool-12          7561688               155.2 ns/op             8 B/op          1 allocs/op
PASS
```

**Key Improvements:**

- **15% faster execution** (183.2 ns/op → 155.2 ns/op)
- **89% reduction in memory allocations** (72 B/op → 8 B/op)
- **50% fewer allocations per operation** (2 allocs/op → 1 allocs/op)

### Memory Profiling Analysis

To understand where memory allocations occur, we can use Go's built-in profiling tools.

#### Without sync.Pool

```bash
go test -bench=BenchmarkLogNoPool -benchmem -memprofile=mem.out
go tool pprof -alloc_space ./go-sync-pool.test mem.out
```

```
(pprof) top
Showing nodes accounting for 428.02MB, 99.65% of 429.53MB total
Dropped 22 nodes (cum <= 2.15MB)
      flat  flat%   sum%        cum   cum%
382.02MB 88.94% 88.94% 382.02MB 88.94%  bytes.(*Buffer).grow
    46MB 10.71% 99.65%    46.50MB 10.83%  time.Time.Format
         0     0% 99.65% 382.02MB 88.94%  bytes.(*Buffer).WriteString
         0     0% 99.65% 428.53MB 99.77%  github.com/ArditZubaku/go-sync-pool.BenchmarkLogNoPool
         0     0% 99.65% 428.53MB 99.77%  github.com/ArditZubaku/go-sync-pool.logNoPool
         0     0% 99.65% 428.53MB 99.77%  testing.(*B).run1.func1
         0     0% 99.65% 428.53MB 99.77%  testing.(*B).runN
```

**Analysis:** The major memory consumer is `bytes.(*Buffer).grow` (88.94% of total allocations), which happens because we create a new buffer for every log operation and it needs to grow to accommodate the data.

#### With sync.Pool

```bash
go test -bench=BenchmarkLogWithPool -memprofile=mem.out
go tool pprof -alloc_space ./go-sync-pool.test mem.out
```

```
(pprof) top
Showing nodes accounting for 57MB, 100% of 57MB total
Showing top 10 nodes out of 23
      flat  flat%   sum%        cum   cum%
   55.50MB 97.37% 97.37%    55.50MB 97.37%  time.Time.Format
    0.50MB  0.88% 98.24%     0.50MB  0.88%  runtime.allocm
    0.50MB  0.88% 99.12%     0.50MB  0.88%  sync.(*Pool).pinSlow
    0.50MB  0.88%   100%     0.50MB  0.88%  runtime.acquireSudog
         0     0%   100%     0.50MB  0.88%  github.com/ArditZubaku/go-sync-pool.(*TypedPool[go.shape.*uint8]).Put (inline)
         0     0%   100%       56MB 98.24%  github.com/ArditZubaku/go-sync-pool.BenchmarkLogWithPool
         0     0%   100%       56MB 98.24%  github.com/ArditZubaku/go-sync-pool.logWithPool
         0     0%   100%     0.50MB  0.88%  runtime.gcBgMarkWorker
         0     0%   100%     0.50MB  0.88%  runtime.gcMarkDone
         0     0%   100%     0.50MB  0.88%  runtime.mstart
```

**Analysis:**

- **87% reduction in total memory usage** (429.53MB → 57MB)
- The `bytes.(*Buffer).grow` allocation completely disappeared
- Most memory usage now comes from `time.Time.Format` (97.37%), which is unavoidable
- Small overhead from sync.Pool management (`sync.(*Pool).pinSlow`)

### Key Takeaways

1. **sync.Pool eliminates buffer growth allocations** - The most expensive part of the original implementation
2. **Memory reuse is highly effective** - 87% reduction in total memory allocations
3. **Performance gains compound** - Fewer allocations mean less GC pressure and faster execution
4. **Time formatting remains the bottleneck** - Further optimizations could target timestamp generation

The memory profiling clearly demonstrates why sync.Pool is effective: it eliminates the need to repeatedly allocate and grow buffers, which was consuming nearly 90% of the memory in the original implementation.
