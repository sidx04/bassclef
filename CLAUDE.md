# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`bassclef` is a learning-focused Go implementation of a Shazam-style audio fingerprinting system, built from first principles rather than using an ML black box. The full recognition pipeline is implemented:

```
Audio → Gain normalize → Frames → Window → FFT → Spectrogram → Peaks → Landmark Pairs → Hashes → SQLite → Time-Offset Voting → Song Match → (optional) Spotify metadata
```

`README.md` is the user/contributor-facing doc (usage, CLI, HTTP API, layout). `NOTES.md` holds background notes and references. `bassclef-plan.md` (gitignored, local only) is the original design doc; it's useful for algorithm rationale but its roadmap predates the current code.

## Commands

```
go build ./...                                 # build everything
go build -o bassclef ./cmd/bassclef            # build the binary
go run ./cmd/bassclef <command> --help         # info|spectrogram|peaks|fingerprint|ingest|recognize|listen|serve
go test ./...                                  # run all tests
go test ./internal/dsp/...                     # test a single package
go test ./internal/dsp/ -run TestName          # run a single test
go test -bench=. ./internal/dsp/ ./internal/fingerprint/
go vet ./...
```

Audio fixtures for manual testing live in `testdata/` (gitignored). The catalog DB defaults to `bassclef.db` in the working directory (`*.db` is gitignored).

## Architecture

- **`cmd/bassclef/`** — the binary (`package main`), using `alecthomas/kong`. `main.go` declares the `CLI` struct; each subcommand is a field. `cli.go` holds all subcommands except `serve`, plus the shared recognition path: `recognizeCandidates` (fingerprint → match → Spotify enrich, returns `[]MatchResult`) and `recognizeBuffer` (prints it). File, mic, and HTTP recognition all go through `recognizeCandidates` so they can't drift. `api.go` holds `ServeCmd`, the `POST /recognize` handler (`newRecognizeHandler`, store/provider injected for tests), and the `go:embed`-ed `webui/` static UI. Add a subcommand by defining a `<X>Cmd` struct with `Run() error` and adding it to `CLI`.
- **`internal/audio/`** — decodes to the canonical `AudioBuffer{Samples []float64, SampleRate int}`: mono, normalized to `-1.0..1.0`. `AudioDecoder` interface (`Decode(io.Reader)`), implemented by `WAVDecoder` and `MP3Decoder`. `Load(path)` picks a decoder by file extension (`decodersByExt` in `load.go`). `Capture(duration)` records from the default mic via `malgo` (cgo) at 44,100 Hz. `utils.go` has stereo→mono and bit-depth normalization helpers.
- **`internal/dsp/`** — `NormalizeGain` (peak-scale to full scale, in place), `SplitFrames`, `WindowFunc` / `HannWindow`, `FFT` interface (`Forward([]float64) []complex128`) with `CmplxFFT` (gonum) as the implementation, `Spectrogram(frames, window, fft)` → magnitudes. `Config` / `DefaultConfig()`: FFT size 4096, hop 2048, Hann window.
- **`internal/fingerprint/`** — `FindPeaks` (local-max search with `PeakConfig{FreqWindow, TimeWindow, MinMagnitude}`), `GenerateLandmarks` (`LandmarkConfig{MinDeltaFrames, MaxDeltaFrames, Fanout}`), `Hash(f1, f2, dt) uint64` packing `[f1:12][f2:12][dt:14]`, `BinToHz`. `FromBuffer(buf)` runs the whole pipeline with `DefaultPeakConfig()` (3/3/1.0) and `DefaultLandmarkConfig()` (1/50/10) from `defaults.go`.
- **`internal/catalog/`** — `Ingest(store, path, title, artist)`: load → `FromBuffer` → insert song + fingerprints.
- **`internal/storage/`** — `Store` interface (`InsertSong`, `InsertFingerprints`, `LookupHash`, `GetSong`, `CountFingerprints`, `Close`); `SQLiteStore` via pure-Go `modernc.org/sqlite`. Schema: `songs(id, title, artist, duration, source UNIQUE)`, `fingerprints(hash, song_id, time_offset)` indexed on `hash`.
- **`internal/matcher/`** — `Match(store, []QueryLandmark, MatchConfig{TopN, MinVotes})`: `offset = db_time - query_time`, `votes[song][offset]++`; `Score` = best offset's votes, `Confidence` = Score / song's fingerprint count. Ranks by `Confidence`; `MinVotes` filters on raw `Score`.
- **`internal/metadata/`** — `MetadataProvider` interface (`Lookup(title, artist string) (Metadata, error)`), `SpotifyProvider` (Client Credentials flow, token cached). Reads `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`, `SPOTIFY_ACCOUNTS_URL`, `SPOTIFY_API_URL` from env or `.env` (URLs have no defaults). Missing credentials are non-fatal: `newMetadataProvider` warns and returns nil, and recognition skips enrichment.

## Invariants

- Ingest and query must fingerprint identically or hashes never align. Both go through `fingerprint.FromBuffer`; change parameters only in `dsp.DefaultConfig()` and `fingerprint/defaults.go`. Any change to window, FFT/hop size, peak/landmark defaults, or hash layout invalidates existing catalogs (re-ingest required).
- The `peaks` / `fingerprint` CLI commands build their own pipeline from flags (default `--min-magnitude 0`, no gain normalization). They are for experimentation and do not match what `ingest` stores.

## Design principle

Keep audio capture, DSP, fingerprinting, matching, storage, and metadata in separate packages under `internal/`, talking through the `AudioDecoder`, `FFT`, `Store`, and `MetadataProvider` interfaces. The fingerprinting pipeline should stay testable against audio fixtures without microphone hardware or network access (tests use fake stores/providers and `httptest` servers).
