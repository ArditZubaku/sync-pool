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

## Key Takeaways

1. **sync.Pool eliminates buffer growth allocations** - The most expensive part of the original implementation
2. **Memory reuse is highly effective** - 87% reduction in total memory allocations
3. **Performance gains compound** - Fewer allocations mean less GC pressure and faster execution
4. **Time formatting remains the bottleneck** - Further optimizations could target timestamp generation

The memory profiling clearly demonstrates why sync.Pool is effective: it eliminates the need to repeatedly allocate and grow buffers, which was consuming nearly 90% of the memory in the original implementation.

### Time Formatting Optimization: Eliminating the Remaining Bottleneck

After implementing `sync.Pool`, we achieved an 87% reduction in memory usage. However, memory profiling revealed that `time.Time.Format()` was now consuming **97.37% of all remaining allocations**. This section explains why this happened, how we discovered it, and how we eliminated it completely using `AppendFormat` with `AvailableBuffer()`.

### The Problem: Why `time.Time.Format()` Allocates So Much

When we profiled the code after adding `sync.Pool`, we discovered that `time.Time.Format()` became the dominant memory consumer:

```
(pprof) top
   55.50MB 97.37% 97.37%   55.50MB 97.37%  time.Time.Format
```

**Why does `time.Time.Format()` allocate so much memory?**

1. **String immutability in Go**: Strings in Go are immutable. Every time you format a time, Go must create a new string object in memory. There's no way to reuse or modify an existing string.

2. **Internal implementation**: The `Format()` method internally:
   - Parses the format string (e.g., `"15:04:05"`)
   - Allocates a temporary buffer to build the formatted string
   - Converts the buffer to a string (which allocates new memory)
   - Returns the new string

3. **High-frequency operations**: In logging scenarios, time formatting happens on every log call. If you're logging millions of times, you're allocating millions of strings.

4. **Hidden allocations**: Even though `Format()` returns a single string, the internal process involves multiple allocations:
   - Buffer allocation for building the formatted output
   - String conversion allocation
   - Potential intermediate allocations during parsing

### How We Discovered the Problem

We used Go's built-in memory profiler to identify the bottleneck:

```bash
go test -bench=BenchmarkLogWithPool -memprofile=mem.out
go tool pprof -alloc_space ./go-sync-pool.test mem.out
```

The profile clearly showed that after eliminating buffer allocations with `sync.Pool`, `time.Time.Format` was now the dominant memory consumer at 97.37% of all allocations.

### The Solution: Using `AppendFormat` with `AvailableBuffer()`

Instead of using `time.Now().Format("15:04:05")`, we use `AppendFormat` with the buffer's available capacity:

```go
    b.Write(time.Now().AppendFormat(b.AvailableBuffer(), "15:04:05"))
```

### Why This Solution Works

1. **`AvailableBuffer()` returns unused capacity**: Introduced in Go 1.21, `AvailableBuffer()` returns an empty slice backed by the buffer's unused capacity. This slice points to memory that's already allocated but not yet used.

2. **`AppendFormat` reuses the underlying array**: When you pass a slice with capacity to `AppendFormat`, it appends to that slice. If the capacity is sufficient (which it is, since the buffer is pooled and reused), `AppendFormat` will reuse the underlying array without allocating.

3. **Pooled buffers have capacity**: Since we're using `sync.Pool`, buffers are reused across operations. After the first use, buffers retain their capacity, so `AvailableBuffer()` returns a slice with plenty of space for the formatted time string (which is only 8 bytes for "15:04:05").

4. **Zero allocations**: The entire operation has zero memory allocations:
   - `AvailableBuffer()` returns a slice view of existing memory (no allocation)
   - `AppendFormat` appends to that slice, reusing the underlying array (no allocation)
   - `Write()` writes the result back to the buffer (no allocation)

### The Results: Final Optimized Performance

After implementing `AppendFormat` with `AvailableBuffer()`, we ran the benchmarks again:

