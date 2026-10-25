package bmur2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var lastfmBaseURL = "https://ws.audioscrobbler.com/2.0/"

type Track struct {
	Artist     string
	Name       string
	URL        string
	DateUTS    int64
	NowPlaying bool
}

type LastfmClient struct {
	apiKey     string
	username   string
	httpClient *http.Client
}

func NewLastfmClient(apiKey, username string) *LastfmClient {
	return &LastfmClient{
		apiKey:     apiKey,
		username:   username,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *LastfmClient) RecentTracks(ctx context.Context, limit int) ([]Track, error) {
	endpoint, err := url.Parse(lastfmBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse lastfm base url: %w", err)
	}
	query := endpoint.Query()
	query.Set("method", "user.getrecenttracks")
	query.Set("user", c.username)
	query.Set("api_key", c.apiKey)
	query.Set("limit", strconv.Itoa(limit))
	query.Set("format", "json")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build lastfm request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call lastfm: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read lastfm response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lastfm returned status %d", resp.StatusCode)
	}

	var payload lastfmRecentTracksResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode lastfm response: %w", err)
	}
	if payload.Error != 0 {
		return nil, fmt.Errorf("lastfm error %d: %s", payload.Error, payload.Message)
	}

	tracks := make([]Track, 0, len(payload.RecentTracks.Tracks))
	for _, t := range payload.RecentTracks.Tracks {
		track := Track{
			Artist: t.Artist.Text,
			Name:   t.Name,
			URL:    t.URL,
		}
		if t.Date != nil {
			track.DateUTS = t.Date.UTS
		}
		track.NowPlaying = t.Attr != nil && t.Attr.NowPlaying == "true"
		tracks = append(tracks, track)
	}
	return tracks, nil
}

func (c *LastfmClient) LatestTrack(ctx context.Context) (*Track, error) {
	tracks, err := c.RecentTracks(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, nil
	}
	return &tracks[0], nil
}

type lastfmRecentTracksResponse struct {
	RecentTracks lastfmRecentTracks `json:"recenttracks"`
	Error        int                `json:"error"`
	Message      string             `json:"message"`
}

type lastfmRecentTracks struct {
	Tracks lastfmTrackList `json:"track"`
}

type lastfmTrackList []lastfmTrack

func (l *lastfmTrackList) UnmarshalJSON(data []byte) error {
	var tracks []lastfmTrack
	if err := json.Unmarshal(data, &tracks); err == nil {
		*l = tracks
		return nil
	}
	var single lastfmTrack
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*l = lastfmTrackList{single}
	return nil
}

type lastfmTrack struct {
	Artist struct {
		Text string `json:"#text"`
	} `json:"artist"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Date *struct {
		UTS int64 `json:"uts,string"`
	} `json:"date"`
	Attr *struct {
		NowPlaying string `json:"now_playing"`
	} `json:"@attr"`
}
