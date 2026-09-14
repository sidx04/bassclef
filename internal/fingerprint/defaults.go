package fingerprint

// DefaultPeakConfig is the peak-detection configuration used by catalog
// ingest and query-side recognition. Both sides must use the same config,
// or hashes computed from their landmarks will never align.
func DefaultPeakConfig() PeakConfig {
	return PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 0}
}

// DefaultLandmarkConfig is the landmark-generation configuration used by
// catalog ingest and query-side recognition; see DefaultPeakConfig.
func DefaultLandmarkConfig() LandmarkConfig {
	return LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 50, Fanout: 10}
}
