package fingerprint

import (
	"math"
	"testing"

	"github.com/sidx04/bassclef/internal/audio"
)

func TestFromBuffer(t *testing.T) {
	const sampleRate = 44100
	numSamples := sampleRate / 2

	samples := make([]float64, numSamples)
	for i := range samples {
		// test for a simple sine wave signal
		samples[i] = math.Sin(2 * math.Pi * 440 * float64(i) / float64(sampleRate))
	}

	buf := &audio.AudioBuffer{Samples: samples, SampleRate: sampleRate}

	landmarks, err := FromBuffer(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(landmarks) == 0 {
		t.Fatal("expected at least one landmark")
	}
}

func TestFromBufferTooShort(t *testing.T) {
	buf := &audio.AudioBuffer{Samples: []float64{0, 0, 0}, SampleRate: 44100}

	_, err := FromBuffer(buf)
	if err == nil {
		t.Fatal("expected an error")
	}
}
