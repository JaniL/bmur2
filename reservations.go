package bmur2

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"

	"golang.org/x/net/html"
)

var ilotaloBaseURL = "https://ilotalo.matlu.fi"

const (
	upcomingWindowMonths    = 3
	maxUpcomingReservations = 15
)

type Reservation struct {
	ID          int
	Starts      time.Time
	Name        string
	Association string
	Closed      bool
}

type IlotaloClient struct {
	httpClient *http.Client
}

func NewIlotaloClient() *IlotaloClient {
	return &IlotaloClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{MaxVersion: tls.VersionTLS12},
			},
		},
	}
}

func (c *IlotaloClient) UpcomingReservations(ctx context.Context) ([]Reservation, error) {
	reservations, err := c.fetchReservations(ctx)
	if err != nil {
		return nil, err
	}

	now := timeNow()
	deadline := now.AddDate(0, upcomingWindowMonths, 0)

	upcoming := make([]Reservation, 0, len(reservations))
	for _, reservation := range reservations {
		if !reservation.Starts.Before(now) && !reservation.Starts.After(deadline) {
			upcoming = append(upcoming, reservation)
		}
	}

	sort.SliceStable(upcoming, func(i, j int) bool {
		return upcoming[i].Starts.Before(upcoming[j].Starts)
	})

	if len(upcoming) > maxUpcomingReservations {
		upcoming = upcoming[:maxUpcomingReservations]
	}
	return upcoming, nil
}

func (c *IlotaloClient) fetchReservations(ctx context.Context) ([]Reservation, error) {
	endpoint, err := url.Parse(ilotaloBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse ilotalo base url: %w", err)
	}
	endpoint.Path = "/index.php"
	query := endpoint.Query()
	query.Set("page", "reservation")
	query.Set("f", "3")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build ilotalo request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ilotalo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ilotalo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ilotalo returned status %d", resp.StatusCode)
	}

	return parseReservations(string(body)), nil
}

func reservationLink(reservation Reservation) string {
	return strings.TrimSuffix(ilotaloBaseURL, "/") + "/index.php?page=res&id=" + strconv.Itoa(reservation.ID)
}

var (
	helsinkiOnce sync.Once
	helsinkiLoc  *time.Location
)

func helsinkiLocation() *time.Location {
	helsinkiOnce.Do(func() {
		loc, err := time.LoadLocation("Europe/Helsinki")
		if err != nil {
			loc = time.UTC
		}
		helsinkiLoc = loc
	})
	return helsinkiLoc
}

var reservationDateLayouts = []string{"02.01.2006 15:04:05", "02.01.2006 15:04"}

func parseReservationDate(text string) (time.Time, bool) {
	dateText := strings.TrimSpace(text)
	for _, layout := range reservationDateLayouts {
		starts, err := time.ParseInLocation(layout, dateText, helsinkiLocation())
		if err == nil {
			return starts, true
		}
	}
	return time.Time{}, false
}

func reservationID(href string) (int, bool) {
	idx := strings.Index(href, "&id=")
	if idx == -1 {
		return 0, false
	}
	idText := href[idx+len("&id="):]
	if end := strings.IndexAny(idText, "&#"); end != -1 {
		idText = idText[:end]
	}
	id, err := strconv.Atoi(idText)
	if err != nil {
		return 0, false
	}
	return id, true
}

type reservationRow struct {
	href   string
	cells  []string
	cell   strings.Builder
	inCell bool
}

func (r *reservationRow) reservation() (Reservation, bool) {
	if len(r.cells) < 4 {
		return Reservation{}, false
	}
	id, ok := reservationID(r.href)
	if !ok {
		return Reservation{}, false
	}
	starts, ok := parseReservationDate(r.cells[0])
	if !ok {
		return Reservation{}, false
	}
	return Reservation{
		ID:          id,
		Starts:      starts,
		Name:        r.cells[1],
		Association: r.cells[2],
		Closed:      r.cells[3] == "suljettu",
	}, true
}

func parseReservations(pageHTML string) []Reservation {
	tokenizer := html.NewTokenizer(strings.NewReader(pageHTML))
	reservations := []Reservation{}
	var row reservationRow

	flushRow := func() {
		if reservation, ok := row.reservation(); ok {
			reservations = append(reservations, reservation)
		}
		row = reservationRow{}
	}

	for {
		switch tokenType := tokenizer.Next(); tokenType {
		case html.ErrorToken:
			flushRow()
			return reservations
		case html.StartTagToken:
			token := tokenizer.Token()
			switch token.Data {
			case "tr":
				flushRow()
			case "td":
				row.inCell = true
				row.cell.Reset()
			case "a":
				if row.inCell {
					for _, attr := range token.Attr {
						if attr.Key == "href" {
							row.href = attr.Val
						}
					}
				}
			}
		case html.TextToken:
			if row.inCell {
				row.cell.Write(tokenizer.Text())
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			switch token.Data {
			case "tr", "tbody", "table":
				flushRow()
			case "td":
				if row.inCell {
					row.cells = append(row.cells, strings.TrimSpace(row.cell.String()))
					row.cell.Reset()
					row.inCell = false
				}
			}
		}
	}
}
