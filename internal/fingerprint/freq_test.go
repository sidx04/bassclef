package fingerprint

import "testing"

func TestBinToHz(t *testing.T) {
	got := BinToHz(1024, 44100, 4096)
	want := 11025.0

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBinToHzZero(t *testing.T) {
	got := BinToHz(0, 44100, 4096)
	want := 0.0

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
