package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS songs (
		id       INTEGER PRIMARY KEY,
		title    TEXT NOT NULL,
		artist   TEXT,
		duration REAL,
		source   TEXT UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS fingerprints (
		hash        INTEGER NOT NULL,
		song_id     INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
		time_offset INTEGER NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_fingerprint_hash ON fingerprints(hash)`,
}

// SQLiteStore is a Store backed by a local SQLite file.
type SQLiteStore struct {
	db *sql.DB
}

// Open creates the database file if it doesn't exist, and ensures the
// schema and foreign-key enforcement are in place.
func Open(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// A single connection keeps PRAGMA foreign_keys in effect for every
	// statement (database/sql pools connections, and pragmas are set
	// per-connection) and avoids write contention against one file.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to initialize schema: %w", err)
		}
	}

	return &SQLiteStore{db: db}, nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// InsertSong adds a song row and returns its assigned id. Fails if
// song.Source duplicates an existing row (songs.source is UNIQUE).
func (s *SQLiteStore) InsertSong(song Song) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO songs (title, artist, duration, source) VALUES (?, ?, ?, ?)",
		song.Title, song.Artist, song.Duration, song.Source,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert song: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted song id: %w", err)
	}

	return id, nil
}

// InsertFingerprints writes all fps for songID inside one transaction:
// either every row lands, or none do.
func (s *SQLiteStore) InsertFingerprints(songID int64, fps []Fingerprint) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO fingerprints (hash, song_id, time_offset) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, fp := range fps {
		if _, err := stmt.Exec(int64(fp.Hash), songID, fp.TimeOffset); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert fingerprint: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LookupHash returns every (song, time offset) a hash was recorded at.
func (s *SQLiteStore) LookupHash(hash uint64) ([]FingerprintMatch, error) {
	rows, err := s.db.Query(
		"SELECT song_id, time_offset FROM fingerprints WHERE hash = ?",
		int64(hash),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query fingerprints: %w", err)
	}
	defer rows.Close()

	var matches []FingerprintMatch

	for rows.Next() {
		var m FingerprintMatch
		if err := rows.Scan(&m.SongID, &m.TimeOffset); err != nil {
			return nil, fmt.Errorf("failed to scan fingerprint match: %w", err)
		}
		matches = append(matches, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading fingerprint matches: %w", err)
	}

	return matches, nil
}

// GetSong fetches one song by id.
func (s *SQLiteStore) GetSong(id int64) (Song, error) {
	song := Song{ID: id}

	err := s.db.QueryRow(
		"SELECT title, artist, duration, source FROM songs WHERE id = ?",
		id,
	).Scan(&song.Title, &song.Artist, &song.Duration, &song.Source)
	if err != nil {
		return Song{}, fmt.Errorf("failed to get song %d: %w", id, err)
	}

	return song, nil
}
