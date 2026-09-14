package matcher

import (
	"reflect"
	"testing"

	"github.com/sidx04/bassclef/internal/storage"
)

type fakeStore struct {
	matches map[uint64][]storage.FingerprintMatch
}

func (f *fakeStore) InsertSong(storage.Song) (int64, error)                { return 0, nil }
func (f *fakeStore) InsertFingerprints(int64, []storage.Fingerprint) error { return nil }
func (f *fakeStore) GetSong(int64) (storage.Song, error)                   { return storage.Song{}, nil }
func (f *fakeStore) Close() error                                          { return nil }

func (f *fakeStore) LookupHash(hash uint64) ([]storage.FingerprintMatch, error) {
	return f.matches[hash], nil
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

	want := []Candidate{{SongID: 100, Score: 3}}

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
		{SongID: 100, Score: 1},
		{SongID: 200, Score: 1},
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

	want := []Candidate{{SongID: 200, Score: 2}}

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
