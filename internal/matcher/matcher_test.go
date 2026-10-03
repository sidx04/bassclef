package matcher

import (
	"reflect"
	"testing"

	"github.com/sidx04/bassclef/internal/storage"
)

type fakeStore struct {
	matches map[uint64][]storage.FingerprintMatch
	// fingerprintCounts overrides CountFingerprints per song id; a song
	// not present here defaults to a count of 1, so Confidence equals
	// Score for tests that don't care about density normalization.
	fingerprintCounts map[int64]int
}

func (f *fakeStore) InsertSong(storage.Song) (int64, error)                { return 0, nil }
func (f *fakeStore) InsertFingerprints(int64, []storage.Fingerprint) error { return nil }
func (f *fakeStore) GetSong(int64) (storage.Song, error)                   { return storage.Song{}, nil }
func (f *fakeStore) Close() error                                          { return nil }

func (f *fakeStore) LookupHash(hash uint64) ([]storage.FingerprintMatch, error) {
	return f.matches[hash], nil
}

func (f *fakeStore) CountFingerprints(songID int64) (int, error) {
	if count, ok := f.fingerprintCounts[songID]; ok {
		return count, nil
	}
	return 1, nil
}

func TestMatchSingleWinner(t *testing.T) {
	store := &fakeStore{
		matches: map[uint64][]storage.FingerprintMatch{
			1: {{SongID: 100, TimeOffset: 10}},
			2: {{SongID: 100, TimeOffset: 11}},
			3: {{SongID: 100, TimeOffset: 12}},
		},
	}

	query := []QueryLandmark{
		{Hash: 1, Time: 0},
		{Hash: 2, Time: 1},
		{Hash: 3, Time: 2},
	}

	got, err := Match(store, query, MatchConfig{TopN: 5, MinVotes: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Candidate{{SongID: 100, Score: 3, Confidence: 3}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMatchTieBreakBySongID(t *testing.T) {
	store := &fakeStore{
		matches: map[uint64][]storage.FingerprintMatch{
			1: {{SongID: 200, TimeOffset: 5}, {SongID: 100, TimeOffset: 5}},
		},
	}

	query := []QueryLandmark{{Hash: 1, Time: 0}}

	got, err := Match(store, query, MatchConfig{TopN: 5, MinVotes: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Candidate{
		{SongID: 100, Score: 1, Confidence: 1},
		{SongID: 200, Score: 1, Confidence: 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMatchMinVotesExcludesWeak(t *testing.T) {
	store := &fakeStore{
		matches: map[uint64][]storage.FingerprintMatch{
			1: {{SongID: 100, TimeOffset: 5}},
			2: {{SongID: 200, TimeOffset: 5}},
			3: {{SongID: 200, TimeOffset: 5}},
		},
	}

	query := []QueryLandmark{
		{Hash: 1, Time: 0},
		{Hash: 2, Time: 0},
		{Hash: 3, Time: 0},
	}

	got, err := Match(store, query, MatchConfig{TopN: 5, MinVotes: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Candidate{{SongID: 200, Score: 2, Confidence: 2}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMatchTopNTruncates(t *testing.T) {
	store := &fakeStore{
		matches: map[uint64][]storage.FingerprintMatch{
			1: {
				{SongID: 100, TimeOffset: 5},
				{SongID: 200, TimeOffset: 5},
				{SongID: 300, TimeOffset: 5},
			},
		},
	}

	query := []QueryLandmark{{Hash: 1, Time: 0}}

	got, err := Match(store, query, MatchConfig{TopN: 2, MinVotes: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d candidates, want 2", len(got))
	}
}

func TestMatchRanksByConfidenceNotRawScore(t *testing.T) {
	// Song 100 (dense, 1,000,000 fingerprints) gets more raw votes than
	// song 200 (sparse, 100 fingerprints), purely from having more
	// fingerprints to incidentally collide with — but song 200 matches a
	// far larger fraction of its own fingerprints, so it must rank first.
	matches := make(map[uint64][]storage.FingerprintMatch)
	var query []QueryLandmark

	for i := uint64(0); i < 100; i++ {
		matches[i] = []storage.FingerprintMatch{{SongID: 100, TimeOffset: 5}}
		query = append(query, QueryLandmark{Hash: i, Time: 0})
	}

	for i := uint64(100); i < 150; i++ {
		matches[i] = []storage.FingerprintMatch{{SongID: 200, TimeOffset: 7}}
		query = append(query, QueryLandmark{Hash: i, Time: 0})
	}

	store := &fakeStore{
		matches: matches,
		fingerprintCounts: map[int64]int{
			100: 1_000_000,
			200: 100,
		},
	}

	got, err := Match(store, query, MatchConfig{TopN: 5, MinVotes: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d candidates, want 2", len(got))
	}

	if got[0].SongID != 200 {
		t.Errorf("got top candidate song %d, want 200 (higher confidence despite lower raw score)", got[0].SongID)
	}

	if got[0].Score != 50 || got[1].Score != 100 {
		t.Errorf("got scores %d, %d, want raw score 50 (song 200) then 100 (song 100)", got[0].Score, got[1].Score)
	}

	if got[0].Confidence <= got[1].Confidence {
		t.Errorf("got confidence %v, %v, want song 200's confidence higher", got[0].Confidence, got[1].Confidence)
	}
}

func TestMatchEmptyQuery(t *testing.T) {
	got, err := Match(&fakeStore{}, nil, MatchConfig{TopN: 5, MinVotes: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("got %#v, want nil", got)
	}
}

func TestMatchInvalidConfig(t *testing.T) {
	store := &fakeStore{}
	query := []QueryLandmark{{Hash: 1, Time: 0}}

	tests := []struct {
		name string
		cfg  MatchConfig
	}{
		{"zero top n", MatchConfig{TopN: 0, MinVotes: 1}},
		{"negative top n", MatchConfig{TopN: -1, MinVotes: 1}},
		{"negative min votes", MatchConfig{TopN: 5, MinVotes: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Match(store, query, tt.cfg)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
