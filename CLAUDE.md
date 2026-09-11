# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`bassclef` is a learning-focused Go implementation of a Shazam-style audio fingerprinting system, built from first principles rather than using an ML black box. Core recognition pipeline (target, not fully built):

```
Audio → Spectrogram → Spectral Peaks → Landmark Pairs → Hashes → Database → Time-Offset Voting → Song Match
```

`bassclef-plan.md` (gitignored, local only) is the full design doc this file is distilled from — it's the source of truth for algorithm details and the roadmap below. Re-read it directly if a task needs more than the summary here.

## Commands

```
go build ./...                          # build everything
go run ./cmd/bassclef info <file.wav>   # only implemented CLI subcommand today
go test ./...                           # run all tests
go test ./internal/dsp/...              # test a single package
go test ./internal/dsp/ -run TestName   # run a single test
go vet ./...
```

Sample WAV fixtures for manual testing/tests live in `testdata/` (gitignored).

## Architecture: what's implemented

- **`cmd/bassclef/`** — the CLI binary (`package main`). Uses `alecthomas/kong` for command parsing; `main.go` declares a single `CLI` struct where each subcommand is a field (commented-out fields show planned-but-unbuilt commands: spectrogram, peaks, fingerprint, ingest, recognize — see roadmap below). Each subcommand's flags/args and `Run()` method live in its own file (e.g. `cli.go` has `InfoCmd`). Add a new subcommand by defining a `<X>Cmd` struct with a `Run() error` method and wiring it into the `CLI` struct in `main.go`.
- **`internal/audio/`** — decodes source audio into the canonical `AudioBuffer{Samples []float64, SampleRate int}`, normalized to `-1.0..1.0` mono float64 regardless of input format. `AudioDecoder` is an interface (`Decode(io.Reader) (*AudioBuffer, error)`) so additional decoders (MP3, mic capture) can plug in alongside `WAVDecoder` without changing downstream code. `LoadWAV` is the file-path convenience entry point; `utils.go` holds the stereo→mono averaging and bit-depth normalization helpers the decoder relies on.
- **`internal/dsp/`** — turns a sample buffer into FFT-ready frames. `SplitFrames` chops the signal into overlapping windows (frame/hop size come from `Config`; `DefaultConfig()` is FFT size 4096 / hop size 2048, matching the plan's target sample rate of 44,100 Hz). `HannWindow` + `ApplyWindow` taper frame edges to prevent spectral leakage before an FFT is run (see README for the periodicity/leakage rationale). The FFT and spectrogram step itself is **not implemented yet** — this package currently only prepares frames for it.
- Root-level `main.go` (`package main`, distinct from `cmd/bassclef/main.go`) is currently an empty stub — the real entry point is `cmd/bassclef/main.go`.

## Architecture: planned but not present

Per `bassclef-plan.md`, these packages don't exist yet — check there before assuming a design:

- **`internal/dsp`** (extension) — an `FFT` interface (`Forward(input []float64) []complex128`) is meant to decouple the DSP pipeline from any one FFT library; magnitudes across frames form the spectrogram.
- **`internal/fingerprint/`** — peak detection, landmark pairing, hashing:
  - `Peak{Time, Frequency int; Magnitude float64}` — strongest local maxima per frame (start simple: top 10–30 peaks/frame; later add thresholds, log-magnitude scaling, frequency bands, neighborhood suppression).
  - `Landmark{AnchorFrequency, TargetFrequency, DeltaTime, AnchorTime int}` built from `LandmarkConfig{MinDeltaFrames, MaxDeltaFrames, Fanout}` — pair each anchor peak with nearby target peaks (delta ~1 to 20–50 frames, fanout ~5–15 targets/anchor).
  - `Hash(f1, f2, dt uint32) uint64` — deterministic, compact (not cryptographic) hash packing `[f1][f2][dt]`.
- **`internal/matcher/`** — time-offset voting: `offset = database_time - query_time`, `votes[song][offset]++`; best match is the song with the strongest vote cluster at one offset. This is what lets a short clip match anywhere within a full song.
- **`internal/catalog/`** and **`internal/storage/`** — SQLite-backed fingerprint DB (`songs` table; `fingerprints(hash, song_id, time_offset)` indexed on `hash`). Not meant to scale to a large catalog initially — Postgres/KV storage is a later swap if needed.
- **`MetadataProvider`** interface (`Lookup(song Song) (Metadata, error)`) — keeps recognition independent of external services like Spotify; local metadata first.

### Roadmap phases (plan doc §5, in order)

1. Offline WAV processing (done: decode → mono → normalize)
2. FFT + spectrogram (not started — next up)
3. Peak detection
4. Landmark generation + hashing
5. SQLite database + offset-voting matcher
6. Microphone capture (`bassclef listen --duration 8s`)
7. External metadata providers

Milestone CLI shape to match as commands land: `bassclef spectrogram|peaks|fingerprint|ingest|recognize|listen <file>`.

## Design principle

Keep audio capture, DSP, fingerprinting, matching, and storage in separate packages under `internal/`. The fingerprinting pipeline should stay testable against WAV fixtures without requiring microphone hardware.
