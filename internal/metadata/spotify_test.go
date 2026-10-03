package metadata

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(t *testing.T, tokenHandler, searchHandler http.HandlerFunc) *SpotifyProvider {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/token", tokenHandler)
	mux.HandleFunc("/v1/search", searchHandler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return &SpotifyProvider{
		clientID:     "id",
		clientSecret: "secret",
		httpClient:   server.Client(),
		accountsURL:  server.URL,
		apiURL:       server.URL,
	}
}

func TestSpotifyProviderLookup(t *testing.T) {
	provider := newTestProvider(t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("token request method = %s, want POST", r.Method)
			}

			auth := r.Header.Get("Authorization")
			wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("id:secret"))
			if auth != wantAuth {
				t.Errorf("token request auth = %q, want %q", auth, wantAuth)
			}

			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		},
		func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Errorf("search request auth = %q, want %q", got, "Bearer test-token")
			}

			json.NewEncoder(w).Encode(map[string]any{
				"tracks": map[string]any{
					"items": []map[string]any{
						{
							"name":    "Song Title",
							"artists": []map[string]any{{"name": "Artist Name"}},
							"album":   map[string]any{"name": "Album Name"},
							"external_urls": map[string]any{
								"spotify": "https://open.spotify.com/track/abc123",
							},
						},
					},
				},
			})
		},
	)

	got, err := provider.Lookup("Song Title", "Artist Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Metadata{
		Title:  "Song Title",
		Artist: "Artist Name",
		Album:  "Album Name",
		URL:    "https://open.spotify.com/track/abc123",
	}

	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestSpotifyProviderLookupNotFound(t *testing.T) {
	provider := newTestProvider(t,
		func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		},
		func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{
				"tracks": map[string]any{
					"items": []map[string]any{},
				},
			})
		},
	)

	_, err := provider.Lookup("Unknown Song", "Unknown Artist")
	if err != ErrNotFound {
		t.Errorf("got error %v, want %v", err, ErrNotFound)
	}
}

func TestSpotifyProviderCachesToken(t *testing.T) {
	tokenRequests := 0

	provider := newTestProvider(t,
		func(w http.ResponseWriter, r *http.Request) {
			tokenRequests++
			json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		},
		func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{
				"tracks": map[string]any{
					"items": []map[string]any{
						{
							"name":          "A",
							"artists":       []map[string]any{{"name": "B"}},
							"album":         map[string]any{"name": "C"},
							"external_urls": map[string]any{"spotify": "url"},
						},
					},
				},
			})
		},
	)

	if _, err := provider.Lookup("A", "B"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := provider.Lookup("A", "B"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenRequests != 1 {
		t.Errorf("got %d token requests, want 1 (token should be cached)", tokenRequests)
	}
}

func TestSpotifyProviderTokenError(t *testing.T) {
	provider := newTestProvider(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Error("search should not be called when token fetch fails")
		},
	)

	_, err := provider.Lookup("A", "B")
	if err == nil {
		t.Fatal("expected an error")
	}
}
