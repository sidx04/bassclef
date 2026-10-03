package metadata

import "errors"

// ErrNotFound is returned by MetadataProvider.Lookup when no matching
// track was found.
var ErrNotFound = errors.New("no matching track found")

// Metadata is the external information a MetadataProvider can attach to a
// recognized song.
type Metadata struct {
	Title  string
	Artist string
	Album  string
	URL    string
}

// MetadataProvider looks up external metadata for a song by title and
// artist, keeping recognition independent of any one external service.
type MetadataProvider interface {
	Lookup(title, artist string) (Metadata, error)
}

var _ MetadataProvider = (*SpotifyProvider)(nil)
