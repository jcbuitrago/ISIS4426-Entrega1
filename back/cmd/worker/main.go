package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/models"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/services"
	appdb "ISIS4426-Entrega1/db"

	"github.com/hibiken/asynq"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type server struct{ svc *services.VideoService }

func (s *server) process(ctx context.Context, t *asynq.Task) error {
	var p async.ProcessVideoPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	if err := s.svc.UpdateStatus(p.VideoID, models.StatusProcessing); err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	return s.svc.UpdateStatus(p.VideoID, models.StatusProcessed)
}

func main() {
	sqlDB := appdb.MustOpen()
	svc := services.NewVideoService(repos.NewVideoRepoPG(sqlDB))

	redis := getenv("REDIS_ADDR", "127.0.0.1:6379")
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: redis}, asynq.Config{Queues: map[string]int{"videos": 10}})
	mux := asynq.NewServeMux()
	s := &server{svc: svc}
	mux.HandleFunc(async.TypeProcessVideo, s.process)

	log.Printf("Worker listening on %s", redis)
	log.Fatal(srv.Run(mux))
}
