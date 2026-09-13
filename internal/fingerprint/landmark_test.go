package fingerprint

import (
	"reflect"
	"testing"
)

func TestGenerateLandmarksSinglePair(t *testing.T) {
	peaks := []Peak{
		{Time: 0, Frequency: 100, Magnitude: 10},
		{Time: 5, Frequency: 200, Magnitude: 10},
	}

	cfg := LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 10, Fanout: 5}

	got, err := GenerateLandmarks(peaks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Landmark{
		{AnchorFrequency: 100, TargetFrequency: 200, DeltaTime: 5, AnchorTime: 0},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestGenerateLandmarksFanoutCap(t *testing.T) {
	peaks := []Peak{
		{Time: 0, Frequency: 100, Magnitude: 10},
		{Time: 1, Frequency: 101, Magnitude: 10},
		{Time: 2, Frequency: 102, Magnitude: 10},
		{Time: 3, Frequency: 103, Magnitude: 10},
	}

	cfg := LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 10, Fanout: 2}

	got, err := GenerateLandmarks(peaks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// anchor at Time 0 must only pair with the two nearest targets (Time 1, 2)
	wantFirstTwo := []Landmark{
		{AnchorFrequency: 100, TargetFrequency: 101, DeltaTime: 1, AnchorTime: 0},
		{AnchorFrequency: 100, TargetFrequency: 102, DeltaTime: 2, AnchorTime: 0},
	}

	if !reflect.DeepEqual(got[:2], wantFirstTwo) {
		t.Errorf("got %#v, want %#v", got[:2], wantFirstTwo)
	}
}

func TestGenerateLandmarksDeltaWindowBoundaries(t *testing.T) {
	peaks := []Peak{
		{Time: 0, Frequency: 100, Magnitude: 10},
		{Time: 1, Frequency: 101, Magnitude: 10}, // delta 1: below MinDeltaFrames(2), excluded
		{Time: 2, Frequency: 102, Magnitude: 10}, // delta 2: at MinDeltaFrames, included
		{Time: 4, Frequency: 104, Magnitude: 10}, // delta 4: at MaxDeltaFrames, included
		{Time: 5, Frequency: 105, Magnitude: 10}, // delta 5: above MaxDeltaFrames(4), excluded
	}

	cfg := LandmarkConfig{MinDeltaFrames: 2, MaxDeltaFrames: 4, Fanout: 10}

	got, err := GenerateLandmarks(peaks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Landmark{
		{AnchorFrequency: 100, TargetFrequency: 102, DeltaTime: 2, AnchorTime: 0},
		{AnchorFrequency: 100, TargetFrequency: 104, DeltaTime: 4, AnchorTime: 0},
		{AnchorFrequency: 101, TargetFrequency: 104, DeltaTime: 3, AnchorTime: 1},
		{AnchorFrequency: 101, TargetFrequency: 105, DeltaTime: 4, AnchorTime: 1},
		{AnchorFrequency: 102, TargetFrequency: 104, DeltaTime: 2, AnchorTime: 2},
		{AnchorFrequency: 102, TargetFrequency: 105, DeltaTime: 3, AnchorTime: 2},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestGenerateLandmarksUnsortedInput(t *testing.T) {
	peaks := []Peak{
		{Time: 5, Frequency: 200, Magnitude: 10},
		{Time: 0, Frequency: 100, Magnitude: 10},
	}

	cfg := LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 10, Fanout: 5}

	got, err := GenerateLandmarks(peaks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Landmark{
		{AnchorFrequency: 100, TargetFrequency: 200, DeltaTime: 5, AnchorTime: 0},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestGenerateLandmarksEmptyPeaks(t *testing.T) {
	cfg := LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 10, Fanout: 5}

	got, err := GenerateLandmarks(nil, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("got %#v, want nil", got)
	}
}

func TestGenerateLandmarksInvalidConfig(t *testing.T) {
	peaks := []Peak{{Time: 0, Frequency: 100, Magnitude: 10}}

	tests := []struct {
		name string
		cfg  LandmarkConfig
	}{
		{"negative min delta", LandmarkConfig{MinDeltaFrames: -1, MaxDeltaFrames: 10, Fanout: 5}},
		{"zero max delta", LandmarkConfig{MinDeltaFrames: 0, MaxDeltaFrames: 0, Fanout: 5}},
		{"max less than min", LandmarkConfig{MinDeltaFrames: 5, MaxDeltaFrames: 2, Fanout: 5}},
		{"zero fanout", LandmarkConfig{MinDeltaFrames: 0, MaxDeltaFrames: 10, Fanout: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateLandmarks(peaks, tt.cfg)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
