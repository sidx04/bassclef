package metadata

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// SpotifyProvider is a MetadataProvider backed by the Spotify Web API,
// authenticated via the Client Credentials flow (app-only auth — no user
// login, since lookups never touch user-specific data).
type SpotifyProvider struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	accountsURL  string
	apiURL       string

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

// NewSpotifyProvider reads SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET
// (loading a .env file first if one is present; a missing file is not an
// error, since the vars may already be exported) and returns a provider
// ready to query the real Spotify API.
func NewSpotifyProvider() (*SpotifyProvider, error) {
	_ = godotenv.Load()

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	defaultAccountsURL := os.Getenv("SPOTIFY_ACCOUNTS_URL")
	defaultAPIURL := os.Getenv("SPOTIFY_API_URL")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("Spotify credentials not found")
	}

	return &SpotifyProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		accountsURL:  defaultAccountsURL,
		apiURL:       defaultAPIURL,
	}, nil
}

// Implements the generic MetadataProvider trait.
// Lookup searches Spotify for title or artist and returns the first result.
func (p *SpotifyProvider) Lookup(title, artist string) (Metadata, error) {
	token, err := p.token()
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to get access token: %w", err)
	}

	query := fmt.Sprintf("track:%s artist:%s", title, artist)

	reqURL := p.apiURL + "/v1/search?" + url.Values{
		"q":     {query},
		"type":  {"track"},
		"limit": {"1"},
	}.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to build search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("search request failed: status %d", resp.StatusCode)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return Metadata{}, fmt.Errorf("failed to decode search response: %w", err)
	}

	if len(sr.Tracks.Items) == 0 {
		return Metadata{}, ErrNotFound
	}

	item := sr.Tracks.Items[0]

	result := Metadata{
		Title: item.Name,
		Album: item.Album.Name,
		URL:   item.ExternalURLs.Spotify,
	}

	if len(item.Artists) > 0 {
		result.Artist = item.Artists[0].Name
	}

	return result, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (p *SpotifyProvider) token() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.expiresAt) {
		return p.accessToken, nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}

	req, err := http.NewRequest(http.MethodPost, p.accountsURL+"/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to build token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(p.clientID + ":" + p.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: status %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	p.accessToken = tr.AccessToken
	// Refresh a little early so a token never expires mid-request.
	p.expiresAt = time.Now().Add(time.Duration(tr.ExpiresIn)*time.Second - 30*time.Second)

	return p.accessToken, nil
}

type searchResponse struct {
	Tracks struct {
		Items []trackItem `json:"items"`
	} `json:"tracks"`
}

type trackItem struct {
	Name         string       `json:"name"`
	Artists      []artist     `json:"artists"`
	Album        album        `json:"album"`
	ExternalURLs externalURLs `json:"external_urls"`
}

type artist struct {
	Name string `json:"name"`
}

type album struct {
	Name string `json:"name"`
}

type externalURLs struct {
	Spotify string `json:"spotify"`
}
