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

	"github.com/hibiken/asynq"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}

func run(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func handleProcessVideo(svc *services.VideoService, enq *async.Enqueuer) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p async.ProcessVideoPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil { return err }

		_ = enq.SetStatus(ctx, p.JobID, "processing:trim", 24*time.Hour)

		// paths
		input := p.InputPath
		work := filepath.Dir(input)
		mainTrim := filepath.Join(work, fmt.Sprintf("video_%d_trim.mp4", p.VideoID))
		scaledMain := filepath.Join(work, fmt.Sprintf("video_%d_720p.mp4", p.VideoID))
		intro := "/assets/intro.mp4"
		outro := "/assets/outro.mp4"
		intro720 := filepath.Join(work, "intro_720.mp4")
		outro720 := filepath.Join(work, "outro_720.mp4")
		final := filepath.Join("/data", "processed", fmt.Sprintf("%d_final.mp4", p.VideoID))

		_ = os.MkdirAll(filepath.Dir(final), 0o775)

		// 1) Trim
		if err := run(trimTo30(input, mainTrim)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:trim", 24*time.Hour)
			return err
		}
		// 2) Scale main
		_ = enq.SetStatus(ctx, p.JobID, "processing:scale", 24*time.Hour)
		if err := run(to720p16x9(mainTrim, scaledMain)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_main", 24*time.Hour)
			return err
		}
		// 3) Scale intro/outro
		if err := run(to720p16x9(intro, intro720)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_intro", 24*time.Hour)
			return err
		}
		if err := run(to720p16x9(outro, outro720)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_outro", 24*time.Hour)
			return err
		}
		// 4) Concat
		_ = enq.SetStatus(ctx, p.JobID, "processing:concat", 24*time.Hour)
		if err := run(concatIntroMainOutro(intro720, scaledMain, outro720, final)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:concat", 24*time.Hour)
			return err
		}
		// 5) Eliminar audio del final (requisito)
		noaudio := filepath.Join("/data", "processed", fmt.Sprintf("%d_final_noaudio.mp4", p.VideoID))
		cmdNoAudio := exec.Command("ffmpeg", "-y", "-i", final, "-an", "-c:v", "copy", noaudio)
		if err := run(cmdNoAudio); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:mute", 24*time.Hour)
			return err
		}
		final = noaudio

		// 6) Actualizar DB
		if err := svc.UpdateStatus(ctx, p.VideoID, models.StatusProcessing); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_status_processing", 24*time.Hour)
			return err
		}
		publicURL := fmt.Sprintf("/static/processed/%d_final_noaudio.mp4", p.VideoID)
		if err := svc.UpdateProcessedURL(ctx, p.VideoID, publicURL); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_processed_url", 24*time.Hour)
			return err
		}
		if err := svc.UpdateStatus(ctx, p.VideoID, models.StatusProcessed); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_status_processed", 24*time.Hour)
			return err
		}
		_ = enq.SetStatus(ctx, p.JobID, "completed", 24*time.Hour)
		return nil
	}
}

func main() {
	redis := getenv("REDIS_ADDR", "redis:6379")
	dsn := getenv("DB_DSN", "postgres://postgres:postgres@db:5432/anb-showcase?sslmode=disable")

	db := repos.MustOpenPostgres(dsn)
	defer db.Close()

	repo := repos.NewVideoRepoPG(db)
	svc := services.NewVideoService(repo)
	enq := async.NewEnqueuer(redis)

	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: redis},
		asynq.Config{Queues: map[string]int{"videos": 10}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(async.TypeProcessVideo, handleProcessVideo(svc, enq))

	log.Println("worker: listening on redis", redis, "queue=videos")
	if err := srv.Run(mux); err != nil { log.Fatal(err) }
}
