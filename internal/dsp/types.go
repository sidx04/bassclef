package dsp

// Config holds the parameters that shape frame extraction and the FFT run
// over each frame: how many samples per frame (FFTWindowSize) and how far the
// window advances between frames (HopSize). HopSize < FFTWindowSize means frames
// overlap.
type Config struct {
	// FFTWindowSize is the number of samples per analysis frame, and the size of
	// the FFT run over it. Must be a size the chosen FFT implementation
	// supports (a power of two, for CmplxFFT).
	FFTWindowSize int
	// HopSize is the number of samples the frame window advances between
	// consecutive frames.
	HopSize int
	// Window is the tapering function applied on each frame of the audio signal
	// before any FFT is carried out. This is to mitigate [spectral leakage] in the
	// audio signal. 'HannWindow' is used by default.
	//
	// The entire fingerprint is affected by the window function. Changing this
	// requires re-ingestion of the entire catalouge.
	//
	// [spectral leakage]: https://dsp.stackexchange.com/questions/63405/spectral-leakage-in-laymans-terms
	Window WindowFunc
}

func DefaultConfig() Config {
	return Config{FFTWindowSize: 4096, HopSize: 2048, Window: HannWindow}
}
