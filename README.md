# bassclef

A Shazam-style audio recognition system written in Go, built from first principles.

bassclef fingerprints songs into a local SQLite catalog. It then identifies short clips (from a file, your microphone, or a browser) by matching them against that catalog. It does not use machine learning. Each stage of the pipeline is plain signal processing that you can read, test, and change.

It is based on Avery Wang's 2003 paper, [*An Industrial-Strength Audio Search Algorithm*](https://www.researchgate.net/publication/220723446_An_Industrial_Strength_Audio_Search_Algorithm). Background reading and design notes are in [`NOTES.md`](NOTES.md).

## How it works

```
audio ─► decode ─► gain normalize ─► frames ─► Hann window ─► FFT ─► spectrogram
                                                                         │
song match ◄─ offset voting ◄─ hash lookup ◄─ hashes ◄─ landmark pairs ◄─ peaks
```

1. **Decode.** WAV, MP3, or microphone input becomes a mono `float64` buffer normalized to `-1.0..1.0`.
2. **Gain normalize.** The buffer is scaled so its loudest sample reaches full scale. This makes quiet and loud recordings of the same song comparable.
3. **Frame and window.** The signal is cut into overlapping frames (4096 samples, hop 2048). A Hann window tapers each frame's edges to reduce spectral leakage.
4. **Spectrogram.** An FFT of each frame gives frequency magnitudes over time.
5. **Peaks.** Local maxima in the spectrogram are kept if they are above a minimum magnitude. These are the "constellation" points.
6. **Landmarks.** Each anchor peak is paired with up to 10 later peaks within 1–50 frames.
7. **Hashes.** Each pair is packed into a 64-bit key: `[anchor freq: 12 bits][target freq: 12 bits][Δt: 14 bits]`.
8. **Matching.** For each query hash found in the catalog, the matcher computes `offset = catalog_time - query_time` and counts votes per song per offset. A real match produces many votes at one offset. Songs are ranked by their best offset's votes divided by their total fingerprint count, so songs with many fingerprints are not favored by chance collisions.

Step 8 is what lets an 8-second clip match anywhere inside a full-length song.

## Requirements

