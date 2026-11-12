package audio

import "io"

// Canonical in-memory representation of decoded audio, which are
// mono PCM samples normalised to the range -1.0 <= sample <= 1.0, regardless
// of the source format's channel count or bit depth.
//
// Every decoder in this package produces this shape so downstream code
// (DSP, fingerprinting) never needs to know which source format the audio came from.
type AudioBuffer struct {
	// Holds one normalised amplitude per mono sample, in order.
	Samples []float64
	// The number of samples per second, in Hz.
	SampleRate int
}

// Decodes an audio stream into an AudioBuffer. Implementations span across
// [Wav] and [MP3].
//
// [Wav]: https://github.com/sidx04/bassclef/blob/main/internal/audio/wav.go#L10
// [MP3]: https://github.com/sidx04/bassclef/blob/main/internal/audio/mp3.go#L12
type AudioDecoder interface {
	Decode(r io.Reader) (*AudioBuffer, error)
}

// WAVDecoder is an AudioDecoder for PCM WAV audio. Use NewWAVDecoder to construct one.
type WAVDecoder struct{}

// MP3Decoder is an AudioDecoder for MP3 audio. Use NewMP3Decoder to construct one.
type MP3Decoder struct{}

// defaultCaptureSampleRate must match the sample rate WAV fixtures are ingested
// at (44100 Hz). Landmark time offsets are a sample-count-based unit,
// not a real-time one, so a mismatched capture rate
// would silently misalign matching against the catalog.
const defaultCaptureSampleRate = 44100 // in hertz
