package dsp

// Config holds the parameters that shape frame extraction and the FFT run
// over each frame: how many samples per frame (FFTSize) and how far the
// window advances between frames (HopSize). HopSize < FFTSize means frames
// overlap.
type Config struct {
	// FFTSize is the number of samples per analysis frame, and the size of
	// the FFT run over it. Must be a size the chosen FFT implementation
	// supports (a power of two, for CmplxFFT).
	FFTSize int
	// HopSize is the number of samples the frame window advances between
	// consecutive frames.
	HopSize int
}

func DefaultConfig() Config {
	return Config{FFTSize: 4096, HopSize: 2048}
}
