package dsp

import (
	"math"
	"math/rand/v2"
	"testing"
)

const (
	benchSampleRate = 44100
	benchSeconds    = 240 // about the length of a typical song, 4 mins
)

// Returns a deterministic synthetic song-length signal: a few
// summed sines plus seeded noise. It stands in for real audio so the
// benchmarks don't depend on the gitignored testdata/ fixtures or on MP3
// decoding cost.
func createBenchSignals(numSamples int) []float64 {
	rng := rand.New(rand.NewPCG(1, 2))
	freqs := []float64{220, 440, 880, 1760, 3520}

	samples := make([]float64, numSamples)
	for i := range samples {
		t := float64(i) / benchSampleRate
		var s float64
		for _, f := range freqs {
			s += math.Sin(2 * math.Pi * f * t)
		}
		samples[i] = s/float64(len(freqs))*0.8 + (rng.Float64()*2-1)*0.2
	}
	return samples
}

// BenchmarkForward measures one FFT over a single windowed frame at the
// default FFT size, isolating the transform from the rest of the pipeline.
func BenchmarkForward(b *testing.B) {
	cfg := DefaultConfig() // FFTWindowSize: 4096, HopSize: 2048

	window, err := HannWindow(cfg.FFTWindowSize)
	if err != nil {
		b.Fatal(err)
	}

	frame := createBenchSignals(cfg.FFTWindowSize)

	if err := applyWindow(frame, window); err != nil {
		b.Fatal(err)
	}

	fft := NewFFT(cfg.FFTWindowSize)

	b.ReportAllocs()
	for b.Loop() {
		fft.Forward(frame)
	}
}

// BenchmarkSpectrogram measures the spectrogram of a song-length signal
// (about 3,900 frames at the default config). Beyond the FFT itself, it
// captures the per-frame copy, windowing, and magnitude allocations.
func BenchmarkSpectrogram(b *testing.B) {
	cfg := DefaultConfig()

	frames, err := SplitFrames(createBenchSignals(benchSampleRate*benchSeconds), cfg.FFTWindowSize, cfg.HopSize)
	if err != nil {
		b.Fatal(err)
	}

	window, err := HannWindow(cfg.FFTWindowSize)
	if err != nil {
		b.Fatal(err)
	}

	fft := NewFFT(cfg.FFTWindowSize)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := Spectrogram(frames, window, fft); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSplitFramesCopy measures SplitFrames on a song-length signal at
// the default config. It captures the per-frame allocation and copy.
func BenchmarkSplitFrames(b *testing.B) {
	cfg := DefaultConfig()
	samples := createBenchSignals(benchSampleRate * benchSeconds)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := SplitFrames(samples, cfg.FFTWindowSize, cfg.HopSize); err != nil {
			b.Fatal(err)
		}
	}
}
