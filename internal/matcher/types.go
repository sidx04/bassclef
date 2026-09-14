package matcher

// QueryLandmark is one hash observed in a query clip, and the frame time
// (the landmark's AnchorTime) it was observed at.
type QueryLandmark struct {
	Hash uint64
	Time int
}

// Candidate is a ranked catalog song and its vote score.
type Candidate struct {
	SongID int64
	Score  int
}

// MatchConfig controls Match's ranking.
type MatchConfig struct {
	// TopN caps how many candidates Match returns.
	TopN int
	// MinVotes excludes any candidate scoring below this.
	MinVotes int
}
