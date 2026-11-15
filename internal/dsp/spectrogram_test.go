package dsp

import (
	"math"
	"testing"
)

// stubFFT returns its input as the spectrum, one bin per sample. It follows
// the FFT contract: it writes into dst when dst is non-nil.
type stubFFT struct{}

func (stubFFT) Forward(dst []complex128, input []float64) []complex128 {
	if dst == nil {
		dst = make([]complex128, len(input))
	}
	for i, v := range input {
		dst[i] = complex(v, 0)
	}
	return dst
}

// old Spectrogram test; unoptimised
func TestSpectrogram(t *testing.T) {
	frames := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	window := []float64{1, 1, 1}

	got, err := Spectrogram(frames, window, stubFFT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	const epsilon = 1e-9

	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > epsilon {
				t.Errorf("frame %d bin %d: got %v, want %v", i, j, got[i][j], want[i][j])
			}
		}
	}
}

// Rows share one backing array, so each row's capacity must be capped at
// its length: an append on one row must not overwrite the next row.
func TestSpectrogramRowsCapped(t *testing.T) {
	frames := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	got, err := Spectrogram(frames, []float64{1, 1, 1}, stubFFT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, row := range got {
		if cap(row) != len(row) {
			t.Errorf("row %d: cap %d, want %d", i, cap(row), len(row))
		}
	}

	_ = append(got[0], 99)
	if got[1][0] != 4 {
		t.Errorf("append on row 0 overwrote row 1: got %v, want 4", got[1][0])
	}
}

func TestSpectrogramWindowMismatch(t *testing.T) {
	frames := [][]float64{
		{1, 2, 3},
	}

	window := []float64{1, 1}

	_, err := Spectrogram(frames, window, stubFFT{})
	if err == nil {
		t.Fatal("expected an error")
	}
}

// A later frame with the wrong length must be rejected, not silently
// truncated into the reused window buffer.
func TestSpectrogramLaterFrameMismatch(t *testing.T) {
	frames := [][]float64{
		{1, 2, 3},
		{4, 5},
	}

	_, err := Spectrogram(frames, []float64{1, 1, 1}, stubFFT{})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestSpectrogramEmpty(t *testing.T) {
	got, err := Spectrogram(nil, []float64{1, 1, 1}, stubFFT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d rows, want 0", len(got))
	}
}

func TestMagnitudeReusesDst(t *testing.T) {
	spectrum := []complex128{3 + 4i, 0, -5}
	dst := make([]float64, len(spectrum))

	got := Magnitude(dst, spectrum)

	if &got[0] != &dst[0] {
		t.Error("Magnitude allocated instead of writing into dst")
	}

	want := []float64{5, 0, 5}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("bin %d: got %v, want %v", i, got[i], want[i])
		}
	}
}
