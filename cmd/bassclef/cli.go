package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/catalog"
	"github.com/sidx04/bassclef/internal/dsp"
	"github.com/sidx04/bassclef/internal/fingerprint"
	"github.com/sidx04/bassclef/internal/matcher"
	"github.com/sidx04/bassclef/internal/metadata"
	"github.com/sidx04/bassclef/internal/storage"
)

// newMetadataProvider builds the Spotify metadata provider. Missing or
// invalid credentials are not fatal to recognition — only enrichment is
// skipped, with a one-line warning.
func newMetadataProvider() metadata.MetadataProvider {
	provider, err := metadata.NewSpotifyProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: Spotify metadata unavailable: %v\n", err)
		return nil
	}

	return provider
}

type InfoCmd struct {
	File string `arg:"" name:"file" help:"Path to the audio file." type:"path"`
}

func (cmd *InfoCmd) Run() error {
	buf, err := audio.Load(cmd.File)
	if err != nil {
		log.Fatalf("failed to load audio: %v", err)
	}

	duration := float64(len(buf.Samples)) / float64(buf.SampleRate)

	fmt.Printf("Sample rate: %d Hz\n", buf.SampleRate)
	fmt.Printf("Samples: %d\n", len(buf.Samples))
	fmt.Printf("Duration: %.2f seconds\n", duration)

	return nil
}

type SpectrogramCmd struct {
	File string `arg:"" name:"file" help:"Path to the audio file." type:"path"`
}

