package fingerprint

import (
	"fmt"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
)

// FromBuffer runs the full audio-to-landmarks fingerprinting pipeline -
// frame splitting, windowing, FFT, spectrogram, peak detection, and
// landmark generation - using DefaultPeakConfig and DefaultLandmarkConfig.
// catalog.Ingest and every CLI recognition path share this, so both sides
// of matching always use identical config.
func FromBuffer(buf *audio.AudioBuffer) ([]Landmark, error) {
	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		return nil, fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		return nil, fmt.Errorf("failed to build window: %w", err)
	}

	spectrogram, err := dsp.Spectrogram(frames, window, dsp.NewFFT(cfg.FFTSize))
	if err != nil {
		return nil, fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	peaks, err := FindPeaks(spectrogram, DefaultPeakConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to find peaks: %w", err)
	}

	landmarks, err := GenerateLandmarks(peaks, DefaultLandmarkConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to generate landmarks: %w", err)
	}

	return landmarks, nil
}
