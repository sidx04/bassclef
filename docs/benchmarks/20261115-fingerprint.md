# GenerateLandmarks: preallocate and skip the sort

| | |
| --- | --- |
| Date | 15th November, 2026 |
| Files | `internal/fingerprint/landmark.go` |
| Benchmarks | `BenchmarkGenerateLandmarks`, `BenchmarkFromBuffer` (`internal/fingerprint/bench_test.go`) |

## Summary

| | Before | After | Change |
| --- | --- | --- | --- |
| `GenerateLandmarks`, time/op | 31.79 ms | 9.57 ms | 3.3x faster |
| `GenerateLandmarks`, memory/op | 397.8 MB | 65.4 MB | 6.1x less |
| `GenerateLandmarks`, allocs/op | 46 | 1 | |
| `FromBuffer`, time/op | 470.2 ms | 450.9 ms | 4% faster |
| `FromBuffer`, memory/op | 511.1 MB | 178.8 MB | 2.9x less |
| `FromBuffer`, allocs/op | 87 | 42 | |
| Landmarks | 2,043,230 | 2,043,230 | identical |

All times are the mean of 10 runs. The output is identical landmark for landmark, so **no re-ingest is needed**.

Like the `Spectrogram` buffer reuse in [20261115-dsp.md](20261115-dsp.md), this is mainly a memory win. `GenerateLandmarks` is about 6% of `FromBuffer`'s CPU time, so saving 22 ms per call moves the full pipeline by only 4%. The 332 MB it no longer allocates is about two thirds of what `FromBuffer` used to allocate.

## Finding the problem

A memory profile of `BenchmarkFromBuffer` showed `GenerateLandmarks` as the largest allocator in the pipeline, ahead of `Spectrogram`. The line-level view (`go tool pprof -list GenerateLandmarks -sample_index=alloc_space`, 3 iterations) pointed at two lines:

```
1.09 GB   landmark.go:60   landmarks = append(landmarks, Landmark{...})
14.04 MB  landmark.go:37   sorted := make([]Peak, len(peaks))
```

## Change

### 1. Preallocate the landmarks slice

Before, `landmarks` started empty (`var landmarks []Landmark`) and grew through `append`. When a slice runs out of capacity, `append` allocates a bigger array and copies every element into it. For large slices, Go grows capacity by only about 1.25x each time, so building 2,043,230 landmarks took about 45 reallocations. Summed over all of them, the copies come to about 5x the final 65 MB, which is the ~330 MB the profile showed.

Each anchor produces at most `cfg.Fanout` landmarks, so `len(peaks) * cfg.Fanout` is a hard upper bound on the result. The slice is now allocated once with that capacity:

```go
landmarks := make([]Landmark, 0, len(peaks)*cfg.Fanout)
```

With the default config, the capacity is 2,043,880 for 2,043,230 landmarks: 99.97% of the slots are used. With sparse peaks or a narrow delta window, more capacity goes unused, but it never exceeds `len(peaks) × Fanout × 32 B`. Counting the landmarks exactly in a first pass would avoid that, at the cost of running the pairing loop twice.

### 2. Skip the copy and sort when peaks are already sorted

Before, every call copied `peaks` and sorted the copy with `sort.SliceStable`. `FindPeaks` loops over time first (`internal/fingerprint/peaks.go`), so its output is already in time order, and the copy and sort did nothing.

The function now checks first, and copies and sorts only when it has to:

```go
sortByTimeFunc := func(a, b Peak) int {
	return cmp.Compare(a.Time, b.Time)
}

sortedPeaks := peaks
if !slices.IsSortedFunc(peaks, sortByTimeFunc) {
	sortedPeaks = slices.Clone(peaks)
	slices.SortStableFunc(sortedPeaks, sortByTimeFunc)
}
```

- `GenerateLandmarks` still accepts unsorted peaks, as its doc comment promises.
- The caller's slice is never modified. When the input is already sorted, the function only reads it. When it isn't, the function sorts its own clone.
- `slices.SortStableFunc` replaces `sort.SliceStable`, which works through reflection; the generic version avoids that overhead. The sort is still stable, so peaks in the same frame keep their frequency order and the landmark order does not change.
- The `IsSortedFunc` check is one O(n) pass. It replaces a 4.7 MB copy (204,388 peaks × 24 B) and an O(n log n) sort on every call.

