// back/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/routers"
	appdb "ISIS4426-Entrega1/db"

	"github.com/gorilla/mux"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	// 1) Abrir Postgres (usa env: DB_DSN o DB_HOST/DB_USER/DB_PASSWORD/DB_NAME)
	sqlDB := appdb.MustOpen()

	// 2) Inyectar repo PG
	repo := repos.NewVideoRepoPG(sqlDB)

	// 3) Encolador Asynq (Redis)
	redisAddr := getenv("REDIS_ADDR", "127.0.0.1:6379")
	enq := async.NewEnqueuer(redisAddr)
	defer enq.Client.Close()

	// 4) Router y handlers
	h := routers.NewVideosHandler(repo, enq)

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/videos", h.Create).Methods("POST")
	api.HandleFunc("/videos", h.List).Methods("GET")
	api.HandleFunc("/videos/{id}", h.GetByID).Methods("GET")
	api.HandleFunc("/videos/{id}", h.Delete).Methods("DELETE")

	addr := ":" + getenv("PORT", "8080")
	log.Println("API listening on http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
