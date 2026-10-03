package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/metadata"
	"github.com/sidx04/bassclef/internal/storage"
)

//go:embed webui
var webUI embed.FS

const (
	recognizeMaxBodyBytes = 50 << 20 // 50MB
	recognizeTopN         = 5
	recognizeMinVotes     = 1
)

type recognizeResponse struct {
	Matches []MatchResult `json:"matches"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

// newRecognizeHandler builds the POST /recognize handler. store and
// provider are injected (not hardcoded to a concrete type) so tests can
// substitute fakes instead of a real database or Spotify.
func newRecognizeHandler(store storage.Store, provider metadata.MetadataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var decoder audio.AudioDecoder
		switch r.Header.Get("Content-Type") {
		case "audio/wav":
			decoder = audio.NewWAVDecoder()
		case "audio/mpeg":
			decoder = audio.NewMP3Decoder()
		default:
			writeJSONError(w, http.StatusUnsupportedMediaType, "unsupported content type")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, recognizeMaxBodyBytes)

		buf, err := decoder.Decode(r.Body)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to decode audio: %v", err))
			return
		}

		results, err := recognizeCandidates(store, buf, recognizeTopN, recognizeMinVotes, provider)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to recognize audio")
			return
		}

		if results == nil {
			results = []MatchResult{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recognizeResponse{Matches: results})
	}
}

type ServeCmd struct {
	Port int    `name:"port" default:"8080" help:"Port to listen on."`
	DB   string `name:"db" default:"bassclef.db" help:"Path to the catalog database." type:"path"`
}

func (cmd *ServeCmd) Run() error {
	store, err := storage.Open(cmd.DB)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	webUIFS, err := fs.Sub(webUI, "webui")
	if err != nil {
		return fmt.Errorf("failed to load web UI: %w", err)
	}

	http.Handle("/", http.FileServer(http.FS(webUIFS)))
	http.HandleFunc("/recognize", newRecognizeHandler(store, newMetadataProvider()))

	addr := fmt.Sprintf(":%d", cmd.Port)
	fmt.Printf("Listening on %s\n", addr)

	return http.ListenAndServe(addr, nil)
}
