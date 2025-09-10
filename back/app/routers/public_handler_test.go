package routers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

// helper para construir handler con sqlmock
func newHandlerWithMockDB(t *testing.T) (*PublicHandler, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return NewPublicHandler(db), mock, db
}

func TestPublic_ListVideos_OK(t *testing.T) {
	h, mock, db := newHandlerWithMockDB(t)
	defer db.Close()

	// Espera query con LIMIT $1 OFFSET $2
	rows := sqlmock.NewRows([]string{
		"id", "title", "processed_url", "thumb_url", "votes", "first_name", "last_name", "city",
	}).AddRow(10, "Video A", "http://x/10.mp4", "http://x/10.jpg", 7, "Ana", "Gomez", "Bogotá").
		AddRow(9, "Video B", "http://x/9.mp4", "http://x/9.jpg", 5, "Luis", "Ruiz", "Medellín")

	mock.ExpectQuery(`SELECT v\.id, v\.title, v\.processed_url, v\.thumb_url, v\.votes, u\.first_name, u\.last_name, u\.city`).
		WithArgs(2, 1). // limit=2, offset=1
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/public/videos?limit=2&offset=1", nil)
	rr := httptest.NewRecorder()
	h.ListVideos(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rr.Code)
	}
	body := rr.Body.String()
	if want := `"video_id":10`; !contains(body, want) {
		t.Errorf("response missing %s; got %s", want, body)
	}
	if want := `"author":"Ana Gomez"`; !contains(body, want) {
		t.Errorf("response missing %s; got %s", want, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestPublic_ListVideos_DBError(t *testing.T) {
	h, mock, db := newHandlerWithMockDB(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT v\.id, v\.title, v\.processed_url`).
		WithArgs(20, 0). // defaults (limit=20, offset=0)
		WillReturnError(assertErr("boom"))

	req := httptest.NewRequest(http.MethodGet, "/api/public/videos", nil)
	rr := httptest.NewRecorder()
	h.ListVideos(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d; want 500", rr.Code)
	}
	_ = mock.ExpectationsWereMet()
}

func TestPublic_Rankings_All_OK(t *testing.T) {
	h, mock, db := newHandlerWithMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"first_name", "last_name", "city", "total"}).
		AddRow("Ana", "Gomez", "Bogotá", 12).
		AddRow("Luis", "Ruiz", "Medellín", 9)

	mock.ExpectQuery(`SELECT u\.first_name, u\.last_name, u\.city, SUM\(v\.votes\) as total`).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/public/rankings", nil)
	rr := httptest.NewRecorder()
	h.Rankings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rr.Code)
	}
	if body := rr.Body.String(); !contains(body, `"position":1`) || !contains(body, `"username":"Ana Gomez"`) {
		t.Errorf("unexpected body: %s", body)
	}
	_ = mock.ExpectationsWereMet()
}

func TestPublic_Rankings_ByCity_OK(t *testing.T) {
	h, mock, db := newHandlerWithMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"first_name", "last_name", "city", "total"}).
		AddRow("Ana", "Gomez", "Bogotá", 12)

	mock.ExpectQuery(`SELECT u\.first_name, u\.last_name, u\.city, SUM\(v\.votes\) as total`).
		WithArgs("Bogotá").
		WillReturnRows(rows)

	u := url.URL{Path: "/api/public/rankings"}
	q := u.Query()
	q.Set("city", "Bogotá")
	u.RawQuery = q.Encode()

	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	rr := httptest.NewRecorder()
	h.Rankings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rr.Code)
	}
	if body := rr.Body.String(); !contains(body, `"city":"Bogotá"`) {
		t.Errorf("unexpected body: %s", body)
	}
	_ = mock.ExpectationsWereMet()
}

func TestPublic_MyVotes_Unauthorized(t *testing.T) {
	h, _, db := newHandlerWithMockDB(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/public/my-votes", nil)
	rr := httptest.NewRecorder()
	h.MyVotes(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rr.Code)
	}
}

func TestPublic_Vote_Unauthorized(t *testing.T) {
	h, _, db := newHandlerWithMockDB(t)
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/public/videos/{id}/vote", h.Vote).Methods(http.MethodPost)

	req := httptest.NewRequest(http.MethodPost, "/api/public/videos/5/vote", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rr.Code)
	}
}

func TestPublic_Unvote_Unauthorized(t *testing.T) {
	h, _, db := newHandlerWithMockDB(t)
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/public/videos/{id}/vote", h.Unvote).Methods(http.MethodDelete)

	req := httptest.NewRequest(http.MethodDelete, "/api/public/videos/5/vote", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rr.Code)
	}
}

// ---------------- helpers ----------------

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && indexOf(s, sub) >= 0))
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// error helper
type assertErr string

func (e assertErr) Error() string { return string(e) }
