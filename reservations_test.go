package bmur2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func withIlotaloServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func ilotaloRow(dateText string, id int, name, association, openness string) string {
	class := "open"
	if openness == "suljettu" {
		class = "closed"
	}
	return "<tr class=" + class + ">\n" +
		"<td>" + dateText + "</td>\n" +
		"<td><a href=\"index.php?page=res&id=" + strconv.Itoa(id) + "\" class=\"eventlink\">" + name + "</a></td>\n" +
		"<td>" + association + "</td>\n" +
		"<td>" + openness + "</td>\n" +
		"\n"
}

func ilotaloPage(rows ...string) string {
	return "<!DOCTYPE html>\n<html lang=\"fi\">\n<body>\n<div class=\"keyTable\">\n<br>\n" +
		"<table class=\"keys\" border=\"0\" cellpadding=\"5\" cellspacing=\"5\">\n<tbody>\n" +
		"<tr>\n<th>Ajankohta</th>\n<th>Varaus</th>\n<th>Järjestö</th>\n<th>Avoimuus</th>\n</tr>\n" +
		strings.Join(rows, "") +
		"</tbody>\n</table>\n</div>\n</body>\n</html>\n"
}

func helsinkiDate(t time.Time, layout string) string {
	return t.In(helsinkiLocation()).Format(layout)
}

func freezeTimeNow(t *testing.T, now time.Time) {
	t.Helper()
	original := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = original })
}

func useIlotaloServer(t *testing.T, server *httptest.Server) {
	t.Helper()
	original := ilotaloBaseURL
	ilotaloBaseURL = server.URL
	t.Cleanup(func() { ilotaloBaseURL = original })
}

func TestUpcomingReservationsParsesAndFilters(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	page := ilotaloPage(
		ilotaloRow(helsinkiDate(now.Add(-24*time.Hour), "02.01.2006 15:04:05"), 1, "Past kokous", "Matlu", "suljettu"),
		ilotaloRow(helsinkiDate(now.Add(2*time.Hour), "02.01.2006 15:04:05"), 1669, "Matrix-leffailta", "Matrix", "suljettu"),
		ilotaloRow(helsinkiDate(now.AddDate(0, 2, 0), "02.01.2006 15:04"), 3, "TKO-&auml;ly sitsit", "TKO-äly", "avoin"),
		ilotaloRow(helsinkiDate(now.AddDate(0, 4, 0), "02.01.2006 15:04:05"), 4, "Too far ahead", "Matrix", "avoin"),
	)
	server := withIlotaloServer(t, http.StatusOK, page)
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 2 {
		t.Fatalf("reservations = %d, want 2", len(reservations))
	}

	first := reservations[0]
	if first.ID != 1669 {
		t.Errorf("first ID = %d, want %d", first.ID, 1669)
	}
	if !first.Starts.Equal(now.Add(2 * time.Hour)) {
		t.Errorf("first Starts = %v, want %v", first.Starts, now.Add(2*time.Hour))
	}
	if first.Name != "Matrix-leffailta" {
		t.Errorf("first Name = %q, want %q", first.Name, "Matrix-leffailta")
	}
	if first.Association != "Matrix" {
		t.Errorf("first Association = %q, want %q", first.Association, "Matrix")
	}
	if !first.Closed {
		t.Error("first Closed = false, want true")
	}

	second := reservations[1]
	if second.ID != 3 {
		t.Errorf("second ID = %d, want %d", second.ID, 3)
	}
	if !second.Starts.Equal(now.AddDate(0, 2, 0)) {
		t.Errorf("second Starts = %v, want %v", second.Starts, now.AddDate(0, 2, 0))
	}
	if second.Name != "TKO-äly sitsit" {
		t.Errorf("second Name = %q, want %q", second.Name, "TKO-äly sitsit")
	}
	if second.Closed {
		t.Error("second Closed = true, want false")
	}
}

