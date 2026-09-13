package fingerprint

// Peak is a strong local maximum in the spectrogram: a candidate landmark
// point for fingerprinting.
type Peak struct {
	// Time is the frame index the peak occurs at.
	Time int
	// Frequency is the FFT bin index the peak occurs at. Use BinToHz to
	// convert to Hz.
	Frequency int
	Magnitude float64
}

// PeakConfig controls FindPeaks' local-maximum search.
type PeakConfig struct {
	// FreqWindow is the number of bins on each side of a candidate that
	// must not exceed it for the candidate to count as a peak.
	FreqWindow int
	// TimeWindow is the number of frames on each side of a candidate that
	// must not exceed it for the candidate to count as a peak.
	TimeWindow int
	// MinMagnitude rejects any candidate below this magnitude outright,
	// regardless of whether it's a local maximum.
	MinMagnitude float64
}
