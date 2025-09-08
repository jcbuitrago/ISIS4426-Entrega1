package main

import (
	"log"
	"net/http"
	"os"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/routers"
	"ISIS4426-Entrega1/app/services"
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
	sqlDB := appdb.MustOpen()
	repo := repos.NewVideoRepoPG(sqlDB)
	svc := services.NewVideoService(repo)

	enq := async.NewEnqueuer(getenv("REDIS_ADDR", "redis:6379"))
	defer enq.Client.Close()

	h := routers.NewVideosHandler(enq, svc)
	hJobs := routers.NewJobsHandler(enq)

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/videos", h.Create).Methods("POST")
	api.HandleFunc("/videos", h.List).Methods("GET") // general o ?user_id=
	api.HandleFunc("/videos/{id}", h.GetByID).Methods("GET")
	api.HandleFunc("/videos/{id}", h.Delete).Methods("DELETE")

	api.HandleFunc("/jobs/{id}", hJobs.Get).Methods("GET")
	fs := http.FileServer(http.Dir("/data"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	addr := ":" + getenv("PORT", "8080")
	log.Println("api: listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
