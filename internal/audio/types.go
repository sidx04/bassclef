package audio

import "io"

// Records the audio buffer samples and sample rate,
// in a normalised representation of -1.0 <= sample <= 1.0
// so that it can work with any PCM
type AudioBuffer struct {
	Samples    []float64
	SampleRate int
}

type AudioDecoder interface {
	Decode(r io.Reader) (*AudioBuffer, error)
}

type WAVDecoder struct{}