func TestUpcomingReservationsParsesWellFormedRows(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	page := ilotaloPage(
		"<tr class=\"closed\">\n" +
			"<td>" + helsinkiDate(now.Add(2*time.Hour), "02.01.2006 15:04:05") + "</td>\n" +
			"<td><a href=\"index.php?page=res&id=1669\" class=\"eventlink\">Matrix-leffailta</a></td>\n" +
			"<td>Matrix</td>\n" +
			"<td>suljettu</td>\n" +
			"</tr>",
	)
	server := withIlotaloServer(t, http.StatusOK, page)
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 1 {
		t.Fatalf("reservations = %d, want 1", len(reservations))
	}
	if reservations[0].ID != 1669 {
		t.Errorf("ID = %d, want 1669", reservations[0].ID)
	}
	if !reservations[0].Closed {
		t.Error("Closed = false, want true")
	}
}

func TestUpcomingReservationsIncludesWindowEdge(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	page := ilotaloPage(
		ilotaloRow(helsinkiDate(now.AddDate(0, upcomingWindowMonths, 0), "02.01.2006 15:04:05"), 8, "Edge case", "Matlu", "avoin"),
	)
	server := withIlotaloServer(t, http.StatusOK, page)
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 1 {
		t.Fatalf("reservations = %d, want 1", len(reservations))
	}
}

func TestUpcomingReservationsSortsByStarts(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	later := now.AddDate(0, 2, 0)
	sooner := now.Add(2 * time.Hour)
	page := ilotaloPage(
		ilotaloRow(helsinkiDate(later, "02.01.2006 15:04:05"), 20, "Later", "Matlu", "avoin"),
		ilotaloRow(helsinkiDate(sooner, "02.01.2006 15:04:05"), 10, "Sooner", "Matlu", "avoin"),
	)
	server := withIlotaloServer(t, http.StatusOK, page)
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 2 {
		t.Fatalf("reservations = %d, want 2", len(reservations))
	}
	if reservations[0].ID != 10 || reservations[1].ID != 20 {
		t.Errorf("order = [%d, %d], want [10, 20]", reservations[0].ID, reservations[1].ID)
	}
}

func TestUpcomingReservationsCapsAtLimit(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	rows := make([]string, 0, 20)
	for i := 1; i <= 20; i++ {
		starts := now.Add(time.Duration(i) * 24 * time.Hour)
		rows = append(rows, ilotaloRow(helsinkiDate(starts, "02.01.2006 15:04"), i, "Event "+strconv.Itoa(i), "Matlu", "avoin"))
	}
	server := withIlotaloServer(t, http.StatusOK, ilotaloPage(rows...))
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != maxUpcomingReservations {
		t.Fatalf("reservations = %d, want %d", len(reservations), maxUpcomingReservations)
	}
	if reservations[0].ID != 1 {
		t.Errorf("first ID = %d, want 1", reservations[0].ID)
	}
	if reservations[len(reservations)-1].ID != maxUpcomingReservations {
		t.Errorf("last ID = %d, want %d", reservations[len(reservations)-1].ID, maxUpcomingReservations)
	}
}

func TestUpcomingReservationsSkipsInvalidRows(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	page := ilotaloPage(
		"<tr class=\"open\">\n<td>"+helsinkiDate(now.Add(2*time.Hour), "02.01.2006 15:04:05")+"</td>\n"+
			"<td><a href=\"index.php?page=res\" class=\"eventlink\">No id</a></td>\n"+
			"<td>Matlu</td>\n<td>avoin</td>\n</tr>",
		ilotaloRow("eggs in these difficult times", 5, "Bad date", "Matlu", "avoin"),
		ilotaloRow(helsinkiDate(now.Add(3*time.Hour), "02.01.2006 15:04:05"), 7, "Good", "Matlu", "avoin"),
	)
	server := withIlotaloServer(t, http.StatusOK, page)
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 1 {
		t.Fatalf("reservations = %d, want 1", len(reservations))
	}
	if reservations[0].ID != 7 {
		t.Errorf("ID = %d, want 7", reservations[0].ID)
	}
}

