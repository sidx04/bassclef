package dsp

import "gonum.org/v1/gonum/dsp/fourier"

// FFT decouples the DSP pipeline from any one FFT implementation. Forward
// runs a forward transform over one frame of real-valued samples and
// returns the full complex spectrum (length equal to len(input); bins are
// conjugate-symmetric for real input, per the standard DFT).
type FFT interface {
	Forward(input []float64) []complex128
}

// CmplxFFT is an FFT backed by gonum's complex-to-complex FFT. Construct
// with NewFFT; the size passed there must match the length of every input
// slice later passed to Forward.
type CmplxFFT struct {
	fft *fourier.CmplxFFT
}

func NewFFT(size int) *CmplxFFT {
	return &CmplxFFT{fft: fourier.NewCmplxFFT(size)}
}

func (c *CmplxFFT) Forward(input []float64) []complex128 {
	complexInput := make([]complex128, len(input))
	for i, v := range input {
		complexInput[i] = complex(v, 0)
	}

	return c.fft.Coefficients(nil, complexInput)
}
