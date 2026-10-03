package fingerprint

// DefaultPeakConfig is the peak-detection configuration used by catalog
// ingest and query-side recognition. Both sides must use the same config,
// or hashes computed from their landmarks will never align.
//
// MinMagnitude: 1.0 cuts noise-floor local maxima before they become
// peaks. Measured against real tracks, ~75-90% of magnitude-0 "peaks" sit
// below 1.0 — near-silent bins that are only a local max because their
// neighbors are quieter still. Left uncut, a bright/percussive track can
// generate 3x+ the peak density of a quieter one purely from this noise,
// which skews catalog matching toward whichever song has the most
// fingerprints rather than the one that actually matches.
func DefaultPeakConfig() PeakConfig {
	return PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 1.0}
}

// DefaultLandmarkConfig is the landmark-generation configuration used by
// catalog ingest and query-side recognition; see DefaultPeakConfig.
func DefaultLandmarkConfig() LandmarkConfig {
	return LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 50, Fanout: 10}
}