func TestUpcomingReservationsEmptyPage(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	server := withIlotaloServer(t, http.StatusOK, ilotaloPage())
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 0 {
		t.Fatalf("reservations = %d, want 0", len(reservations))
	}
}

func TestUpcomingReservationsHTTPError(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, helsinkiLocation())
	freezeTimeNow(t, now)

	server := withIlotaloServer(t, http.StatusInternalServerError, "boom")
	useIlotaloServer(t, server)

	if _, err := NewIlotaloClient().UpcomingReservations(context.Background()); err == nil {
		t.Error("UpcomingReservations() error = nil, want error")
	}
}

func TestUpcomingReservationsSendsExpectedRequest(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(ilotaloPage()))
	}))
	t.Cleanup(server.Close)
	useIlotaloServer(t, server)

	if _, err := NewIlotaloClient().UpcomingReservations(context.Background()); err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if gotPath != "/index.php" {
		t.Errorf("path = %q, want %q", gotPath, "/index.php")
	}
	if gotQuery.Get("page") != "reservation" {
		t.Errorf("page = %q, want %q", gotQuery.Get("page"), "reservation")
	}
	if gotQuery.Get("f") != "3" {
		t.Errorf("f = %q, want %q", gotQuery.Get("f"), "3")
	}
}

func TestFormatUpcomingReservations(t *testing.T) {
	starts := time.Date(2026, 6, 2, 18, 0, 0, 0, helsinkiLocation())
	reservations := []Reservation{
		{ID: 42, Starts: starts, Name: "Guild <night>", Association: "TKO-äly"},
		{ID: 43, Starts: starts.Add(2 * time.Hour), Name: "  Hallituksen kokous  ", Association: "Matlu"},
	}
	want := "- 02.06.2026 18:00: <a href=\"" + ilotaloBaseURL + "/index.php?page=res&amp;id=42\">Guild &lt;night&gt;</a>, TKO-äly\n" +
		"- 02.06.2026 20:00: <a href=\"" + ilotaloBaseURL + "/index.php?page=res&amp;id=43\">Hallituksen kokous</a>, Matlu"
	if got := FormatUpcomingReservations(reservations); got != want {
		t.Errorf("FormatUpcomingReservations() = %q, want %q", got, want)
	}
}

func TestParseReservationDateRejectsEmpty(t *testing.T) {
	if _, ok := parseReservationDate(""); ok {
		t.Error("parseReservationDate(\"\") = ok, want not ok")
	}
	if _, ok := parseReservationDate("   "); ok {
		t.Error("parseReservationDate(whitespace) = ok, want not ok")
	}
}

func TestReservationIDFromFragment(t *testing.T) {
	id, ok := reservationID("index.php?page=res&id=99#anchor")
	if !ok {
		t.Fatal("reservationID() ok = false, want true")
	}
	if id != 99 {
		t.Errorf("id = %d, want 99", id)
	}
}

func TestHelsinkiLocationLoads(t *testing.T) {
	if got := helsinkiLocation().String(); got != "Europe/Helsinki" {
		t.Errorf("helsinkiLocation() = %q, want %q", got, "Europe/Helsinki")
	}
}

func TestUpcomingReservationsComparesAcrossTimezones(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	freezeTimeNow(t, now)

	starts := time.Date(2026, 10, 26, 14, 0, 0, 0, helsinkiLocation())
	page := ilotaloPage(ilotaloRow(helsinkiDate(starts, "02.01.2006 15:04:05"), 6, "TZ check", "Matlu", "avoin"))
	server := withIlotaloServer(t, http.StatusOK, page)
	useIlotaloServer(t, server)

	reservations, err := NewIlotaloClient().UpcomingReservations(context.Background())
	if err != nil {
		t.Fatalf("UpcomingReservations() error = %v", err)
	}
	if len(reservations) != 1 {
		t.Fatalf("reservations = %d, want 1", len(reservations))
	}
	if !reservations[0].Starts.Equal(starts) {
		t.Errorf("Starts = %v, want %v", reservations[0].Starts, starts)
	}
}
