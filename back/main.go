package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

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
	log.Println("Starting ANB API Server...")

	log.Println("Initializing database connection...")
	sqlDB := appdb.MustOpen()
	log.Println("Database connection established! ✅")

	log.Println("Initializing repositories and services...")
	repo := repos.NewVideoRepoPG(sqlDB)
	svc := services.NewVideoService(repo)
	log.Println("Repositories and services initialized ✅")

	queueURL := getenv("SQS_QUEUE_URL", "")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL is required")
	}

	enq, err := async.NewSQSEnqueuer(context.Background(), queueURL, sqlDB)
	if err != nil {
		log.Fatalf("Cannot initialize SQS enqueuer: %v", err)
	}
	defer enq.Close()

	// Initialize S3 client from SSM parameters
	log.Println("Initializing S3 service...")
	s3Client, err := s3client.NewFromSSM(
		context.Background(),
		"/anb/s3/uploads-bucket",
		"/anb/s3/processed-bucket",
	)
	if err != nil {
		log.Printf("❌ S3 client initialization failed: %v", err)
		log.Fatal("Cannot initialize S3 client")
	}
	log.Println("✅ S3 client initialized")

	// auth
	log.Println("Initializing auth services...")
	userRepo := repos.NewUserRepoPG(sqlDB)
	authSvc := services.NewAuthService(userRepo)
	authH := routers.NewAuthHandler(authSvc, s3Client) // Pass S3 client for avatar uploads
	log.Println("✅ Auth services initialized")

	// videos - pass S3 client to handler
	log.Println("🎬 Initializing video handlers...")
	h := routers.NewVideosHandler(enq, svc, s3Client)
	hJobs := routers.NewJobsHandler(enq)
	pubH := routers.NewPublicHandler(sqlDB)
	log.Println("✅ Video handlers initialized")

	log.Println("🌐 Setting up routes...")
	r := mux.NewRouter()

	// Add request logging middleware
	r.Use(loggingMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	}).Methods("GET")

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
	api.HandleFunc("/jobs/{id}", hJobs.GetJobStatus).Methods("GET")

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

	log.Println("✅ Routes configured")

	// Remove static file serving - files will be served directly from S3
	// r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/data/"))))

	allowedOrigins := []string{
		"http://localhost:3000",    // Local development
		"http://localhost:5173",    // Vite dev server
		getenv("FRONTEND_URL", ""), // Your S3 website URL
		"http://anb-frontend.s3-website-us-east-1.amazonaws.com",
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
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "Origin"}),
		handlers.ExposedHeaders([]string{"Content-Length"}),
		handlers.AllowCredentials(),
		handlers.MaxAge(300), // Cache preflight requests for 5 minutes
	)

	port := getenv("PORT", "8080")
	log.Printf("🎯 API server starting on port %s", port)
	log.Printf("📍 Environment: DB_DSN=%s", getenv("DB_DSN", "not set"))
	log.Printf("📍 Environment: SQS_QUEUE_URL=%s", queueURL)
	log.Printf("📍 Environment: AWS_REGION=%s", getenv("AWS_REGION", "not set"))

	log.Printf("✨ ANB API Server ready and listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(r)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("📝 %s %s - %d - %v - %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
