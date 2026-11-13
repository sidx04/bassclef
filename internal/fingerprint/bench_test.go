package fingerprint

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/sidx04/bassclef/internal/audio"
)

// Measures the full fingerprinting pipeline over a
// deterministic song-length synthetic signal (240 s at 44.1 kHz), so the
// FFT's share can be compared against gain, framing, peaks and landmarks.
//
// FromBuffer normalizes gain in place. NormalizeGain is idempotent, so
// reusing the buffer across iterations does the same work each time.
func BenchmarkFromBuffer(b *testing.B) {
	const (
		sampleRate = 44100
		seconds    = 240
	)

	rng := rand.New(rand.NewPCG(1, 2))
	freqs := []float64{220, 440, 880, 1760, 3520}

	samples := make([]float64, sampleRate*seconds)
	for i := range samples {
		t := float64(i) / sampleRate
		var s float64
		for _, f := range freqs {
			s += math.Sin(2 * math.Pi * f * t)
		}
		samples[i] = s/float64(len(freqs))*0.8 + (rng.Float64()*2-1)*0.2
	}

	buf := &audio.AudioBuffer{Samples: samples, SampleRate: sampleRate}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := FromBuffer(buf); err != nil {
			b.Fatal(err)
		}
	}
}