### The one remaining allocation

The only allocation left is the `landmarks` slice itself: 2,043,880 × 32 B = 65,404,160 B, which the allocator rounds up to the 65,404,928 B in the benchmark. It is the function's result, so it has to exist.

## A pitfall along the way: length versus capacity

The first version of change 1 was:

```go
var landmarks = make([]Landmark, len(peaks)*cfg.Fanout)
```

With two arguments, `make` sets the slice's **length**, not only its capacity. The slice started with 2,043,880 zero-valued `Landmark`s, and `append` added the real landmarks after them. Two effects:

- **Wrong output.** Every result started with about 2 million `{0, 0, 0, 0}` landmarks. Ingesting with this would have stored 2 million identical hashes at time 0 for every song and broken matching. Four tests in `landmark_test.go` caught it (`SinglePair`, `FanoutCap`, `DeltaWindowBoundaries`, `UnsortedInput`).
- **More memory, not less.** The slice was full from the start, so the first `append` grew it anyway. `BenchmarkGenerateLandmarks` measured 541.8 MB/op, worse than the 397.8 MB/op before the change.

The fix is the three-argument form, `make([]Landmark, 0, n)`: length 0, capacity `n`.

## Tests

- New: `TestGenerateLandmarksDoesNotMutateInput` (`internal/fingerprint/landmark_test.go`). It passes sorted and unsorted peaks and checks that the caller's slice is unchanged afterwards. To confirm the test catches the bug it is meant for, it was also run against a copy of `landmark.go` with the `slices.Clone` removed: the `unsorted` case failed with `input modified`.
- The existing landmark tests pass unchanged.
- The output was checked against the `HEAD` version of `GenerateLandmarks` on the full benchmark signal, using a temporary test run through `go test -overlay`:
  - Sorted input (the normal path, no copy): 2,043,230 landmarks, identical to `HEAD`.
  - Reversed input (forces the clone and sort): identical to `HEAD`.
  - The input slice was unchanged in both cases.

## Setup

- Input: synthetic 240 seconds signal at 44,100 Hz (`createBenchBuffer`): 10,584,000 samples
- Config: `DefaultConfig()` (FFT size 4096, hop 2048, 5,166 frames), `DefaultPeakConfig()`, `DefaultLandmarkConfig()` (min delta 1, max delta 50, fanout 10)
- Peaks fed to `GenerateLandmarks`: 204,388, produced by the same steps as `FromBuffer`
- Go 1.27.1, darwin/arm64, Apple M5
- 10 runs per version (`-count=10`)

`BenchmarkGenerateLandmarks` is new in this change. It runs gain normalization, framing, the spectrogram, and `FindPeaks` once, outside the timed loop, so only `GenerateLandmarks` is timed. The "before" numbers come from the `HEAD` version of `landmark.go`, swapped in with `go test -overlay` so the working tree did not change.

Reproduce:

```sh
go test -run '^$' -bench . -count=10 ./internal/fingerprint/ > new.txt
```

## Results

| Benchmark | Version | Mean | Min | Max | Std dev |
| --- | --- | --- | --- | --- | --- |
| `GenerateLandmarks` | Before | 31,794,695 ns | 30,525,533 ns | 32,718,870 ns | 619,022 ns |
| `GenerateLandmarks` | After | 9,574,922 ns | 8,747,314 ns | 12,522,598 ns | 1,018,419 ns |
| `FromBuffer` | Before | 470,187,228 ns | 459,569,444 ns | 508,584,792 ns | 14,154,588 ns |
| `FromBuffer` | After | 450,886,243 ns | 441,442,250 ns | 480,549,930 ns | 11,536,924 ns |

## Next target

