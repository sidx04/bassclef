# Contributing to bassclef

Thanks for your interest in bassclef. This project is a learning-focused audio fingerprinting system, so readable code and clear explanations matter as much as speed.

## Getting set up

You need:

- Go 1.26.5 or later (see `go.mod`).
- A C toolchain, only if you work on microphone capture (`listen` uses cgo through `malgo`). On macOS, Xcode Command Line Tools is enough.

Clone the repository and build:

```sh
go build ./...
make help        
```

For manual testing, put your own audio files in `testdata/`. That directory is gitignored. Do not commit audio files.

## Before you open a pull request

Run these checks:

```sh
make fmt         
make check       
```

The tests must pass without a microphone, Spotify credentials, or network access. Use fakes for `storage.Store` and `metadata.MetadataProvider`, and use `httptest` servers for HTTP code.

If your change touches the DSP or fingerprinting pipeline, run the benchmarks before and after and include the comparison in the pull request:

```sh
make bench-save OUT=old.txt     # on code pre-dating any sort of your improvements
make bench-save OUT=new.txt     # on code after improvements
make bench-compare OLD=old.txt NEW=new.txt
```

See `benchmark.md` for details on the benchmarks.

## Rules for the pipeline

- **Keep ingest and query in sync.** Ingest and recognition both call `fingerprint.FromBuffer`. Change pipeline parameters only in `dsp.DefaultConfig()` and `internal/fingerprint/defaults.go`.
- **Say when a change invalidates catalogs.** A change to the window function, FFT size, hop size, peak or landmark defaults, or hash layout changes every fingerprint. Existing `.db` files must be deleted and re-ingested. State this clearly in the pull request.
- **Keep stages separate.** Capture, DSP, fingerprinting, matching, storage, and metadata each live in their own `internal/` package and talk through the `AudioDecoder`, `FFT`, `Store`, and `MetadataProvider` interfaces.

## Where code goes

- **New CLI subcommand:** define a `<X>Cmd` struct with a `Run() error` method and add it as a field of the `CLI` struct in `cmd/bassclef/main.go`.
- **HTTP code:** `cmd/bassclef/api.go`.
- **New audio format:** implement `audio.AudioDecoder` and register it by extension in `decodersByExt` in `internal/audio/load.go`.

## Branches, commits, and pull requests

- Branch from `main`. Name the branch after the change, for example `fix/6-song-id-response` or `feat/flac-decoder`.
- Use [Conventional Commits](https://www.conventionalcommits.org/) prefixes: `feat:`, `fix:`, `perf:`, `bench:`, `docs:`, `chore:`. An optional scope is fine, for example `feat(dsp): ...`.
- Keep each pull request focused on one change. Link the issue it closes, if there is one.
- Update `README.md` when you change CLI flags, the HTTP API, or the project layout.

## Reporting bugs and ideas

Open a GitHub issue. For a bug, include the command you ran, what you expected, what happened, and your OS and Go version. If the bug depends on a specific audio file, describe the file - format, sample rate, length - instead of attaching copyrighted audio.
