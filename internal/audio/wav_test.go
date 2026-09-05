package audio

import (
	"math"
	"testing"
)

func TestStereoToMono(t *testing.T) {
	tests := []struct {
		name     string
		samples  []int
		channels int
		want     []float64
	}{
		{
			name:     "mono",
			samples:  []int{100, 200, 300},
			channels: 1,
			want:     []float64{100, 200, 300},
		},
		{
			name:     "stereo",
			samples:  []int{100, 200, 300, 400},
			channels: 2,
			want:     []float64{150, 350},
		},
		{
			name:     "three channels",
			samples:  []int{3, 6, 9, 12, 15, 18},
			channels: 3,
			want:     []float64{6, 15},
		},
		{
			name:     "zero channels",
			samples:  []int{100, 200},
			channels: 0,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, _ := stereoToMono(tt.samples, tt.channels)

			if len(expected) != len(tt.want) {
				t.Fatalf(
					"length mismatch: expected %d, want %d",
					len(expected),
					len(tt.want),
				)
			}

			for i := range expected {
				if expected[i] != tt.want[i] {
					t.Errorf(
						"sample %d: expected %v, want %v",
						i,
						expected[i],
						tt.want[i],
					)
				}
			}
		})
	}
}

func TestNormalizeSamples16Bit(t *testing.T) {
	samples := []float64{
		-32768,
		-16384,
		0,
		16384,
		32767,
	}

	expected := normaliseSamples(samples, 16)

	want := []float64{
		-1.0,
		-0.5,
		0.0,
		0.5,
		float64(32767) / 32768,
	}

	const epsilon = 1e-9

	for i := range expected {
		if math.Abs(expected[i]-want[i]) > epsilon {
			t.Errorf(
				"sample %d: expected %f, want %f",
				i,
				expected[i],
				want[i],
			)
		}
	}
}

func TestNormalizeSamples(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		bitDepth int
		want     []float64
	}{
		{
			name:     "8 bit",
			samples:  []float64{-128, -64, 0, 64, 127},
			bitDepth: 8,
			want: []float64{
				-1.0,
				-0.5,
				0.0,
				0.5,
				127.0 / 128.0,
			},
		},
		{
			name:     "16 bit",
			samples:  []float64{-32768, 0, 32767},
			bitDepth: 16,
			want: []float64{
				-1.0,
				0.0,
				32767.0 / 32768.0,
			},
		},
		{
			name:     "invalid bit depth",
			samples:  []float64{-100, 0, 100},
			bitDepth: 0,
			want:     []float64{-100, 0, 100},
		},
	}

	const epsilon = 1e-9

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := normaliseSamples(
				append([]float64(nil), tt.samples...),
				tt.bitDepth,
			)

			for i := range expected {
				if math.Abs(expected[i]-tt.want[i]) > epsilon {
					t.Errorf(
						"sample %d: expected %f, want %f",
						i,
						expected[i],
						tt.want[i],
					)
				}
			}
		})
	}
}
