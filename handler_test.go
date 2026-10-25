package bmur2

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func resetRuntime(t *testing.T) {
	t.Helper()
	botOnce = sync.Once{}
	bot = nil
	botErr = nil
	runtimeConfig = nil
	t.Setenv("BMUR_TOKEN", "TEST:TOKEN")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "test-secret")
	t.Setenv("BMUR_LASTFM_APIKEY", "test-api-key")
	t.Setenv("BMUR_LASTFM_USERNAME", "test-user")
}

type fakeTelegram struct {
	mu       sync.Mutex
	requests []map[string]string
}

func (f *fakeTelegram) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fields := make(map[string]string)
	for name, values := range r.MultipartForm.Value {
		if len(values) > 0 {
			fields[name] = values[0]
		}
	}
	f.mu.Lock()
	f.requests = append(f.requests, fields)
	f.mu.Unlock()
	_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1,"date":0,"chat":{"id":123,"type":"private"}}}`))
}

func (f *fakeTelegram) sent() []map[string]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]map[string]string, len(f.requests))
	copy(out, f.requests)
	return out
}

func telegramUpdate(text string) string {
	return `{"update_id":1,"message":{"message_id":10,"from":{"id":42,"is_bot":false,"first_name":"Test"},"chat":{"id":123,"type":"private"},"date":1700000000,"text":"` + text + `"}}`
}

func postUpdate(t *testing.T, text, secret string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(telegramUpdate(text)))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secret)
	rec := httptest.NewRecorder()
	Handle(rec, req)
	return rec
}

func TestHandleMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	Handle(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleRejectsMissingSecret(t *testing.T) {
	resetRuntime(t)
	rec := postUpdate(t, "/np", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if bot != nil {
		t.Error("bot must not be initialized for unauthorized requests")
	}
}

func TestHandleRejectsWrongSecret(t *testing.T) {
	resetRuntime(t)
	rec := postUpdate(t, "/np", "wrong-secret")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if bot != nil {
		t.Error("bot must not be initialized for unauthorized requests")
	}
}

func TestHandleInvalidUpdatePayload(t *testing.T) {
	resetRuntime(t)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-json"))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "test-secret")
	rec := httptest.NewRecorder()
	Handle(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if bot != nil {
		t.Error("bot must not be initialized for invalid payloads")
	}
}

func TestHandleMissingConfiguration(t *testing.T) {
	resetRuntime(t)
	t.Setenv("BMUR_TOKEN", "")
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(telegramUpdate("/np")))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "test-secret")
	rec := httptest.NewRecorder()
	Handle(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleNowPlaying(t *testing.T) {
	resetRuntime(t)

	lastfmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"recenttracks":{"track":[{"artist":{"#text":"Artist A"},"name":"Song A","url":"https://example.com/a","@attr":{"now_playing":"true"}}]}}`))
	}))
	t.Cleanup(lastfmSrv.Close)
	lastfmBaseURL = lastfmSrv.URL

	telegram := &fakeTelegram{}
	tgSrv := httptest.NewServer(telegram)
	t.Cleanup(tgSrv.Close)
	apiServerURL = tgSrv.URL

	rec := postUpdate(t, "/np", "test-secret")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	sent := telegram.sent()
	if len(sent) != 1 {
		t.Fatalf("telegram requests = %d, want 1", len(sent))
	}
	msg := sent[0]
	if msg["chat_id"] != "123" {
		t.Errorf("chat_id = %q, want %q", msg["chat_id"], "123")
	}
	want := `<a href="https://example.com/a">Artist A - Song A</a>`
	if msg["text"] != want {
		t.Errorf("text = %q, want %q", msg["text"], want)
	}
	if msg["parse_mode"] != "HTML" {
		t.Errorf("parse_mode = %q, want %q", msg["parse_mode"], "HTML")
	}
	if !strings.Contains(msg["link_preview_options"], "is_disabled") {
		t.Errorf("link_preview_options = %q, want it to contain is_disabled", msg["link_preview_options"])
	}
}

