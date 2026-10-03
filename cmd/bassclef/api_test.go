package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/storage"
)

func generateTestWAVBytes(t *testing.T) []byte {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.wav")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create wav file: %v", err)
	}

	enc := wav.NewEncoder(file, 44100, 16, 1, 1)

	data := make([]int, 44100/2)
	for i := range data {
		data[i] = int(10000 * math.Sin(2*math.Pi*440*float64(i)/44100))
	}

	ib := &goaudio.IntBuffer{
		Format:         &goaudio.Format{NumChannels: 1, SampleRate: 44100},
		Data:           data,
		SourceBitDepth: 16,
	}

	if err := enc.Write(ib); err != nil {
		t.Fatalf("failed to write wav data: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("failed to close wav encoder: %v", err)
	}
	file.Close()

	wavBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated wav: %v", err)
	}

	return wavBytes
}

func generateTestMP3Bytes(t *testing.T) []byte {
	t.Helper()

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found in PATH, skipping MP3 handler test")
	}

	path := filepath.Join(t.TempDir(), "test.mp3")

	cmd := exec.Command(ffmpeg,
		"-y", "-f", "lavfi", "-i", "sine=frequency=440:duration=0.5",
		"-ar", "44100", "-ac", "1", path,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate test mp3: %v", err)
	}

	mp3Bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated mp3: %v", err)
	}

	return mp3Bytes
}

func TestRecognizeWAVMatch(t *testing.T) {
	wavBytes := generateTestWAVBytes(t)

	setupBuf, err := audio.NewWAVDecoder().Decode(bytes.NewReader(wavBytes))
	if err != nil {
		t.Fatalf("failed to decode generated wav: %v", err)
	}

	store := selfMatchingStore(t, setupBuf, 1, storage.Song{ID: 1, Title: "Song A", Artist: "Artist A"})
	handler := newRecognizeHandler(store, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader(wavBytes))
	req.Header.Set("Content-Type", "audio/wav")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp recognizeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(resp.Matches))
	}

	if resp.Matches[0].Title != "Song A" {
		t.Errorf("got title %q, want %q", resp.Matches[0].Title, "Song A")
	}
}

func TestRecognizeMP3Match(t *testing.T) {
	mp3Bytes := generateTestMP3Bytes(t)

	setupBuf, err := audio.NewMP3Decoder().Decode(bytes.NewReader(mp3Bytes))
	if err != nil {
		t.Fatalf("failed to decode generated mp3: %v", err)
	}

	store := selfMatchingStore(t, setupBuf, 1, storage.Song{ID: 1, Title: "Song B", Artist: "Artist B"})
	handler := newRecognizeHandler(store, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader(mp3Bytes))
	req.Header.Set("Content-Type", "audio/mpeg")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp recognizeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(resp.Matches))
	}

	if resp.Matches[0].Title != "Song B" {
		t.Errorf("got title %q, want %q", resp.Matches[0].Title, "Song B")
	}
}

func TestRecognizeNoMatch(t *testing.T) {
	wavBytes := generateTestWAVBytes(t)

	handler := newRecognizeHandler(&fakeStore{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader(wavBytes))
	req.Header.Set("Content-Type", "audio/wav")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"matches":[]`)) {
		t.Errorf("got body %s, want literal empty array (not null) for matches", rec.Body.String())
	}
}

func TestRecognizeWrongMethod(t *testing.T) {
	handler := newRecognizeHandler(&fakeStore{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/recognize", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestRecognizeMissingContentType(t *testing.T) {
	handler := newRecognizeHandler(&fakeStore{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader([]byte("irrelevant")))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnsupportedMediaType)
	}
}

func TestRecognizeUnsupportedContentType(t *testing.T) {
	handler := newRecognizeHandler(&fakeStore{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader([]byte("irrelevant")))
	req.Header.Set("Content-Type", "audio/flac")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnsupportedMediaType)
	}
}

func TestRecognizeCorruptBody(t *testing.T) {
	handler := newRecognizeHandler(&fakeStore{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader([]byte("not a real wav file")))
	req.Header.Set("Content-Type", "audio/wav")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRecognizeStoreError(t *testing.T) {
	wavBytes := generateTestWAVBytes(t)

	store := &fakeStore{lookupErr: errors.New("db failure")}
	handler := newRecognizeHandler(store, nil)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader(wavBytes))
	req.Header.Set("Content-Type", "audio/wav")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRecognizeBodyTooLarge(t *testing.T) {
	handler := newRecognizeHandler(&fakeStore{}, nil)

	oversized := make([]byte, recognizeMaxBodyBytes+1)

	req := httptest.NewRequest(http.MethodPost, "/recognize", bytes.NewReader(oversized))
	req.Header.Set("Content-Type", "audio/wav")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
