package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var decodersByExt = map[string]AudioDecoder{
	".wav": NewWAVDecoder(),
	".mp3": NewMP3Decoder(),
}

// Load opens the file at path and decodes it with the decoder matching
// its extension (case-insensitive). No format flag needed — every caller
// goes through this, not a per-format Load function.
func Load(path string) (*AudioBuffer, error) {
	ext := strings.ToLower(filepath.Ext(path))

	decoder, ok := decodersByExt[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported audio format: %q", ext)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening audio file: %w", err)
	}
	defer file.Close()

	buf, err := decoder.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error loading audio file: %w", err)
	}

	return buf, nil
}
