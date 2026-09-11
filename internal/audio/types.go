package audio

import "io"

// AudioBuffer is the canonical in-memory representation of decoded audio:
// mono PCM samples normalised to the range -1.0 <= sample <= 1.0, regardless
// of the source format's channel count or bit depth. Every decoder in this
// package produces this shape so downstream code (DSP, fingerprinting) never
// needs to know which source format the audio came from.
type AudioBuffer struct {
	// Samples holds one normalised amplitude per mono sample, in order.
	Samples []float64
	// SampleRate is the number of samples per second, in Hz.
	SampleRate int
}

// AudioDecoder decodes an audio stream into an AudioBuffer. Implementations
// handle one source format (see WAVDecoder); adding a new format means
// adding a new AudioDecoder implementation, not changing callers.
type AudioDecoder interface {
	Decode(r io.Reader) (*AudioBuffer, error)
}

// WAVDecoder is an AudioDecoder for PCM WAV audio. Use NewWAVDecoder to
// construct one, or LoadWAV as a file-path convenience wrapper.
type WAVDecoder struct{}
