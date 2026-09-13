package fingerprint

// BinToHz converts an FFT bin index into the frequency, in Hz, it
// represents for a transform of the given size run at the given sample
// rate.
func BinToHz(bin, sampleRate, fftSize int) float64 {
	return float64(bin) * float64(sampleRate) / float64(fftSize)
}