```bash
go test -bench=BenchmarkLogWithPool -benchmem
```

**Performance Results:**
```
goos: linux
goarch: amd64
pkg: github.com/ArditZubaku/go-sync-pool
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkLogWithPool-14         15856262                70.81 ns/op
PASS
```

**Performance Improvements:**
- **61% faster** than original implementation (183.2 ns/op → 70.81 ns/op)
- **100% elimination of allocations** (2 allocs/op → 0 allocs/op)
- **Zero memory overhead** from logging operations

**Memory Profile Results:**

```bash
go test -bench=BenchmarkLogWithPool -memprofile=mem.out
go tool pprof -alloc_space ./go-sync-pool.test mem.out
```

```
(pprof) top
Showing nodes accounting for 2564.03kB, 100% of 2564.03kB total
Showing top 10 nodes out of 14
      flat  flat%   sum%        cum   cum%
    2052kB 80.03% 80.03%     2052kB 80.03%  runtime.allocm
  512.02kB 19.97%   100%   512.02kB 19.97%  testing.(*B).ResetTimer
         0     0%   100%     1026kB 40.02%  runtime.mcall
         0     0%   100%     1026kB 40.02%  runtime.mstart
         0     0%   100%     1026kB 40.02%  runtime.mstart0
         0     0%   100%     1026kB 40.02%  runtime.mstart1
         0     0%   100%     2052kB 80.03%  runtime.newm
         0     0%   100%     1026kB 40.02%  runtime.park_m
         0     0%   100%     2052kB 80.03%  runtime.resetspinning
         0     0%   100%     2052kB 80.03%  runtime.schedule
```

**Memory Improvements:**
- **99.4% reduction** in total memory allocations (429.53MB → 2.56MB)
- **`time.Time.Format` completely eliminated** - no longer appears in the profile!
- **Zero allocations from our logging code** - all remaining allocations are from Go runtime internals
- **Buffer growth allocations eliminated** - sync.Pool handles all buffer reuse

**What Remains:**
- `runtime.allocm` (80.03%) - Go runtime thread management overhead (unavoidable system cost)
- `testing.(*B).ResetTimer` (19.97%) - Benchmark framework overhead (testing-only, not production)

### Complete Optimization Journey

The complete optimization journey demonstrates the power of profiling and targeted fixes:

1. **Original Implementation**:
   - 88.94% buffer growth allocations
   - 10.71% time formatting allocations
   - Total: 429.53MB allocated
   - Performance: 183.2 ns/op, 2 allocs/op

2. **After sync.Pool**:
   - 0% buffer growth (eliminated!)
   - 97.37% time formatting (now the bottleneck)
   - Total: 57MB allocated (87% reduction)
   - Performance: 155.2 ns/op, 1 alloc/op

3. **After AppendFormat with AvailableBuffer()**:
   - 0% buffer growth (still eliminated)
   - 0% time formatting (eliminated!)
   - Total: 2.56MB allocated (99.4% reduction from original)
   - Performance: 70.81 ns/op, 0 allocs/op
   - Only runtime overhead remains

### Final Key Takeaways

1. **Profiling reveals hidden costs**: Without memory profiling, we wouldn't have known that `time.Time.Format` was the bottleneck after fixing buffers.

2. **Every allocation matters**: Even seemingly small operations like time formatting can dominate memory usage in high-frequency code paths.

3. **Modern Go features enable zero-allocation**: `AvailableBuffer()` (Go 1.21+) combined with `AppendFormat` provides a clean, standard-library way to eliminate allocations without manual formatting code.

4. **Zero-allocation is achievable**: With careful optimization using modern Go features, even high-frequency operations can achieve zero allocation overhead.

5. **Optimizations compound**: Each optimization (sync.Pool, then AppendFormat) built on the previous one, resulting in a 99.4% total reduction and zero allocations.

The final result demonstrates that with profiling-driven optimization and modern Go features, we can achieve zero allocations from application code, leaving only unavoidable runtime overhead.
