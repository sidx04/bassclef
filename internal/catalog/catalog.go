package catalog

import (
	"fmt"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
	"github.com/sidx04/bassclef/internal/fingerprint"
	"github.com/sidx04/bassclef/internal/storage"
)

// Ingest runs the full fingerprinting pipeline against the WAV file at
// path and persists the song and its fingerprints to store.
func Ingest(store storage.Store, path, title, artist string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	buf, err := audio.LoadWAV(path)
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

	duration := float64(len(buf.Samples)) / float64(buf.SampleRate)

	songID, err := store.InsertSong(storage.Song{
		Title:    title,
		Artist:   artist,
		Duration: duration,
		Source:   path,
	})
	if err != nil {
		return fmt.Errorf("failed to insert song: %w", err)
	}

	fps := make([]storage.Fingerprint, len(landmarks))
	for i, lm := range landmarks {
		fps[i] = storage.Fingerprint{
			Hash:       fingerprint.Hash(uint32(lm.AnchorFrequency), uint32(lm.TargetFrequency), uint32(lm.DeltaTime)),
			SongID:     songID,
			TimeOffset: lm.AnchorTime,
		}
	}

	if err := store.InsertFingerprints(songID, fps); err != nil {
		return fmt.Errorf("failed to insert fingerprints: %w", err)
	}

	return nil
}
