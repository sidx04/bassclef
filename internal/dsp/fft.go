package dsp

import "gonum.org/v1/gonum/dsp/fourier"

// FFT decouples the DSP pipeline from any one FFT implementation.
type FFT interface {
	// Runs a forward transform: time domain to frequency domain
	// over one frame of real-valued audio samples.
	// Returns only the non-redundant half of the spectrum: len(input)/2+1 bins,
	// from 0 Hz to the Nyquist frequency.
	// For real input the upper bins are the complex conjugates of the
	// lower ones, so they carry no extra information.
	//
	// If dst is nil, Forward allocates and returns a new slice. Else, it
	// writes the spectrum into dst and returns it, so one buffer can be
	// reused across frames.
	// A non-nil dst must have the length of a previous result,
	// which is len(input)/2+1 for RealFFT.
	Forward(dst []complex128, input []float64) []complex128
}

// RealFFT is a wrapper around gonum's real-to-complex FFT. Construct with
// NewFFT; the size passed there must match the length of every input slice
// later passed to Forward.
type RealFFT struct {
	fft *fourier.FFT
}

func NewFFT(size int) *RealFFT {
	return &RealFFT{fft: fourier.NewFFT(size)}
}

// Forward panics if len(input) differs from the size passed to NewFFT,
// or if dst is non-nil and its length is not (size/2+1).
func (r *RealFFT) Forward(dst []complex128, input []float64) []complex128 {
	return r.fft.Coefficients(dst, input)
}
