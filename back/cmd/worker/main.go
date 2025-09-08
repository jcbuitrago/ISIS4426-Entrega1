package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/models"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/services"
	appdb "ISIS4426-Entrega1/db"

	"github.com/hibiken/asynq"
)

func run(cmd *exec.Cmd) error {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd failed: %v\n%s", err, string(out))
	}
	return nil
}

func handleProcessVideo(svc *services.VideoService, enq *async.Enqueuer) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p async.ProcessVideoPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}
		_ = enq.SetStatus(ctx, p.JobID, "processing:start", 24*time.Hour)

		// Archivos y rutas
		input := p.InputPath // viene de la API (/data/uploads/xxx.mp4)
		workDir := "/data/work"
		_ = os.MkdirAll(workDir, 0o775)

		trimmed := filepath.Join(workDir, "trimmed.mp4")
		scaledMain := filepath.Join(workDir, "main_720p.mp4")
		intro := "/assets/intro.mp4"
		outro := "/assets/outro.mp4"
		intro720 := filepath.Join(workDir, "intro_720p.mp4")
		outro720 := filepath.Join(workDir, "outro_720p.mp4")
		final := filepath.Join("/data/processed", fmt.Sprintf("video_%d_final.mp4", p.VideoID))
		_ = os.MkdirAll(filepath.Dir(final), 0o775)

		// 1) Recorta a 30s
		_ = enq.SetStatus(ctx, p.JobID, "processing:trim", 24*time.Hour)
		if err := run(trimTo30(input, trimmed)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:trim", 24*time.Hour)
			return err
		}

		// 2) Escala/pad a 720p 16:9 el clip principal
		_ = enq.SetStatus(ctx, p.JobID, "processing:scale_main", 24*time.Hour)
		if err := run(to720p16x9(trimmed, scaledMain)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_main", 24*time.Hour)
			return err
		}

		// 3) Asegura que intro/outro también sean 720p 16:9
		_ = enq.SetStatus(ctx, p.JobID, "processing:scale_intro_outro", 24*time.Hour)
		if err := run(to720p16x9(intro, intro720)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_intro", 24*time.Hour)
			return err
		}
		if err := run(to720p16x9(outro, outro720)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_outro", 24*time.Hour)
			return err
		}

		// 4) Concat intro + main + outro
		_ = enq.SetStatus(ctx, p.JobID, "processing:concat", 24*time.Hour)
		if err := run(concatIntroMainOutro(intro720, scaledMain, outro720, final)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:concat", 24*time.Hour)
			return err
		}

		// 5) Guarda en DB
		created, err := svc.Create(p.UserID, p.Title, final)
		if err != nil {
			log.Printf("create failed (title=%s url=%s): %v", p.Title, final, err)
			_ = enq.SetStatus(ctx, p.JobID, "failed:db_create", 24*time.Hour)
			return err
		}

		publicURL := fmt.Sprintf("http://%s:%s/static/processed/%s",
			"localhost",
			"8080",
			filepath.Base(final),
		)
		log.Printf("update processed url (id=%d url=%s)", created.VideoID, publicURL)

		if err := svc.UpdateProcessedURL(ctx, created.VideoID, publicURL); err != nil {
			log.Printf("update processed url failed (id=%d url=%s): %v", created.VideoID, publicURL, err)
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_processed_url", 24*time.Hour)
			return err
		}

		if err := svc.UpdateStatus(ctx, created.VideoID, models.StatusProcessed); err != nil {
			log.Printf("update status failed (id=%d status=%s): %v", created.VideoID, models.StatusProcessed, err)
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_status", 24*time.Hour)
			return err
		}

		_ = enq.SetStatus(ctx, p.JobID, "done", 24*time.Hour)

		// Limpieza opcional de temporales
		_ = os.Remove(trimmed)
		_ = os.Remove(scaledMain)
		_ = os.Remove(intro720)
		_ = os.Remove(outro720)

		return nil
	}
}

func main() {
	// DB y dependencias
	sqlDB := appdb.MustOpen()
	repo := repos.NewVideoRepoPG(sqlDB)
	svc := services.NewVideoService(repo)

	// Redis
	redis := os.Getenv("REDIS_ADDR")
	if redis == "" {
		redis = "redis:6379"
	}
	enq := async.NewEnqueuer(redis)
	defer enq.Client.Close()

	// Worker Asynq
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redis},
		asynq.Config{
			// Debe coincidir con la cola usada al encolar en la API (p. ej. "videos")
			Queues: map[string]int{"videos": 10},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(async.TypeProcessVideo, handleProcessVideo(svc, enq))

	log.Println("worker: listening on redis", redis, "queue=videos")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
