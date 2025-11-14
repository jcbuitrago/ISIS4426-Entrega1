package routers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"ISIS4426-Entrega1/app/middleware"

	"github.com/gorilla/mux"
)

const DBerror = "db error"
const HeaderClass = "Content-Type"
const HeaderJSON = "application/json"
const TXerror = "tx error"

type PublicHandler struct{ DB *sql.DB }

func NewPublicHandler(db *sql.DB) *PublicHandler { return &PublicHandler{DB: db} }

// GET /api/public/my-votes (JWT)
func (h *PublicHandler) MyVotes(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := h.DB.Query(`SELECT video_id FROM votes WHERE user_id=$1`, uid)
	if err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			http.Error(w, DBerror, http.StatusInternalServerError)
			return
		}
		ids = append(ids, id)
	}
	type resp struct {
		VideoIDs  []int `json:"video_ids"`
		Used      int   `json:"used"`
		Remaining int   `json:"remaining"`
	}
	used := len(ids)
	if used < 0 {
		used = 0
	}
	rem := 2 - used
	if rem < 0 {
		rem = 0
	}
	w.Header().Set(HeaderClass, HeaderJSON)
	json.NewEncoder(w).Encode(resp{VideoIDs: ids, Used: used, Remaining: rem})
}

// GET /api/public/videos
func (h *PublicHandler) ListVideos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	const qsql = `
	SELECT v.id, v.title, v.processed_url, v.thumb_url, v.votes, u.first_name, u.last_name, u.city
	FROM videos v
	JOIN users u ON u.id = v.user_id
	WHERE v.status = 'processed' AND v.processed_url IS NOT NULL
	ORDER BY v.votes DESC, v.id DESC
	LIMIT $1 OFFSET $2`
	rows, err := h.DB.Query(qsql, limit, offset)
	if err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type item struct {
		VideoID      int    `json:"video_id"`
		Title        string `json:"title"`
		ProcessedURL string `json:"processed_url"`
		ThumbURL     string `json:"thumb_url"`
		Votes        int    `json:"votes"`
		Author       string `json:"author"`
		City         string `json:"city"`
	}
	var out []item
	for rows.Next() {
		var it item
		var fn, ln string
		if err := rows.Scan(&it.VideoID, &it.Title, &it.ProcessedURL, &it.ThumbURL, &it.Votes, &fn, &ln, &it.City); err != nil {
			http.Error(w, DBerror, http.StatusInternalServerError)
			return
		}
		it.Author = fn + " " + ln
		out = append(out, it)
	}
	w.Header().Set(HeaderClass, HeaderJSON)
	json.NewEncoder(w).Encode(out)
}

// POST /api/public/videos/{id}/vote (JWT)
func (h *PublicHandler) Vote(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := mux.Vars(r)["id"]
	vid, err := strconv.Atoi(idStr)
	if err != nil || vid <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, TXerror, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var totalUserVotes int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM votes WHERE user_id=$1`, uid).Scan(&totalUserVotes); err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	if totalUserVotes >= 2 {
		http.Error(w, "Has alcanzado el límite de 2 votos.", http.StatusBadRequest)
		return
	}

	if _, err := tx.Exec(`INSERT INTO votes(video_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, vid, uid); err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	var cnt int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM votes WHERE video_id=$1 AND user_id=$2`, vid, uid).Scan(&cnt); err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	if cnt == 0 {
		http.Error(w, "Ya has votado por este video.", http.StatusBadRequest)
		return
	}

	if _, err := tx.Exec(`UPDATE videos SET votes = votes + 1 WHERE id=$1`, vid); err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, TXerror, http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderClass, HeaderJSON)
	json.NewEncoder(w).Encode(map[string]string{"message": "Voto registrado exitosamente."})
}

// DELETE /api/public/videos/{id}/vote (JWT)
func (h *PublicHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := mux.Vars(r)["id"]
	vid, err := strconv.Atoi(idStr)
	if err != nil || vid <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, TXerror, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(`DELETE FROM votes WHERE video_id=$1 AND user_id=$2`, vid, uid)
	if err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		http.Error(w, "No habías votado este video.", http.StatusBadRequest)
		return
	}

	if _, err := tx.Exec(`UPDATE videos SET votes = GREATEST(votes - 1, 0) WHERE id=$1`, vid); err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, TXerror, http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderClass, HeaderJSON)
	json.NewEncoder(w).Encode(map[string]string{"message": "Voto retirado."})
}

// GET /api/public/rankings?city=...
func (h *PublicHandler) Rankings(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	var rows *sql.Rows
	var err error
	if city != "" {
		rows, err = h.DB.Query(`
			SELECT u.first_name, u.last_name, u.city, SUM(v.votes) as total
			FROM videos v JOIN users u ON u.id=v.user_id
			WHERE v.status='processed' AND u.city=$1
			GROUP BY u.id ORDER BY total DESC`, city)
	} else {
		rows, err = h.DB.Query(`
			SELECT u.first_name, u.last_name, u.city, SUM(v.votes) as total
			FROM videos v JOIN users u ON u.id=v.user_id
			WHERE v.status='processed'
			GROUP BY u.id ORDER BY total DESC`)
	}
	if err != nil {
		http.Error(w, DBerror, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type row struct {
		Position int    `json:"position"`
		Username string `json:"username"`
		City     string `json:"city"`
		Votes    int    `json:"votes"`
	}
	out := []row{}
	i := 1
	for rows.Next() {
		var fn, ln, c string
		var total int
		if err := rows.Scan(&fn, &ln, &c, &total); err != nil {
			http.Error(w, DBerror, http.StatusInternalServerError)
			return
		}
		out = append(out, row{Position: i, Username: fn + " " + ln, City: c, Votes: total})
		i++
	}
	w.Header().Set(HeaderClass, HeaderJSON)
	json.NewEncoder(w).Encode(out)
}
