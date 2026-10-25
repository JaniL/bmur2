package bmur2

import (
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	original := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = original }()

	cases := []struct {
		name string
		uts  int64
		want string
	}{
		{"seconds", now.Unix() - 10, "a few seconds ago"},
		{"a minute", now.Unix() - 60, "a minute ago"},
		{"minutes", now.Unix() - 5*60, "5 minutes ago"},
		{"an hour", now.Unix() - 60*60, "an hour ago"},
		{"hours", now.Unix() - 3*3600, "3 hours ago"},
		{"a day", now.Unix() - 24*3600, "a day ago"},
		{"days", now.Unix() - 3*86400, "3 days ago"},
		{"a month", now.Unix() - 30*86400, "a month ago"},
		{"months", now.Unix() - 100*86400, "3 months ago"},
		{"a year", now.Unix() - 400*86400, "a year ago"},
		{"years", now.Unix() - 900*86400, "2 years ago"},
		{"future clamps to now", now.Unix() + 60, "a few seconds ago"},
	}

	for _, tc := range cases {
		if got := RelativeTime(tc.uts); got != tc.want {
			t.Errorf("%s: RelativeTime(%d) = %q, want %q", tc.name, tc.uts, got, tc.want)
		}
	}
}

func TestFormatTrackEscapesHTML(t *testing.T) {
	track := Track{Artist: "AC & DC", Name: "Back <In> Black", URL: "https://example.com/a?b=1&c=2"}
	got := FormatTrack(track)
	want := `<a href="https://example.com/a?b=1&amp;c=2">AC &amp; DC - Back &lt;In&gt; Black</a>`
	if got != want {
		t.Errorf("FormatTrack() = %q, want %q", got, want)
	}
}

func TestFormatTrackWithDate(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	original := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = original }()

	track := Track{Artist: "Artist A", Name: "Song A", URL: "https://example.com/a", DateUTS: now.Unix() - 3*3600}
	got := FormatTrack(track)
	want := `<a href="https://example.com/a">Artist A - Song A</a> (3 hours ago)`
	if got != want {
		t.Errorf("FormatTrack() = %q, want %q", got, want)
	}
}

func TestFormatRecentTracks(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	original := timeNow
	timeNow = func() time.Time { return now }
	defer func() { timeNow = original }()

	tracks := []Track{
		{Artist: "Artist A", Name: "Song A", URL: "https://example.com/a", DateUTS: now.Unix() - 3*3600},
		{Artist: "Artist B", Name: "Song B", URL: "https://example.com/b", DateUTS: now.Unix() - 2*86400},
	}
	got := FormatRecentTracks(tracks)
	want := "- <a href=\"https://example.com/a\">Artist A - Song A</a> (3 hours ago)\n" +
		"- <a href=\"https://example.com/b\">Artist B - Song B</a> (2 days ago)"
	if got != want {
		t.Errorf("FormatRecentTracks() = %q, want %q", got, want)
	}
}

func TestFormatTrackNowPlayingHasNoDateSuffix(t *testing.T) {
	track := Track{Artist: "Artist A", Name: "Song A", URL: "https://example.com/a", NowPlaying: true}
	got := FormatTrack(track)
	want := `<a href="https://example.com/a">Artist A - Song A</a>`
	if got != want {
		t.Errorf("FormatTrack() = %q, want %q", got, want)
	}
}
