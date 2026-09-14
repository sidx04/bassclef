package catalog

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
	"github.com/sidx04/bassclef/internal/fingerprint"
	"github.com/sidx04/bassclef/internal/storage"
)

func generateTestWAV(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create wav file: %v", err)
	}
	defer file.Close()

	const sampleRate = 44100
	const durationSeconds = 0.5

	enc := wav.NewEncoder(file, sampleRate, 16, 1, 1)

	numSamples := int(sampleRate * durationSeconds)
	data := make([]int, numSamples)
	for i := range data {
		data[i] = int(10000 * math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate)))
	}

	buf := &goaudio.IntBuffer{
		Format:         &goaudio.Format{NumChannels: 1, SampleRate: sampleRate},
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

// computeExpectedLandmarks re-derives landmarks with the same pipeline and
// config Ingest uses, so the test can confirm a real landmark was
// persisted without any package reaching around storage's public API.
func computeExpectedLandmarks(t *testing.T, wavPath string) []fingerprint.Landmark {
	t.Helper()

	buf, err := audio.LoadWAV(wavPath)
	if err != nil {
		t.Fatalf("failed to load wav: %v", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		t.Fatalf("failed to split frames: %v", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		t.Fatalf("failed to build window: %v", err)
	}

	spectrogram, err := dsp.Spectrogram(frames, window, dsp.NewFFT(cfg.FFTSize))
	if err != nil {
		t.Fatalf("failed to compute spectrogram: %v", err)
	}

	peaks, err := fingerprint.FindPeaks(spectrogram, fingerprint.DefaultPeakConfig())
	if err != nil {
		t.Fatalf("failed to find peaks: %v", err)
	}

	landmarks, err := fingerprint.GenerateLandmarks(peaks, fingerprint.DefaultLandmarkConfig())
	if err != nil {
		t.Fatalf("failed to generate landmarks: %v", err)
	}

	return landmarks
}

func TestIngest(t *testing.T) {
	dir := t.TempDir()

	wavPath := filepath.Join(dir, "test.wav")
	generateTestWAV(t, wavPath)

	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	if err := Ingest(store, wavPath, "Test Song", "Test Artist"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	song, err := store.GetSong(1)
	if err != nil {
		t.Fatalf("failed to get song: %v", err)
	}

	if song.Title != "Test Song" || song.Artist != "Test Artist" || song.Source != wavPath {
		t.Errorf("got song %#v, want Title=Test Song Artist=Test Artist Source=%s", song, wavPath)
	}

	landmarks := computeExpectedLandmarks(t, wavPath)
	if len(landmarks) == 0 {
		t.Fatal("test fixture produced no landmarks; fixture needs adjusting")
	}

	lm := landmarks[0]
	hash := fingerprint.Hash(uint32(lm.AnchorFrequency), uint32(lm.TargetFrequency), uint32(lm.DeltaTime))

	matches, err := store.LookupHash(hash)
	if err != nil {
		t.Fatalf("failed to look up hash: %v", err)
	}

	found := false
	for _, m := range matches {
		if m.SongID == 1 && m.TimeOffset == lm.AnchorTime {
			found = true
		}
	}
	if !found {
		t.Errorf("expected fingerprint for hash %d not found in store", hash)
	}
}

func TestIngestEmptyTitle(t *testing.T) {
	dir := t.TempDir()

	store, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	err = Ingest(store, "irrelevant.wav", "", "Test Artist")
	if err == nil {
		t.Fatal("expected an error")
	}
}
