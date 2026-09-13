package fingerprint

import "fmt"

// FindPeaks scans a spectrogram for local-maximum peaks: points whose
// magnitude is at least MinMagnitude and no smaller than any neighbor
// within config's time/frequency window. Windows are clamped at the
// spectrogram's edges rather than padded.
func FindPeaks(spectrogram [][]float64, cfg PeakConfig) ([]Peak, error) {
	if len(spectrogram) == 0 {
		return nil, fmt.Errorf("spectrogram cannot be empty")
	}

	if cfg.FreqWindow <= 0 {
		return nil, fmt.Errorf("freq window must be positive: %d", cfg.FreqWindow)
	}

	if cfg.TimeWindow <= 0 {
		return nil, fmt.Errorf("time window must be positive: %d", cfg.TimeWindow)
	}

	var peaks []Peak

	for t := range spectrogram {
		for f := range spectrogram[t] {
			magnitude := spectrogram[t][f]

			if magnitude < cfg.MinMagnitude {
				continue
			}

			if isLocalMax(spectrogram, t, f, cfg) {
				peaks = append(peaks, Peak{Time: t, Frequency: f, Magnitude: magnitude})
			}
		}
	}

	return peaks, nil
}

func isLocalMax(spectrogram [][]float64, t, f int, cfg PeakConfig) bool {
	candidate := spectrogram[t][f]

	minT := max(0, t-cfg.TimeWindow)
	maxT := min(len(spectrogram)-1, t+cfg.TimeWindow)

	for nt := minT; nt <= maxT; nt++ {
		minF := max(0, f-cfg.FreqWindow)
		maxF := min(len(spectrogram[nt])-1, f+cfg.FreqWindow)

		for nf := minF; nf <= maxF; nf++ {
			if nt == t && nf == f {
				continue
			}

			if spectrogram[nt][nf] > candidate {
				return false
			}
		}
	}

	return true
}