func TestHandleNowPlayingWithMentionAndArgs(t *testing.T) {
	resetRuntime(t)

	lastfmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"recenttracks":{"track":[{"artist":{"#text":"Artist A"},"name":"Song A","url":"https://example.com/a","@attr":{"now_playing":"true"}}]}}`))
	}))
	t.Cleanup(lastfmSrv.Close)
	lastfmBaseURL = lastfmSrv.URL

	telegram := &fakeTelegram{}
	tgSrv := httptest.NewServer(telegram)
	t.Cleanup(tgSrv.Close)
	apiServerURL = tgSrv.URL

	rec := postUpdate(t, "/np@bmur_bot what's playing", "test-secret")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(telegram.sent()) != 1 {
		t.Fatalf("telegram requests = %d, want 1", len(telegram.sent()))
	}
}

func TestHandleRecentTracks(t *testing.T) {
	resetRuntime(t)

	now := time.Unix(1_700_000_000, 0)
	original := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = original }()

	fixture := `{"recenttracks":{"track":[` +
		`{"artist":{"#text":"Artist A"},"name":"Song A","url":"https://example.com/a","date":{"uts":"` + strconv.FormatInt(now.Unix()-3*3600, 10) + `"}},` +
		`{"artist":{"#text":"Artist B"},"name":"Song B","url":"https://example.com/b","date":{"uts":"` + strconv.FormatInt(now.Unix()-2*86400, 10) + `"}}` +
		`]}}`
	lastfmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fixture))
	}))
	t.Cleanup(lastfmSrv.Close)
	lastfmBaseURL = lastfmSrv.URL

	telegram := &fakeTelegram{}
	tgSrv := httptest.NewServer(telegram)
	t.Cleanup(tgSrv.Close)
	apiServerURL = tgSrv.URL

	rec := postUpdate(t, "/recent", "test-secret")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	sent := telegram.sent()
	if len(sent) != 1 {
		t.Fatalf("telegram requests = %d, want 1", len(sent))
	}
	want := "- <a href=\"https://example.com/a\">Artist A - Song A</a> (3 hours ago)\n" +
		"- <a href=\"https://example.com/b\">Artist B - Song B</a> (2 days ago)"
	if sent[0]["text"] != want {
		t.Errorf("text = %q, want %q", sent[0]["text"], want)
	}
	if sent[0]["parse_mode"] != "HTML" {
		t.Errorf("parse_mode = %q, want %q", sent[0]["parse_mode"], "HTML")
	}
}

func TestHandleLastfmFailureSendsFallback(t *testing.T) {
	resetRuntime(t)

	lastfmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(lastfmSrv.Close)
	lastfmBaseURL = lastfmSrv.URL

	telegram := &fakeTelegram{}
	tgSrv := httptest.NewServer(telegram)
	t.Cleanup(tgSrv.Close)
	apiServerURL = tgSrv.URL

	rec := postUpdate(t, "/np", "test-secret")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	sent := telegram.sent()
	if len(sent) != 1 {
		t.Fatalf("telegram requests = %d, want 1", len(sent))
	}
	if sent[0]["text"] != "Could not fetch current track right now." {
		t.Errorf("text = %q, want fallback message", sent[0]["text"])
	}
	if _, ok := sent[0]["parse_mode"]; ok {
		t.Errorf("parse_mode = %q, want it absent for plain messages", sent[0]["parse_mode"])
	}
}

func TestHandleUnknownCommandIsIgnored(t *testing.T) {
	resetRuntime(t)

	telegram := &fakeTelegram{}
	tgSrv := httptest.NewServer(telegram)
	t.Cleanup(tgSrv.Close)
	apiServerURL = tgSrv.URL

	rec := postUpdate(t, "/unknown", "test-secret")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(telegram.sent()) != 0 {
		t.Errorf("telegram requests = %d, want 0", len(telegram.sent()))
	}
}
