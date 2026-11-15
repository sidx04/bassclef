package fingerprint

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
)

const (
	benchSampleRate = 44100
	benchSeconds    = 240
)

// Returns a deterministic song-length synthetic signal (240 s at 44.1 kHz):
// a few summed sines plus seeded noise. It stands in for real audio so the
// benchmarks don't depend on the gitignored testdata/ fixtures or on MP3
// decoding cost.
func createBenchBuffer() *audio.AudioBuffer {
	rng := rand.New(rand.NewPCG(1, 2))
	freqs := []float64{220, 440, 880, 1760, 3520}

	samples := make([]float64, benchSampleRate*benchSeconds)
	for i := range samples {
		t := float64(i) / benchSampleRate
		var s float64
		for _, f := range freqs {
			s += math.Sin(2 * math.Pi * f * t)
		}
		samples[i] = s/float64(len(freqs))*0.8 + (rng.Float64()*2-1)*0.2
	}

	return &audio.AudioBuffer{Samples: samples, SampleRate: benchSampleRate}
}

// Measures the full fingerprinting pipeline over a
// deterministic song-length synthetic signal (240 s at 44.1 kHz), so the
// FFT's share can be compared against gain, framing, peaks and landmarks.
//
// FromBuffer normalizes gain in place. NormalizeGain is idempotent, so
// reusing the buffer across iterations does the same work each time.
func BenchmarkFromBuffer(b *testing.B) {
	buf := createBenchBuffer()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := FromBuffer(buf); err != nil {
			b.Fatal(err)
		}
	}
}

// Measures landmark pairing alone. The setup runs the same steps as
// FromBuffer up to FindPeaks, once, outside the timed loop; only
// GenerateLandmarks is timed. GenerateLandmarks does not modify peaks, so
// every iteration gets the same input.
func BenchmarkGenerateLandmarks(b *testing.B) {
	buf := createBenchBuffer()
	dsp.NormalizeGain(buf.Samples)

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTWindowSize, cfg.HopSize)
	if err != nil {
		b.Fatal(err)
	}

	window, err := cfg.Window(cfg.FFTWindowSize)
	if err != nil {
		b.Fatal(err)
	}

	spectrogram, err := dsp.Spectrogram(frames, window, dsp.NewFFT(cfg.FFTWindowSize))
	if err != nil {
		b.Fatal(err)
	}

	peaks, err := FindPeaks(spectrogram, DefaultPeakConfig())
	if err != nil {
		b.Fatal(err)
	}

	landmarkCfg := DefaultLandmarkConfig()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := GenerateLandmarks(peaks, landmarkCfg); err != nil {
			b.Fatal(err)
		}
	}
}
