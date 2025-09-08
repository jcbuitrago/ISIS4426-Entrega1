package worker

import (
	"context"
	"encoding/json"
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

type server struct {
	svc *services.VideoService
	enq *async.Enqueuer
}

func (s *server) processVideo(ctx context.Context, t *asynq.Task) error {
	var p async.ProcessVideoPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	_ = s.enq.SetStatus(ctx, p.JobID, "processing:start", 24*time.Hour)

	// Rutas de trabajo (compartidas por volumen)
	workDir := "/data/work"
	_ = os.MkdirAll(workDir, 0o775)

	trimmed := filepath.Join(workDir, "trim_"+filepath.Base(p.TmpPath))
	scaled := filepath.Join(workDir, "scaled_"+filepath.Base(p.TmpPath))
	final := filepath.Join(workDir, "final_"+filepath.Base(p.TmpPath))

	// 1) Recortar a 30s
	_ = s.enq.SetStatus(ctx, p.JobID, "processing:trim", 24*time.Hour)
	if out, err := exec.Command("ffmpeg", "-y", "-i", p.TmpPath, "-t", "30", "-c", "copy", trimmed).CombinedOutput(); err != nil {
		log.Println("ffmpeg trim:", string(out))
		_ = s.enq.SetStatus(ctx, p.JobID, "failed", 24*time.Hour)
		return err
	}

	// 2) Escalar a 720p 16:9
	_ = s.enq.SetStatus(ctx, p.JobID, "processing:scale", 24*time.Hour)
	filter := "scale=w=1280:h=720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2"
	if out, err := exec.Command("ffmpeg", "-y", "-i", trimmed, "-vf", filter, "-c:a", "copy", scaled).CombinedOutput(); err != nil {
		log.Println("ffmpeg scale:", string(out))
		_ = s.enq.SetStatus(ctx, p.JobID, "failed", 24*time.Hour)
		return err
	}

	// 3) Concat intro/outro (montadas en /assets)
	_ = s.enq.SetStatus(ctx, p.JobID, "processing:intro-outro", 24*time.Hour)
	intro := "/assets/intro.mp4"
	outro := "/assets/outro.mp4"
	list := final + ".txt"
	if err := os.WriteFile(list, []byte("file '"+intro+"'\nfile '"+scaled+"'\nfile '"+outro+"'\n"), 0o644); err != nil {
		_ = s.enq.SetStatus(ctx, p.JobID, "failed", 24*time.Hour)
		return err
	}
	if out, err := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", list, "-c", "copy", final).CombinedOutput(); err != nil {
		log.Println("ffmpeg concat:", string(out))
		_ = s.enq.SetStatus(ctx, p.JobID, "failed", 24*time.Hour)
		return err
	}

	// 4) Insertar en DB cuando termina
	_ = s.enq.SetStatus(ctx, p.JobID, "processing:db-insert", 24*time.Hour)
	now := time.Now()
	v := models.Video{
		Title:        p.Title,
		OriginURL:    p.TmpPath,
		ProcessedURL: final,
		Status:       models.StatusProcessed,
		UserID:       p.UserID,
		UploadedAt:   now,
		ProcessedAt:  now,
	}
	if _, err := s.svc.Create(p.UserID, v.Title, v.OriginURL); err != nil {
		_ = s.enq.SetStatus(ctx, p.JobID, "failed", 24*time.Hour)
		return err
	}

	_ = s.enq.SetStatus(ctx, p.JobID, "done", 24*time.Hour)
	return nil
}

func main() {
	sqlDB := appdb.MustOpen()
	repo := repos.NewVideoRepoPG(sqlDB)
	svc := services.NewVideoService(repo)

	redis := os.Getenv("REDIS_ADDR")
	if redis == "" {
		redis = "redis:6379"
	}

	enq := async.NewEnqueuer(redis)
	s := &server{svc: svc, enq: enq}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redis},
		asynq.Config{Queues: map[string]int{"videos": 10}}, // 👈 MISMO nombre que en enqueue
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(async.TypeProcessVideo, s.processVideo)

	log.Println("worker: listening queue=videos redis=", redis)
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
