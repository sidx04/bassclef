package storage

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestOpenCreatesSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	var name string

	err = store.db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='songs'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("songs table not created: %v", err)
	}

	err = store.db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='fingerprints'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("fingerprints table not created: %v", err)
	}
}

func TestInsertSongReturnsID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	id, err := store.InsertSong(Song{Title: "A", Artist: "B", Duration: 1.5, Source: "a.wav"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id == 0 {
		t.Error("expected nonzero id")
	}
}

func TestInsertSongDuplicateSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	song := Song{Title: "A", Source: "same.wav"}

	if _, err := store.InsertSong(song); err != nil {
		t.Fatalf("unexpected error on first insert: %v", err)
	}

	if _, err := store.InsertSong(song); err == nil {
		t.Fatal("expected an error on duplicate source")
	}
}

func TestInsertFingerprintsAndLookupHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	songID, err := store.InsertSong(Song{Title: "A", Source: "a.wav"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fps := []Fingerprint{
		{Hash: 42, SongID: songID, TimeOffset: 5},
		{Hash: 43, SongID: songID, TimeOffset: 6},
	}

	if err := store.InsertFingerprints(songID, fps); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	matches, err := store.LookupHash(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []FingerprintMatch{{SongID: songID, TimeOffset: 5}}

	if !reflect.DeepEqual(matches, want) {
		t.Errorf("got %#v, want %#v", matches, want)
	}
}

func TestLookupHashUnknown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	matches, err := store.LookupHash(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("got %#v, want empty", matches)
	}
}

func TestGetSong(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	id, err := store.InsertSong(Song{Title: "A", Artist: "B", Duration: 1.5, Source: "a.wav"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.GetSong(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Song{ID: id, Title: "A", Artist: "B", Duration: 1.5, Source: "a.wav"}

	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFingerprintsCascadeDeleteWithSong(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	songID, err := store.InsertSong(Song{Title: "A", Source: "a.wav"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := store.InsertFingerprints(songID, []Fingerprint{{Hash: 1, SongID: songID, TimeOffset: 0}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := store.db.Exec("DELETE FROM songs WHERE id = ?", songID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	matches, err := store.LookupHash(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("expected cascade delete to remove fingerprints, got %#v", matches)
	}
}

func TestOpenEnablesForeignKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	var enabled int
	if err := store.db.QueryRow("PRAGMA foreign_keys").Scan(&enabled); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if enabled != 1 {
		t.Errorf("foreign keys not enabled: got %d", enabled)
	}
}
