package storage

// Song is one catalog entry.
type Song struct {
	ID       int64
	Title    string
	Artist   string
	Duration float64
	Source   string
}

// Fingerprint is one landmark hash tied to a song and the frame time it
// occurred at.
type Fingerprint struct {
	Hash       uint64
	SongID     int64
	TimeOffset int
}

// FingerprintMatch is one row returned by LookupHash: which song a hash
// was seen in, and at what time offset.
type FingerprintMatch struct {
	SongID     int64
	TimeOffset int
}

// Store is the persistence boundary for the catalog and its fingerprints.
// SQLiteStore is the only implementation today; the interface exists so
// catalog and matcher never depend on the concrete database.
type Store interface {
	InsertSong(song Song) (int64, error)
	InsertFingerprints(songID int64, fps []Fingerprint) error
	LookupHash(hash uint64) ([]FingerprintMatch, error)
	GetSong(id int64) (Song, error)
	Close() error
}

var _ Store = (*SQLiteStore)(nil)
