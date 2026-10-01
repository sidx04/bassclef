package audio

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func generateTestMP3(t *testing.T, path string) {
	t.Helper()

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found in PATH, skipping MP3 decode test")
	}

	cmd := exec.Command(ffmpeg,
		"-y",
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration=0.5",
		"-ar", "44100",
		"-ac", "1",
		path,
	)

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate test mp3: %v", err)
	}
}

func TestMP3DecoderDecode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.mp3")
	generateTestMP3(t, path)

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test mp3: %v", err)
	}
	defer file.Close()

	buf, err := NewMP3Decoder().Decode(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.SampleRate != 44100 {
		t.Errorf("got sample rate %d, want 44100", buf.SampleRate)
	}

	if len(buf.Samples) == 0 {
		t.Fatal("expected decoded samples, got none")
	}

	for i, s := range buf.Samples {
		if s < -1.0 || s > 1.0 {
			t.Fatalf("sample %d out of range [-1,1]: %v", i, s)
		}
	}
}
