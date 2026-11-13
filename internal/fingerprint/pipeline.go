package fingerprint

import (
	"fmt"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
)

// Generates audio landmarks via the fingerprinting pipeline -
// gain normalization, frame splitting, windowing, FFT, spectrogram, peak
// detection, and landmark generation. Check [DefaultPeakConfig] and
// [DefaultLandmarkConfig] for peak detection and landmarking defaults.
//
// [DefaultPeakConfig]: https://github.com/sidx04/bassclef/blob/main/internal/fingerprint/defaults.go#L14
// [DefaultLandmarkConfig]: https://github.com/sidx04/bassclef/blob/main/internal/fingerprint/defaults.go#L20
func FromBuffer(buf *audio.AudioBuffer) ([]Landmark, error) {
	dsp.NormalizeGain(buf.Samples)

	cfg := dsp.DefaultConfig() // defaults of [FFTSize: 4096, HopSize: 2048]

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
