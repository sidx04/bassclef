package main

import (
	"errors"
	"math"
	"testing"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/fingerprint"
	"github.com/sidx04/bassclef/internal/metadata"
	"github.com/sidx04/bassclef/internal/storage"
)

// fakeStore is a storage.Store usable in both cli_test.go and
// api_test.go (same package) without touching a real database.
type fakeStore struct {
	songs            map[int64]storage.Song
	matches          map[uint64][]storage.FingerprintMatch
	fingerprintCount int
	lookupErr        error
	getSongErr       error
}

func (f *fakeStore) InsertSong(storage.Song) (int64, error)                { return 0, nil }
func (f *fakeStore) InsertFingerprints(int64, []storage.Fingerprint) error { return nil }
func (f *fakeStore) Close() error                                          { return nil }

func (f *fakeStore) LookupHash(hash uint64) ([]storage.FingerprintMatch, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	return f.matches[hash], nil
}

func (f *fakeStore) GetSong(id int64) (storage.Song, error) {
	if f.getSongErr != nil {
		return storage.Song{}, f.getSongErr
	}
	return f.songs[id], nil
}

func (f *fakeStore) CountFingerprints(int64) (int, error) {
	if f.fingerprintCount == 0 {
		return 1, nil
	}
	return f.fingerprintCount, nil
}

type fakeProvider struct {
	meta metadata.Metadata
	err  error
}

func (f *fakeProvider) Lookup(title, artist string) (metadata.Metadata, error) {
	if f.err != nil {
		return metadata.Metadata{}, f.err
	}
	return f.meta, nil
}

func testBuffer() *audio.AudioBuffer {
	const sampleRate = 44100
	numSamples := sampleRate / 2

	samples := make([]float64, numSamples)
	for i := range samples {
		samples[i] = math.Sin(2 * math.Pi * 440 * float64(i) / float64(sampleRate))
	}

	return &audio.AudioBuffer{Samples: samples, SampleRate: sampleRate}
}

// selfMatchingStore builds a fakeStore whose fingerprints are exactly
// buf's own landmarks, tied to songID/song — so recognizing buf against
// this store is a guaranteed, genuine match, not a guessed one.
func selfMatchingStore(t *testing.T, buf *audio.AudioBuffer, songID int64, song storage.Song) *fakeStore {
	t.Helper()

	landmarks, err := fingerprint.FromBuffer(buf)
	if err != nil {
		t.Fatalf("failed to fingerprint buffer: %v", err)
	}
	if len(landmarks) == 0 {
		t.Fatal("buffer produced no landmarks; fixture needs adjusting")
	}

	matches := make(map[uint64][]storage.FingerprintMatch)
	for _, lm := range landmarks {
		hash := fingerprint.Hash(uint32(lm.AnchorFrequency), uint32(lm.TargetFrequency), uint32(lm.DeltaTime))
		matches[hash] = append(matches[hash], storage.FingerprintMatch{SongID: songID, TimeOffset: lm.AnchorTime})
	}

	return &fakeStore{
		songs:   map[int64]storage.Song{songID: song},
		matches: matches,
	}
}

func TestRecognizeCandidatesMatch(t *testing.T) {
	store := selfMatchingStore(t, testBuffer(), 1, storage.Song{ID: 1, Title: "Song A", Artist: "Artist A"})
	provider := &fakeProvider{meta: metadata.Metadata{Album: "Album A", URL: "http://spotify/a", CoverURL: "http://spotify/a/cover.jpg"}}

	results, err := recognizeCandidates(store, testBuffer(), 5, 1, provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	got := results[0]
	if got.Title != "Song A" || got.Artist != "Artist A" || got.Album != "Album A" || got.SpotifyURL != "http://spotify/a" || got.CoverURL != "http://spotify/a/cover.jpg" {
		t.Errorf("got %#v, unexpected fields", got)
	}
	if got.Score <= 0 {
		t.Errorf("got score %d, want positive", got.Score)
	}
}

func TestRecognizeCandidatesMetadataNotFound(t *testing.T) {
	store := selfMatchingStore(t, testBuffer(), 1, storage.Song{ID: 1, Title: "Song A", Artist: "Artist A"})
	provider := &fakeProvider{err: metadata.ErrNotFound}

	results, err := recognizeCandidates(store, testBuffer(), 5, 1, provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Album != "" || results[0].SpotifyURL != "" || results[0].CoverURL != "" {
		t.Errorf("got %#v, want empty Album/SpotifyURL on lookup failure", results[0])
	}
}

func TestRecognizeCandidatesNilProvider(t *testing.T) {
	store := selfMatchingStore(t, testBuffer(), 1, storage.Song{ID: 1, Title: "Song A", Artist: "Artist A"})

	results, err := recognizeCandidates(store, testBuffer(), 5, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Album != "" || results[0].SpotifyURL != "" || results[0].CoverURL != "" {
		t.Errorf("got %#v, want empty Album/SpotifyURL with nil provider", results[0])
	}
}

func TestRecognizeCandidatesNoMatches(t *testing.T) {
	store := &fakeStore{}

	results, err := recognizeCandidates(store, testBuffer(), 5, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("got %#v, want no results", results)
	}
}

func TestRecognizeCandidatesLookupHashError(t *testing.T) {
	store := &fakeStore{lookupErr: errors.New("db failure")}

	_, err := recognizeCandidates(store, testBuffer(), 5, 1, nil)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestRecognizeCandidatesGetSongError(t *testing.T) {
	store := selfMatchingStore(t, testBuffer(), 1, storage.Song{ID: 1, Title: "Song A"})
	store.getSongErr = errors.New("db failure")

	_, err := recognizeCandidates(store, testBuffer(), 5, 1, nil)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestRecognizeCandidatesInvalidBuffer(t *testing.T) {
	buf := &audio.AudioBuffer{Samples: []float64{0, 0, 0}, SampleRate: 44100}

	_, err := recognizeCandidates(&fakeStore{}, buf, 5, 1, nil)
	if err == nil {
		t.Fatal("expected an error")
	}
}
