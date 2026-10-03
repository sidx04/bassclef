package matcher

// QueryLandmark is one hash observed in a query clip, and the frame time
// (the landmark's AnchorTime) it was observed at.
type QueryLandmark struct {
	Hash uint64
	Time int
}

// Candidate is a ranked catalog song, its raw vote score, and its
// density-normalized Confidence (Score divided by the song's total
// fingerprint count). Ranking uses Confidence, not Score — a song with
// far more fingerprints than others has more chances of an incidental
// hash collision at some offset, so raw vote count alone biases toward
// dense songs regardless of true relevance. MinVotes still filters on the
// raw Score, not Confidence.
type Candidate struct {
	SongID     int64
	Score      int
	Confidence float64
}

// MatchConfig controls Match's ranking.
type MatchConfig struct {
	// TopN caps how many candidates Match returns.
	TopN int
	// MinVotes excludes any candidate scoring below this.
	MinVotes int
}
