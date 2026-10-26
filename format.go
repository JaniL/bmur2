package bmur2

import (
	"fmt"
	"html"
	"strings"
	"time"
)

var timeNow = time.Now

const reservationStartsLayout = "02.01.2006 15:04"

func FormatTrack(track Track) string {
	return trackLink(track)
}

func FormatUpcomingReservations(reservations []Reservation) string {
	lines := make([]string, len(reservations))
	for i, reservation := range reservations {
		lines[i] = "- " + reservationLine(reservation)
	}
	return strings.Join(lines, "\n")
}

func reservationLine(reservation Reservation) string {
	starts := reservation.Starts.Format(reservationStartsLayout)
	label := html.EscapeString(strings.TrimSpace(reservation.Name))
	link := "<a href=\"" + html.EscapeString(reservationLink(reservation)) + "\">" + label + "</a>"
	return starts + ": " + link + ", " + html.EscapeString(strings.TrimSpace(reservation.Association))
}

func FormatRecentTracks(tracks []Track) string {
	lines := make([]string, len(tracks))
	for i, track := range tracks {
		lines[i] = "- " + trackLink(track)
	}
	return strings.Join(lines, "\n")
}

func trackLink(track Track) string {
	label := html.EscapeString(strings.TrimSpace(track.Artist + " - " + track.Name))
	link := "<a href=\"" + html.EscapeString(track.URL) + "\">" + label + "</a>"
	if track.DateUTS > 0 {
		link += " (" + RelativeTime(track.DateUTS) + ")"
	}
	return link
}

func RelativeTime(uts int64) string {
	seconds := int64(timeNow().Sub(time.Unix(uts, 0)).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	switch {
	case seconds < 45:
		return "a few seconds ago"
	case seconds < 90:
		return "a minute ago"
	case seconds < 45*60:
		return countAgo(seconds/60, "minute")
	case seconds < 90*60:
		return "an hour ago"
	case seconds < 22*3600:
		return countAgo(seconds/3600, "hour")
	case seconds < 36*3600:
		return "a day ago"
	case seconds < 25*86400:
		return countAgo(seconds/86400, "day")
	case seconds < 45*86400:
		return "a month ago"
	case seconds < 345*86400:
		return countAgo(seconds/(30*86400), "month")
	case seconds < 548*86400:
		return "a year ago"
	default:
		return countAgo(seconds/(365*86400), "year")
	}
}

func countAgo(count int64, unit string) string {
	return fmt.Sprintf("%d %ss ago", count, unit)
}
