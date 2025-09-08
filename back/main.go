package main

import (
	"log"
	"net/http"
	"os"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/routers"

	"github.com/gorilla/mux"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	// DB

	// Redis / Enqueuer
	redis := getenv("REDIS_ADDR", "redis:6379")
	enq := async.NewEnqueuer(redis)
	defer enq.Client.Close()

	// Handlers
	hVideos := routers.NewVideosHandler(enq) // solo encola
	hJobs := routers.NewJobsHandler(enq)

	// Router
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/videos", hVideos.Create).Methods("POST")
	api.HandleFunc("/jobs/{id}", hJobs.Get).Methods("GET")

	// 404 explícito (útil para depurar rutas)
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "route not found: "+r.URL.Path, http.StatusNotFound)
	})

	addr := ":" + getenv("PORT", "8080")
	log.Println("api: listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
