package main

import (
	"fmt"
	"log"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/catalog"
	"github.com/sidx04/bassclef/internal/dsp"
	"github.com/sidx04/bassclef/internal/fingerprint"
	"github.com/sidx04/bassclef/internal/matcher"
	"github.com/sidx04/bassclef/internal/storage"
)

type InfoCmd struct {
	File string `arg:"" name:"file" help:"Path to the audio file." type:"path"`
}

func (cmd *InfoCmd) Run() error {
	buf, err := audio.LoadWAV(cmd.File)
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
	buf, err := audio.LoadWAV(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTSize)

	spectrogram, err := dsp.Spectrogram(frames, window, fft)
	if err != nil {
		return fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	fmt.Printf("Frames: %d\n", len(spectrogram))
	fmt.Printf("Bins per frame: %d\n", cfg.FFTSize)

	return nil
}

type PeaksCmd struct {
	File         string  `arg:"" name:"file" help:"Path to the audio file." type:"path"`
	FreqWindow   int     `name:"freq-window" default:"3" help:"Bins each side of a candidate peak."`
	TimeWindow   int     `name:"time-window" default:"3" help:"Frames each side of a candidate peak."`
	MinMagnitude float64 `name:"min-magnitude" default:"0" help:"Reject candidates below this magnitude."`
}

func (cmd *PeaksCmd) Run() error {
	buf, err := audio.LoadWAV(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTSize)

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
		hz := fingerprint.BinToHz(p.Frequency, buf.SampleRate, cfg.FFTSize)
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
	buf, err := audio.LoadWAV(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTSize)

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

	buf, err := audio.LoadWAV(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	spectrogram, err := dsp.Spectrogram(frames, window, dsp.NewFFT(cfg.FFTSize))
	if err != nil {
		return fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	peaks, err := fingerprint.FindPeaks(spectrogram, fingerprint.DefaultPeakConfig())
	if err != nil {
		return fmt.Errorf("failed to find peaks: %w", err)
	}

	landmarks, err := fingerprint.GenerateLandmarks(peaks, fingerprint.DefaultLandmarkConfig())
	if err != nil {
		return fmt.Errorf("failed to generate landmarks: %w", err)
	}

	query := make([]matcher.QueryLandmark, len(landmarks))
	for i, lm := range landmarks {
		query[i] = matcher.QueryLandmark{
			Hash: fingerprint.Hash(uint32(lm.AnchorFrequency), uint32(lm.TargetFrequency), uint32(lm.DeltaTime)),
			Time: lm.AnchorTime,
		}
	}

	candidates, err := matcher.Match(store, query, matcher.MatchConfig{TopN: cmd.TopN, MinVotes: cmd.MinVotes})
	if err != nil {
		return fmt.Errorf("failed to match: %w", err)
	}

	if len(candidates) == 0 {
		fmt.Println("No match found.")
		return nil
	}

	for i, c := range candidates {
		song, err := store.GetSong(c.SongID)
		if err != nil {
			return fmt.Errorf("failed to get song %d: %w", c.SongID, err)
		}

		fmt.Printf("%d. %s - %s (score=%d)\n", i+1, song.Title, song.Artist, c.Score)
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
