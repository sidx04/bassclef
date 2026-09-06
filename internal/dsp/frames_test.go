package dsp

import (
	"reflect"
	"testing"
)

func TestSplitFrames(t *testing.T) {
	samples := []float64{
		0, 1, 2, 3, 4,
		5, 6, 7, 8, 9,
	}

	frames, err := SplitFrames(
		samples,
		4,
		2,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := [][]float64{
		{0, 1, 2, 3},
		{2, 3, 4, 5},
		{4, 5, 6, 7},
		{6, 7, 8, 9},
	}

	if !reflect.DeepEqual(frames, want) {
		t.Errorf(
			"frames mismatch:\n got: %#v\nwant: %#v",
			frames,
			want,
		)
	}
}

func TestSplitFramesInvalidConfig(t *testing.T) {
	samples := []float64{1, 2, 3, 4}

	tests := []struct {
		name      string
		frameSize int
		hopSize   int
	}{
		{
			name:      "zero frame size",
			frameSize: 0,
			hopSize:   2,
		},
		{
			name:      "negative frame size",
			frameSize: -1,
			hopSize:   2,
		},
		{
			name:      "zero hop size",
			frameSize: 4,
			hopSize:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SplitFrames(
				samples,
				tt.frameSize,
				tt.hopSize,
			)

			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
