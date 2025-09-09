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
	"ISIS4426-Entrega1/app/middleware"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}

func main() {
	sqlDB := appdb.MustOpen()
	repo := repos.NewVideoRepoPG(sqlDB)
	svc := services.NewVideoService(repo)

	enq := async.NewEnqueuer(getenv("REDIS_ADDR", "redis:6379"))
	defer enq.Client.Close()

	// auth
	userRepo := repos.NewUserRepoPG(sqlDB)
	authSvc := services.NewAuthService(userRepo)
	authH := routers.NewAuthHandler(authSvc)

	// videos
	h := routers.NewVideosHandler(enq, svc)
	hJobs := routers.NewJobsHandler(enq)
	pubH := routers.NewPublicHandler(sqlDB)

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// auth routes
	api.HandleFunc("/auth/signup", authH.Signup).Methods("POST")
	api.HandleFunc("/auth/login", authH.Login).Methods("POST")

	// profile routes (protected)
	me := api.PathPrefix("").Subrouter()
	me.Use(middleware.AuthRequired)
	me.HandleFunc("/me", authH.Me).Methods("GET")
	me.HandleFunc("/me", authH.UpdateMe).Methods("PUT")
	me.HandleFunc("/me/avatar", authH.UploadAvatar).Methods("POST")

	// protected videos
	videos := api.PathPrefix("/videos").Subrouter()
	videos.Use(middleware.AuthRequired)
	videos.HandleFunc("", h.Create).Methods("POST")
	videos.HandleFunc("", h.List).Methods("GET")
	videos.HandleFunc("/{id}", h.GetByID).Methods("GET")
	videos.HandleFunc("/{id}", h.Delete).Methods("DELETE")

	// jobs (optional, public)
	api.HandleFunc("/jobs/{id}", hJobs.Get).Methods("GET")

	// public endpoints
	api.HandleFunc("/public/videos", pubH.ListVideos).Methods("GET")
	vote := api.PathPrefix("/public/videos").Subrouter()
	vote.Use(middleware.AuthRequired)
	vote.HandleFunc("/{id}/vote", pubH.Vote).Methods("POST")
	vote.HandleFunc("/{id}/vote", pubH.Unvote).Methods("DELETE")
	my := api.PathPrefix("/public").Subrouter()
	my.Use(middleware.AuthRequired)
	my.HandleFunc("/my-votes", pubH.MyVotes).Methods("GET")
	api.HandleFunc("/public/rankings", pubH.Rankings).Methods("GET")

	// static files
	fs := http.FileServer(http.Dir("/data"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	addr := ":" + getenv("PORT", "8080")
	log.Println("api: listening on", addr)

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}), // ajustar en prod
	)
	log.Fatal(http.ListenAndServe(addr, cors(r)))
}
