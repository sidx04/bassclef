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

// Landmark pairs an anchor peak with a target peak that follows it in
// time: the acoustic relationship (AnchorFrequency, TargetFrequency,
// DeltaTime) is what gets hashed, with AnchorTime kept alongside for
// time-offset voting during matching.
type Landmark struct {
	AnchorFrequency int
	TargetFrequency int
	// DeltaTime is TargetTime - AnchorTime, always positive: targets only
	// follow anchors, since a song only plays forward.
	DeltaTime  int
	AnchorTime int
}

// LandmarkConfig controls GenerateLandmarks' anchor/target pairing.
type LandmarkConfig struct {
	// MinDeltaFrames is the smallest anchor-to-target gap, in frames, that
	// counts as a pairing.
	MinDeltaFrames int
	// MaxDeltaFrames is the largest anchor-to-target gap, in frames, that
	// counts as a pairing.
	MaxDeltaFrames int
	// Fanout caps how many target peaks one anchor pairs with; when more
	// candidates fall in the window, the nearest in time win.
	Fanout int
}
