# Benchmarking

The benchmarks measure the DSP and fingerprinting pipeline on a synthetic, song-length signal: 240 seconds at 44,100 Hz, built from five summed sine waves plus seeded noise. The signal is deterministic, so runs are comparable, and the benchmarks need no files in `testdata/` and no MP3 decoding.

## Available benchmarks

| Benchmark | Package | What it measures |
|---|---|---|
| `BenchmarkSplitFrames` | `internal/dsp` | Cutting the signal into overlapping frames. |
| `BenchmarkForward` | `internal/dsp` | One FFT over a single windowed 4096-sample frame. It isolates the transform from the rest of the pipeline. |
| `BenchmarkSpectrogram` | `internal/dsp` | Windowing, FFT, and magnitudes for every frame of the signal (about 5,166 frames). |
| `BenchmarkFromBuffer` | `internal/fingerprint` | The full pipeline: gain normalization, frames, spectrogram, peaks, and landmarks. |

All benchmarks use the production defaults from `dsp.DefaultConfig()` and `internal/fingerprint/defaults.go`.

## Running benchmarks

```sh
# All benchmarks in both packages
go test -run '^$' -bench . ./internal/dsp/ ./internal/fingerprint/

# One benchmark
go test -run '^$' -bench '^BenchmarkSpectrogram$' ./internal/dsp/

# Several benchmarks
go test -run '^$' -bench 'Spectrogram|Forward' ./internal/dsp/
```

Useful flags:

| Flag | Effect |
|---|---|
| `-run '^$'` | Matches no test names, so only benchmarks run. Without it, `go test` runs every unit test first. |
| `-bench REGEX` | Runs the benchmarks whose names match `REGEX`. `.` matches all of them. |
| `-benchmem` | Reports `B/op` and `allocs/op`. The bassclef benchmarks call `b.ReportAllocs()`, so they report these columns without the flag. |
| `-count N` | Repeats each benchmark `N` times. Use `-count=10` when you compare two versions, so you can see the run-to-run variation. |
| `-benchtime D` | How long each benchmark runs, as a duration (`5s`) or a fixed number of iterations (`20x`). The default is `1s`. |
| `-cpu LIST` | Sets `GOMAXPROCS` for each run, for example `-cpu 1,4,10`. It matters only for code that uses goroutines. |
| `-timeout D` | The default is `10m`. Raise it for long runs, for example `-count=10` over both packages. |

## Reading the output

```
BenchmarkSpectrogram-10    	       5	 196988364 ns/op	84893696 B/op	       4 allocs/op
```

| Column | Meaning |
|---|---|
| `BenchmarkSpectrogram-10` | Benchmark name. The suffix is `GOMAXPROCS`. |
| `5` | Number of iterations Go ran to get a stable measurement. |
| `196988364 ns/op` | Mean time per iteration (here about 197 ms). |
| `84893696 B/op` | Bytes allocated on the heap per iteration (here about 84.9 MB). |
| `4 allocs/op` | Number of heap allocations per iteration. |

Results depend on the machine. Compare numbers only between runs on the same machine, and note the `goos`, `goarch`, and `cpu` lines that `go test` prints. On a laptop, plug in the power and close other heavy programs before you measure.

## Comparing two versions

Save one run from before the change and one from after, then compare them with [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```sh
go install golang.org/x/perf/cmd/benchstat@latest

# First, measure the current code
go test -run '^$' -bench . -count=10 ./internal/dsp/ ./internal/fingerprint/ > old.txt

# Make your change, then measure again
go test -run '^$' -bench . -count=10 ./internal/dsp/ ./internal/fingerprint/ > new.txt

# Compare with benchstat
benchstat old.txt new.txt
```

`benchstat` prints the mean and variation of each benchmark, the percentage change, and a p-value. A change marked `~` is not statistically significant. If your change is already in the working tree, run `git stash` to measure the baseline, then `git stash pop` to bring the change back.

`*.out` and `*.txt` result files are local only. `*.out` is gitignored; do not commit `old.txt` or `new.txt`.

## Profiling

A benchmark shows how long code takes. A profile shows which functions and lines that time or memory goes to.

**Memory profile.** This records the allocations made during the benchmark:

```sh
go test -run '^$' -bench '^BenchmarkFromBuffer$' -memprofile mem.out ./internal/fingerprint/
```

By default, Go samples about one allocation per 512 KiB. Add `-memprofilerate=1` to record every allocation. This is exact but slower.

**CPU profile.** This records where the CPU time goes:

```sh
go test -run '^$' -bench '^BenchmarkFromBuffer$' -cpuprofile cpu.out ./internal/fingerprint/
```

With either flag, `go test` also keeps the compiled test binary in the current directory (for example `fingerprint.test`). Both `*.out` and `*.test` are gitignored.

**Reading a profile.** Use `go tool pprof`:

```sh
# Functions ranked by total bytes allocated
go tool pprof -top -sample_index=alloc_space mem.out

# Functions ranked by number of allocations
go tool pprof -top -sample_index=alloc_objects mem.out

# Allocations line by line inside matching functions 
go tool pprof -list 'FindPeaks|GenerateLandmarks' -sample_index=alloc_space mem.out

# Functions ranked by CPU time
go tool pprof -top cpu.out

# Interactive web UI with flame graph, call graph, source views, etc.
go tool pprof -http=:8080 cpu.out
```

For memory profiles, use `-sample_index=alloc_space` or `alloc_objects`. The default, `inuse_space`, shows only the memory still in use when the benchmark ends, so it hides short-lived buffers. The call-graph view in the web UI needs [Graphviz](https://graphviz.org/) (`brew install graphviz` on macOS).

## Benchmark reports

Write-ups of past optimizations are in [`docs/benchmarks/`](docs/benchmarks/), named `YYYYMMDD-<package>.md`. Each report lists the change, the setup, the before and after results with raw output, and the next target. Add a report when a change makes a measurable difference.
