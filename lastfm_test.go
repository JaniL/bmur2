package bmur2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withLastfmServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestRecentTracksParsesArray(t *testing.T) {
	fixture := `{"recenttracks":{"track":[` +
		`{"artist":{"#text":"Artist A"},"name":"Song A","url":"https://example.com/a","date":{"uts":"1700000000"}},` +
		`{"artist":{"#text":"Artist B"},"name":"Song B","url":"https://example.com/b"}` +
		`]}}`
	server := withLastfmServer(t, http.StatusOK, fixture)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	tracks, err := client.RecentTracks(context.Background(), 2)
	if err != nil {
		t.Fatalf("RecentTracks() error = %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("tracks = %d, want 2", len(tracks))
	}
	if tracks[0].Artist != "Artist A" || tracks[0].Name != "Song A" || tracks[0].URL != "https://example.com/a" {
		t.Errorf("first track = %+v", tracks[0])
	}
	if tracks[0].DateUTS != 1700000000 {
		t.Errorf("DateUTS = %d, want 1700000000", tracks[0].DateUTS)
	}
	if tracks[0].NowPlaying {
		t.Error("NowPlaying = true, want false")
	}
	if tracks[1].DateUTS != 0 {
		t.Errorf("DateUTS = %d, want 0", tracks[1].DateUTS)
	}
}

func TestRecentTracksParsesSingleObject(t *testing.T) {
	fixture := `{"recenttracks":{"track":{"artist":{"#text":"Artist A"},"name":"Song A","url":"https://example.com/a","date":{"uts":"1700000000"}}}}`
	server := withLastfmServer(t, http.StatusOK, fixture)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	tracks, err := client.RecentTracks(context.Background(), 1)
	if err != nil {
		t.Fatalf("RecentTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("tracks = %d, want 1", len(tracks))
	}
	if tracks[0].Name != "Song A" {
		t.Errorf("Name = %q, want %q", tracks[0].Name, "Song A")
	}
}

func TestRecentTracksNowPlaying(t *testing.T) {
	fixture := `{"recenttracks":{"track":[{"artist":{"#text":"Artist A"},"name":"Song A","url":"https://example.com/a","@attr":{"now_playing":"true"}}]}}`
	server := withLastfmServer(t, http.StatusOK, fixture)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	tracks, err := client.RecentTracks(context.Background(), 1)
	if err != nil {
		t.Fatalf("RecentTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("tracks = %d, want 1", len(tracks))
	}
	if !tracks[0].NowPlaying {
		t.Error("NowPlaying = false, want true")
	}
	if tracks[0].DateUTS != 0 {
		t.Errorf("DateUTS = %d, want 0", tracks[0].DateUTS)
	}
}

func TestRecentTracksHTTPError(t *testing.T) {
	server := withLastfmServer(t, http.StatusInternalServerError, "boom")

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	if _, err := client.RecentTracks(context.Background(), 1); err == nil {
		t.Error("RecentTracks() error = nil, want error")
	}
}

func TestRecentTracksAPIError(t *testing.T) {
	fixture := `{"error":6,"message":"User not found"}`
	server := withLastfmServer(t, http.StatusOK, fixture)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	_, err := client.RecentTracks(context.Background(), 1)
	if err == nil {
		t.Fatal("RecentTracks() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "User not found") {
		t.Errorf("error = %q, want it to contain %q", got, "User not found")
	}
}

func TestLatestTrackEmpty(t *testing.T) {
	fixture := `{"recenttracks":{"track":[]}}`
	server := withLastfmServer(t, http.StatusOK, fixture)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	track, err := client.LatestTrack(context.Background())
	if err != nil {
		t.Fatalf("LatestTrack() error = %v", err)
	}
	if track != nil {
		t.Errorf("track = %+v, want nil", track)
	}
}

func TestRecentTracksSendsQueryParams(t *testing.T) {
	var gotQuery map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = map[string]string{
			"method":  r.URL.Query().Get("method"),
			"user":    r.URL.Query().Get("user"),
			"api_key": r.URL.Query().Get("api_key"),
			"limit":   r.URL.Query().Get("limit"),
			"format":  r.URL.Query().Get("format"),
		}
		_, _ = w.Write([]byte(`{"recenttracks":{"track":[]}}`))
	}))
	t.Cleanup(server.Close)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	if _, err := client.RecentTracks(context.Background(), 15); err != nil {
		t.Fatalf("RecentTracks() error = %v", err)
	}

	want := map[string]string{
		"method":  "user.getrecenttracks",
		"user":    "test-user",
		"api_key": "test-api-key",
		"limit":   "15",
		"format":  "json",
	}
	for key, value := range want {
		if gotQuery[key] != value {
			t.Errorf("query %s = %q, want %q", key, gotQuery[key], value)
		}
	}
}

func TestRecentTracksErrorPayloadWithHTTPError(t *testing.T) {
	fixture := `{"error":6,"message":"User not found"}`
	server := withLastfmServer(t, http.StatusBadRequest, fixture)

	original := lastfmBaseURL
	lastfmBaseURL = server.URL
	defer func() { lastfmBaseURL = original }()

	client := NewLastfmClient("test-api-key", "test-user")
	if _, err := client.RecentTracks(context.Background(), 1); err == nil {
		t.Error("RecentTracks() error = nil, want error")
	}
}