- Go **1.26.5** or later (see `go.mod`).
- A C toolchain, for microphone capture only. `listen` uses [`malgo`](https://github.com/gen2brain/malgo) (miniaudio bindings), which needs cgo. On macOS, Xcode Command Line Tools is enough.
- Optional: a [Spotify developer app](https://developer.spotify.com/dashboard) to add album, link, and cover art to results.

SQLite needs no system install. bassclef uses the pure-Go driver [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite).

## Quick start

```sh
git clone https://github.com/sidx04/bassclef.git
cd bassclef
go build -o bassclef ./cmd/bassclef

# 1. Add songs to the catalog (creates bassclef.db in the current directory)
./bassclef ingest songs/one.wav --title "Song One" --artist "Artist A"
./bassclef ingest songs/two.mp3 --title "Song Two" --artist "Artist B"

# 2. Identify a clip
./bassclef recognize clips/unknown.wav

# 3. Or listen through the microphone
./bassclef listen --duration 8s
```

Example output:

```
1. Song One - Artist A (score=212, confidence=0.0143)
   Album: Some Album
   Spotify: https://open.spotify.com/track/...
```

`score` is the raw vote count at the best offset. `confidence` is that score divided by the song's total fingerprint count. Results are ranked by `confidence`.

## CLI reference

Run `./bassclef --help` or `./bassclef <command> --help` for full flag help. All commands accept `.wav` and `.mp3` input.

| Command | Purpose |
|---|---|
| `info <file>` | Print sample rate, sample count, and duration. |
| `spectrogram <file>` | Compute the spectrogram and print its dimensions. |
| `peaks <file>` | Detect spectral peaks and print each one (time frame, Hz, magnitude). |
| `fingerprint <file>` | Generate landmarks and print each hash with its anchor time. |
| `ingest <file> --title T [--artist A]` | Fingerprint a song and store it in the catalog. |
| `recognize <file>` | Match a clip against the catalog. |
| `listen [--duration 8s]` | Record from the default microphone and match the recording. |
| `serve [--port 8080]` | Start the HTTP API and web UI. |

Shared flags:

- `--db PATH` (`ingest`, `recognize`, `listen`, `serve`): catalog file. Default `bassclef.db`.
- `--top-n N` (`recognize`, `listen`): number of ranked candidates to show. Default `5`.
- `--min-votes N` (`recognize`, `listen`): drop candidates with fewer votes. Default `1`.

`peaks` and `fingerprint` expose tuning flags (`--freq-window`, `--time-window`, `--min-magnitude`, `--min-delta-frames`, `--max-delta-frames`, `--fanout`). These flags are for experiments only. `ingest` and `recognize` always use the defaults in [`internal/fingerprint/defaults.go`](internal/fingerprint/defaults.go). Note that the `peaks` and `fingerprint` commands default to `--min-magnitude 0` and skip gain normalization, so their output differs from what `ingest` stores.

## HTTP API and web UI

```sh
./bassclef serve --port 8080 --db bassclef.db
```

Open `http://localhost:8080` for a web UI. It records from the browser microphone and sends the recording for recognition.

### `POST /recognize`

The request body is raw audio. Set `Content-Type` to `audio/wav` or `audio/mpeg`. The maximum body size is 50 MB.

```sh
curl -X POST http://localhost:8080/recognize \
  -H "Content-Type: audio/wav" \
  --data-binary @clips/unknown.wav
```

Response `200`:

```json
{
  "matches": [
    {
      "title": "Song One",
      "artist": "Artist A",
      "score": 212,
      "confidence": 0.0143,
      "album": "Some Album",
      "spotify_url": "https://open.spotify.com/track/...",
      "cover_url": "https://i.scdn.co/image/..."
    }
  ]
}
```

`matches` is an empty array when nothing matches. `album`, `spotify_url`, and `cover_url` are omitted when Spotify is not configured or has no result. Errors return `{"error": "..."}` with one of these statuses:

| Status | Cause |
|---|---|
| `400` | The audio could not be decoded. |
| `405` | The method is not `POST`. |
| `415` | The `Content-Type` is not supported. |
| `500` | Recognition failed. |

A Postman collection is in [`postman/`](postman/).

## Spotify metadata (optional)

Recognition works without Spotify. If credentials are missing, bassclef prints one warning and returns results without album, link, or cover art.

To enable it, create a `.env` file in the directory you run bassclef from, or export these variables:

```sh
SPOTIFY_CLIENT_ID=your-client-id
SPOTIFY_CLIENT_SECRET=your-client-secret
SPOTIFY_ACCOUNTS_URL=https://accounts.spotify.com
SPOTIFY_API_URL=https://api.spotify.com
```

All four variables are required. The two URL variables have no built-in defaults. bassclef uses the Client Credentials flow, so no user login is needed. Lookups search Spotify by the title and artist stored in the catalog, so accurate `--title` and `--artist` values at ingest give better metadata.

## Project layout

```
cmd/bassclef/          CLI entry point (kong). One <X>Cmd struct per subcommand.
  main.go              Subcommand registry.
  cli.go               Subcommands and the shared recognition path.
  api.go               `serve` command, POST /recognize handler, embedded web UI.
  webui/               Static web UI (HTML/JS/CSS), embedded with go:embed.
internal/
  audio/               Decoders (WAV, MP3), microphone capture, AudioBuffer type.
  dsp/                 Gain normalization, framing, window functions, FFT, spectrogram.
  fingerprint/         Peak detection, landmark pairing, hashing, FromBuffer pipeline.
  catalog/             Ingest: file to fingerprints to storage.
  storage/             Store interface and its SQLite implementation.
  matcher/             Time-offset voting and candidate ranking.
  metadata/            MetadataProvider interface and its Spotify implementation.
```

Each stage has its own package and talks to the others through small interfaces:

- `audio.AudioDecoder`: `Decode(io.Reader) (*AudioBuffer, error)`. Add a format by implementing it and registering it in `internal/audio/load.go`.
- `storage.Store`: the persistence boundary. `catalog` and `matcher` depend only on this interface, never on SQLite directly.
- `metadata.MetadataProvider`: `Lookup(title, artist string) (Metadata, error)`. Recognition never depends on a specific external service.

### Database schema

```sql
songs        (id INTEGER PRIMARY KEY, title TEXT NOT NULL, artist TEXT, duration REAL, source TEXT UNIQUE)
fingerprints (hash INTEGER NOT NULL, song_id INTEGER REFERENCES songs(id) ON DELETE CASCADE, time_offset INTEGER NOT NULL)
INDEX idx_fingerprint_hash ON fingerprints(hash)
```

`source` is unique, so ingesting the same file path twice fails.

## Development

```sh
go build ./...
go vet ./...
go test ./...                                  # all tests
go test ./internal/dsp/...                     # one package
go test ./internal/dsp/ -run TestSpectrogram   # one test
go test -bench=. ./internal/dsp/ ./internal/fingerprint/
```

The tests do not need a microphone or Spotify. HTTP and metadata tests use fakes for `storage.Store` and `metadata.MetadataProvider`. For manual testing, put your own audio files in `testdata/`. That directory is gitignored, and no audio files are committed.

### Contributing guidelines

- **Keep ingest and query in sync.** Catalog ingest and recognition must run the same pipeline with the same parameters, or no hashes will match. Both call `fingerprint.FromBuffer`, so change parameters only in `dsp.DefaultConfig()` and `internal/fingerprint/defaults.go`.
- **A pipeline change invalidates existing catalogs.** A change to the window function, FFT size, hop size, peak or landmark defaults, or hash layout changes every fingerprint. Delete your `.db` file and ingest again after such a change.
- **Keep stages separate.** Capture, DSP, fingerprinting, matching, storage, and metadata each stay in their own `internal/` package. The fingerprint pipeline must stay testable from audio files alone, with no hardware.
- **Add a subcommand** by defining a `<X>Cmd` struct with a `Run() error` method and adding it as a field of the `CLI` struct in `cmd/bassclef/main.go`.
- **Put HTTP code in `cmd/bassclef/api.go`.**
- Run `go vet ./...` and `go test ./...` before you open a pull request.

## Known limitations

- The catalog is a single SQLite file. It is not designed for a large catalog.
- Peak picking is a fixed-size local-maximum search with one absolute magnitude threshold. It does not yet use frequency bands, log-magnitude scaling, or adaptive thresholds.
- `listen` uses the system's default input device only.
- Matching does one database lookup per query hash. This is simple, but not fast for long queries.

## References

- Avery Wang, [*An Industrial-Strength Audio Search Algorithm*](https://www.researchgate.net/publication/220723446_An_Industrial_Strength_Audio_Search_Algorithm), 2003
- Lennevi, Sandh, Sundström, Önnermalm, [*Simulating Shazam*](https://odr.chalmers.se/server/api/core/bitstreams/e64dc2ee-cbb2-4ebc-8617-6a0987672f1d/content)
- [Xzam: An attempt at Shazam's Paper](https://medium.com/@xpanvictor/xzam-an-attempt-at-shazams-paper-84fe44901a7a)
- Roy van Rijn, [Creating Shazam in Java](https://www.royvanrijn.com/blog/2010/06/creating-shazam-in-java/), 2010