`FromBuffer` now allocates 178.8 MB/op. `Spectrogram` (84.9 MB) and the `landmarks` result (65.4 MB) are both outputs that have to exist. That leaves about 28 MB elsewhere, mostly in `FindPeaks`, which builds its `peaks` slice with the same growing `append`. Unlike landmarks, the peak count has no tight upper bound, so preallocating there needs a different approach.

For time, `isLocalMax` (called by `FindPeaks`) is about a third of `FromBuffer`'s CPU time, the largest single cost left. Profile it before changing anything:

```sh
go test -run '^$' -bench '^BenchmarkFromBuffer$' -cpuprofile cpu.out ./internal/fingerprint/
go tool pprof -list 'isLocalMax' cpu.out
```

## Raw output

Before:

```
goos: darwin
goarch: arm64
pkg: github.com/sidx04/bassclef/internal/fingerprint
cpu: Apple M5
BenchmarkFromBuffer-10           	       2	 508584792 ns/op	511108692 B/op	      92 allocs/op
BenchmarkFromBuffer-10           	       3	 459569444 ns/op	511105890 B/op	      88 allocs/op
BenchmarkFromBuffer-10           	       3	 473162306 ns/op	511105869 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 465233945 ns/op	511105906 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 461822472 ns/op	511105832 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 460273958 ns/op	511105869 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 462412333 ns/op	511105869 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 480199014 ns/op	511105832 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 463508778 ns/op	511105869 B/op	      87 allocs/op
BenchmarkFromBuffer-10           	       3	 467105236 ns/op	511105832 B/op	      87 allocs/op
BenchmarkGenerateLandmarks-10    	      38	  31321044 ns/op	397757202 B/op	      47 allocs/op
BenchmarkGenerateLandmarks-10    	      33	  31751633 ns/op	397757389 B/op	      47 allocs/op
BenchmarkGenerateLandmarks-10    	      39	  31166705 ns/op	397757168 B/op	      46 allocs/op
BenchmarkGenerateLandmarks-10    	      31	  32296968 ns/op	397757181 B/op	      46 allocs/op
BenchmarkGenerateLandmarks-10    	      38	  30525533 ns/op	397757200 B/op	      47 allocs/op
BenchmarkGenerateLandmarks-10    	      36	  32207581 ns/op	397757175 B/op	      46 allocs/op
BenchmarkGenerateLandmarks-10    	      45	  32104622 ns/op	397757200 B/op	      47 allocs/op
BenchmarkGenerateLandmarks-10    	      32	  32718870 ns/op	397757182 B/op	      46 allocs/op
BenchmarkGenerateLandmarks-10    	      37	  32237334 ns/op	397757148 B/op	      46 allocs/op
BenchmarkGenerateLandmarks-10    	      34	  31616656 ns/op	397757190 B/op	      46 allocs/op
PASS
ok  	github.com/sidx04/bassclef/internal/fingerprint	37.391s
```

After:

```
goos: darwin
goarch: arm64
pkg: github.com/sidx04/bassclef/internal/fingerprint
cpu: Apple M5
BenchmarkFromBuffer-10           	       3	 461224764 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 450491764 ns/op	178757437 B/op	      47 allocs/op
BenchmarkFromBuffer-10           	       3	 480549930 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 447252750 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 446053778 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 441652000 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 453956125 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 444333653 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 441905417 ns/op	178753672 B/op	      42 allocs/op
BenchmarkFromBuffer-10           	       3	 441442250 ns/op	178753672 B/op	      42 allocs/op
BenchmarkGenerateLandmarks-10    	     135	   8747314 ns/op	65404929 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     127	   9370042 ns/op	65404929 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     127	   9268504 ns/op	65404929 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     100	  12522598 ns/op	65404930 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     120	   9843550 ns/op	65404933 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     130	   9018312 ns/op	65404928 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     130	   9084601 ns/op	65404928 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     128	   9271643 ns/op	65404930 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     128	   9244825 ns/op	65404928 B/op	       1 allocs/op
BenchmarkGenerateLandmarks-10    	     128	   9377829 ns/op	65404928 B/op	       1 allocs/op
PASS
ok  	github.com/sidx04/bassclef/internal/fingerprint	37.078s
```