func (cmd *SpectrogramCmd) Run() error {
	buf, err := audio.Load(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTWindowSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := cfg.Window(cfg.FFTWindowSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTWindowSize)

	spectrogram, err := dsp.Spectrogram(frames, window, fft)
	if err != nil {
		return fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	fmt.Printf("Frames: %d\n", len(spectrogram))
	fmt.Printf("Bins per frame: %d\n", cfg.FFTWindowSize)

	return nil
}

type PeaksCmd struct {
	File         string  `arg:"" name:"file" help:"Path to the audio file." type:"path"`
	FreqWindow   int     `name:"freq-window" default:"3" help:"Bins each side of a candidate peak."`
	TimeWindow   int     `name:"time-window" default:"3" help:"Frames each side of a candidate peak."`
	MinMagnitude float64 `name:"min-magnitude" default:"0" help:"Reject candidates below this magnitude."`
}

func (cmd *PeaksCmd) Run() error {
	buf, err := audio.Load(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTWindowSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := cfg.Window(cfg.FFTWindowSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTWindowSize)

	spectrogram, err := dsp.Spectrogram(frames, window, fft)
	if err != nil {
		return fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	peakCfg := fingerprint.PeakConfig{
		FreqWindow:   cmd.FreqWindow,
		TimeWindow:   cmd.TimeWindow,
		MinMagnitude: cmd.MinMagnitude,
	}

	peaks, err := fingerprint.FindPeaks(spectrogram, peakCfg)
	if err != nil {
		return fmt.Errorf("failed to find peaks: %w", err)
	}

	fmt.Printf("Peaks: %d\n", len(peaks))

	for _, p := range peaks {
		hz := fingerprint.BinToHz(p.Frequency, buf.SampleRate, cfg.FFTWindowSize)
		fmt.Printf("time=%d freq=%.1fHz magnitude=%.4f\n", p.Time, hz, p.Magnitude)
	}

	return nil
}

type FingerprintCmd struct {
	File string `arg:"" name:"file" help:"Path to the audio file." type:"path"`

	FreqWindow   int     `name:"freq-window" default:"3" help:"Bins each side of a candidate peak."`
	TimeWindow   int     `name:"time-window" default:"3" help:"Frames each side of a candidate peak."`
	MinMagnitude float64 `name:"min-magnitude" default:"0" help:"Reject peak candidates below this magnitude."`

	MinDeltaFrames int `name:"min-delta-frames" default:"1" help:"Smallest anchor-to-target gap, in frames."`
	MaxDeltaFrames int `name:"max-delta-frames" default:"50" help:"Largest anchor-to-target gap, in frames."`
	Fanout         int `name:"fanout" default:"10" help:"Max target peaks paired per anchor."`
}

func (cmd *FingerprintCmd) Run() error {
	buf, err := audio.Load(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTWindowSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := cfg.Window(cfg.FFTWindowSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTWindowSize)

	spectrogram, err := dsp.Spectrogram(frames, window, fft)
	if err != nil {
		return fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	peakCfg := fingerprint.PeakConfig{
		FreqWindow:   cmd.FreqWindow,
		TimeWindow:   cmd.TimeWindow,
		MinMagnitude: cmd.MinMagnitude,
	}

	peaks, err := fingerprint.FindPeaks(spectrogram, peakCfg)
	if err != nil {
		return fmt.Errorf("failed to find peaks: %w", err)
	}

	landmarkCfg := fingerprint.LandmarkConfig{
		MinDeltaFrames: cmd.MinDeltaFrames,
		MaxDeltaFrames: cmd.MaxDeltaFrames,
		Fanout:         cmd.Fanout,
	}

	landmarks, err := fingerprint.GenerateLandmarks(peaks, landmarkCfg)
	if err != nil {
		return fmt.Errorf("failed to generate landmarks: %w", err)
	}

	fmt.Printf("Peaks: %d\n", len(peaks))
	fmt.Printf("Landmarks: %d\n", len(landmarks))

	for _, lm := range landmarks {
		hash := fingerprint.Hash(uint32(lm.AnchorFrequency), uint32(lm.TargetFrequency), uint32(lm.DeltaTime))
		fmt.Printf("time=%d hash=%d\n", lm.AnchorTime, hash)
	}

	return nil
}

type RecognizeCmd struct {
	File     string `arg:"" name:"file" help:"Path to the query audio file." type:"path"`
	DB       string `name:"db" default:"bassclef.db" help:"Path to the catalog database." type:"path"`
	TopN     int    `name:"top-n" default:"5" help:"Max number of ranked candidates to show."`
	MinVotes int    `name:"min-votes" default:"1" help:"Minimum vote score to include a candidate."`
}

func (cmd *RecognizeCmd) Run() error {
	store, err := storage.Open(cmd.DB)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	buf, err := audio.Load(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	return recognizeBuffer(store, buf, cmd.TopN, cmd.MinVotes, newMetadataProvider())
}

// MatchResult is one ranked, display-ready recognition candidate: a
// catalog song plus its match score and (if available) Spotify metadata.
// Shared by the CLI's printed output and the HTTP API's JSON response.
type MatchResult struct {
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Score      int     `json:"score"`
	Confidence float64 `json:"confidence"`
	Album      string  `json:"album,omitempty"`
	SpotifyURL string  `json:"spotify_url,omitempty"`
	CoverURL   string  `json:"cover_url,omitempty"`
}

// recognizeCandidates fingerprints buf, matches it against store, and
// enriches each ranked candidate with Spotify metadata when available.
// provider may be nil (Spotify unavailable) — enrichment is then skipped
// for every candidate. A failed or missing metadata lookup is not an
// error; the candidate is returned without Album/SpotifyURL set.
func recognizeCandidates(store storage.Store, buf *audio.AudioBuffer, topN, minVotes int, provider metadata.MetadataProvider) ([]MatchResult, error) {
	landmarks, err := fingerprint.FromBuffer(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to fingerprint audio: %w", err)
	}

	query := make([]matcher.QueryLandmark, len(landmarks))
	for i, lm := range landmarks {
		query[i] = matcher.QueryLandmark{
			Hash: fingerprint.Hash(uint32(lm.AnchorFrequency), uint32(lm.TargetFrequency), uint32(lm.DeltaTime)),
			Time: lm.AnchorTime,
		}
	}

	candidates, err := matcher.Match(store, query, matcher.MatchConfig{TopN: topN, MinVotes: minVotes})
	if err != nil {
		return nil, fmt.Errorf("failed to match: %w", err)
	}

	var results []MatchResult

	for _, c := range candidates {
		song, err := store.GetSong(c.SongID)
		if err != nil {
			return nil, fmt.Errorf("failed to get song %d: %w", c.SongID, err)
		}

		result := MatchResult{
			Title:      song.Title,
			Artist:     song.Artist,
			Score:      c.Score,
			Confidence: c.Confidence,
		}

		if provider != nil {
			if meta, err := provider.Lookup(song.Title, song.Artist); err == nil {
				result.Album = meta.Album
				result.SpotifyURL = meta.URL
				result.CoverURL = meta.CoverURL
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// recognizeBuffer fingerprints buf and matches it against store, printing
// ranked results. Shared by every CLI recognition path (WAV/MP3 file,
// microphone) so they can never drift from each other's pipeline or
// output format.
func recognizeBuffer(store storage.Store, buf *audio.AudioBuffer, topN, minVotes int, provider metadata.MetadataProvider) error {
	results, err := recognizeCandidates(store, buf, topN, minVotes, provider)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No match found.")
		return nil
	}

	for i, r := range results {
		fmt.Printf("%d. %s - %s (score=%d, confidence=%.4f)\n", i+1, r.Title, r.Artist, r.Score, r.Confidence)

		if r.Album != "" {
			fmt.Printf("   Album: %s\n", r.Album)
		}
		if r.SpotifyURL != "" {
			fmt.Printf("   Spotify: %s\n", r.SpotifyURL)
		}
	}

	return nil
}

type IngestCmd struct {
	File   string `arg:"" name:"file" help:"Path to the audio file." type:"path"`
	Title  string `name:"title" required:"" help:"Song title."`
	Artist string `name:"artist" help:"Song artist."`
	DB     string `name:"db" default:"bassclef.db" help:"Path to the catalog database." type:"path"`
}

func (cmd *IngestCmd) Run() error {
	store, err := storage.Open(cmd.DB)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	if err := catalog.Ingest(store, cmd.File, cmd.Title, cmd.Artist); err != nil {
		return fmt.Errorf("failed to ingest song: %w", err)
	}

	fmt.Printf("Ingested %q\n", cmd.File)

	return nil
}

type ListenCmd struct {
	Duration time.Duration `name:"duration" default:"8s" help:"How long to record from the microphone."`
	DB       string        `name:"db" default:"bassclef.db" help:"Path to the catalog database." type:"path"`
	TopN     int           `name:"top-n" default:"5" help:"Max number of ranked candidates to show."`
	MinVotes int           `name:"min-votes" default:"1" help:"Minimum vote score to include a candidate."`
}

func (cmd *ListenCmd) Run() error {
	store, err := storage.Open(cmd.DB)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	fmt.Printf("Listening for %s...\n", cmd.Duration)

	buf, err := audio.Capture(cmd.Duration)
	if err != nil {
		return fmt.Errorf("failed to capture audio: %w", err)
	}

	return recognizeBuffer(store, buf, cmd.TopN, cmd.MinVotes, newMetadataProvider())
}
