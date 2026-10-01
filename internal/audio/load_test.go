package audio

import (
	"os"
	"path/filepath"
	"testing"

	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func generateLoadTestWAV(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create wav file: %v", err)
	}
	defer file.Close()

	enc := wav.NewEncoder(file, 44100, 16, 1, 1)

	data := make([]int, 1000)
	for i := range data {
		data[i] = i % 100
	}

	buf := &goaudio.IntBuffer{
		Format:         &goaudio.Format{NumChannels: 1, SampleRate: 44100},
		Data:           data,
		SourceBitDepth: 16,
	}

	if err := enc.Write(buf); err != nil {
		t.Fatalf("failed to write wav data: %v", err)
	}

	if err := enc.Close(); err != nil {
		t.Fatalf("failed to close wav encoder: %v", err)
	}
}

func TestLoadWAV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wav")
	generateLoadTestWAV(t, path)

	buf, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.SampleRate != 44100 {
		t.Errorf("got sample rate %d, want 44100", buf.SampleRate)
	}

	if len(buf.Samples) != 1000 {
		t.Errorf("got %d samples, want 1000", len(buf.Samples))
	}
}

func TestLoadMP3(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.mp3")
	generateTestMP3(t, path)

	buf, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.SampleRate != 44100 {
		t.Errorf("got sample rate %d, want 44100", buf.SampleRate)
	}

	if len(buf.Samples) == 0 {
		t.Fatal("expected decoded samples, got none")
	}
}

func TestLoadUnsupportedExtension(t *testing.T) {
	_, err := Load("song.flac")
	if err == nil {
		t.Fatal("expected an error")
	}
}
