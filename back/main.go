package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/middleware"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/routers"
	"ISIS4426-Entrega1/app/services"
	appdb "ISIS4426-Entrega1/db"
	"ISIS4426-Entrega1/internal/s3client"

	"github.com/gorilla/handlers"
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

	// Initialize S3 client from SSM parameters
	s3Client, err := s3client.NewFromSSM(
		context.Background(),
		"/anb/s3/uploads-bucket",
		"/anb/s3/processed-bucket",
	)
	if err != nil {
		log.Fatal("Failed to initialize S3 client:", err)
	}

	// auth
	userRepo := repos.NewUserRepoPG(sqlDB)
	authSvc := services.NewAuthService(userRepo)
	authH := routers.NewAuthHandler(authSvc, s3Client) // Pass S3 client for avatar uploads

	// videos - pass S3 client to handler
	h := routers.NewVideosHandler(enq, svc, s3Client)
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

	// Remove static file serving - files will be served directly from S3
	// r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/data/"))))

	allowedOrigins := []string{
		"http://localhost:3000",    // Local development
		"http://localhost:5173",    // Vite dev server
		getenv("FRONTEND_URL", ""), // Your S3 website URL
	}

	// Remove empty strings
	var validOrigins []string
	for _, origin := range allowedOrigins {
		if origin != "" {
			validOrigins = append(validOrigins, origin)
		}
	}

	// Fallback to allow all if no specific origins configured
	if len(validOrigins) == 0 {
		validOrigins = []string{"*"}
	}

	cors := handlers.CORS(
		handlers.AllowedOrigins(validOrigins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		}),
		handlers.ExposedHeaders([]string{"Content-Length"}),
		handlers.AllowCredentials(),
		handlers.MaxAge(300), // Cache preflight requests for 5 minutes
	)

	port := getenv("PORT", "8080")
	log.Printf("API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(r)))
}
