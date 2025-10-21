package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/models"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/services"
	"ISIS4426-Entrega1/internal/s3client"

	"github.com/hibiken/asynq"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func run(cmd *exec.Cmd) error {
	log.Printf("Running: %s", cmd.String())
	return cmd.Run()
}

func handleProcessVideo(svc *services.VideoService, enq *async.Enqueuer, s3Client *s3client.S3Client) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p async.ProcessVideoPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		_ = enq.SetStatus(ctx, p.JobID, "processing:downloading", 24*time.Hour)

		// Create temporary work directory
		workDir := filepath.Join("/tmp", fmt.Sprintf("video_%d_%s", p.VideoID, p.JobID))
		if err := os.MkdirAll(workDir, 0o755); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:create_workdir", 24*time.Hour)
			return fmt.Errorf("failed to create work directory: %w", err)
		}
		defer os.RemoveAll(workDir) // Clean up temp files

		// Download original video from S3 uploads bucket
		originalFile := filepath.Join(workDir, "original.mp4")
		reader, err := s3Client.DownloadFromUploads(ctx, p.InputPath)
		if err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:download_original", 24*time.Hour)
			return fmt.Errorf("failed to download original video: %w", err)
		}
		defer reader.Close()

		// Save to local temp file
		outFile, err := os.Create(originalFile)
		if err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:create_temp_file", 24*time.Hour)
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		if _, err := io.Copy(outFile, reader); err != nil {
			outFile.Close()
			_ = enq.SetStatus(ctx, p.JobID, "failed:save_temp_file", 24*time.Hour)
			return fmt.Errorf("failed to save temp file: %w", err)
		}
		outFile.Close()

		_ = enq.SetStatus(ctx, p.JobID, "processing:trim", 24*time.Hour)

		// Define processing file paths (all in temp directory)
		mainTrim := filepath.Join(workDir, "video_trim.mp4")
		scaledMain := filepath.Join(workDir, "video_720p.mp4")
		intro := "/assets/intro.mp4"
		outro := "/assets/outro.mp4"
		intro720 := filepath.Join(workDir, "intro_720.mp4")
		outro720 := filepath.Join(workDir, "outro_720.mp4")
		final := filepath.Join(workDir, "final.mp4")
		noaudio := filepath.Join(workDir, "final_noaudio.mp4")
		thumb := filepath.Join(workDir, "thumb.jpg")

		// 1) Trim to 30 seconds
		if err := run(trimTo30(originalFile, mainTrim)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:trim", 24*time.Hour)
			return fmt.Errorf("failed to trim video: %w", err)
		}

		// 1.5) Extract thumbnail from original
		if err := run(extractThumbnail(originalFile, thumb)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:thumb", 24*time.Hour)
			return fmt.Errorf("failed to extract thumbnail: %w", err)
		}

		// 2) Scale main video to 720p
		_ = enq.SetStatus(ctx, p.JobID, "processing:scale", 24*time.Hour)
		if err := run(to720p16x9(mainTrim, scaledMain)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_main", 24*time.Hour)
			return fmt.Errorf("failed to scale main video: %w", err)
		}

		// 3) Scale intro/outro to 720p
		if err := run(to720p16x9(intro, intro720)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_intro", 24*time.Hour)
			return fmt.Errorf("failed to scale intro: %w", err)
		}
		if err := run(to720p16x9(outro, outro720)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:scale_outro", 24*time.Hour)
			return fmt.Errorf("failed to scale outro: %w", err)
		}

		// 4) Concatenate intro + main + outro
		_ = enq.SetStatus(ctx, p.JobID, "processing:concat", 24*time.Hour)
		if err := run(concatIntroMainOutro(intro720, scaledMain, outro720, final)); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:concat", 24*time.Hour)
			return fmt.Errorf("failed to concatenate videos: %w", err)
		}

		// 5) Remove audio
		cmdNoAudio := exec.Command("ffmpeg", "-y", "-i", final, "-an", "-c:v", "copy", noaudio)
		if err := run(cmdNoAudio); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:mute", 24*time.Hour)
			return fmt.Errorf("failed to remove audio: %w", err)
		}

		// 6) Upload processed video to S3 processed bucket
		_ = enq.SetStatus(ctx, p.JobID, "processing:uploading", 24*time.Hour)

		processedKey := fmt.Sprintf("processed/%d_final_noaudio.mp4", p.VideoID)
		processedFile, err := os.Open(noaudio)
		if err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:open_processed", 24*time.Hour)
			return fmt.Errorf("failed to open processed video: %w", err)
		}
		defer processedFile.Close()

		if err := s3Client.UploadToProcessed(ctx, processedKey, processedFile); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:upload_processed", 24*time.Hour)
			return fmt.Errorf("failed to upload processed video: %w", err)
		}

		// 7) Upload thumbnail to S3 processed bucket
		thumbKey := fmt.Sprintf("processed/%d_thumb.jpg", p.VideoID)
		thumbFile, err := os.Open(thumb)
		if err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:open_thumb", 24*time.Hour)
			return fmt.Errorf("failed to open thumbnail: %w", err)
		}
		defer thumbFile.Close()

		if err := s3Client.UploadToProcessed(ctx, thumbKey, thumbFile); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:upload_thumb", 24*time.Hour)
			return fmt.Errorf("failed to upload thumbnail: %w", err)
		}

		// 8) Update database with S3 URLs
		if err := svc.UpdateStatus(ctx, p.VideoID, models.StatusProcessing); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_status_processing", 24*time.Hour)
			return fmt.Errorf("failed to update status: %w", err)
		}

		// Generate S3 URLs for processed files
		processedURL := s3Client.GetProcessedFileURL(processedKey)
		thumbURL := s3Client.GetProcessedFileURL(thumbKey)

		if err := svc.UpdateProcessedURL(ctx, p.VideoID, processedURL); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_processed_url", 24*time.Hour)
			return fmt.Errorf("failed to update processed URL: %w", err)
		}

		if err := svc.UpdateThumbURL(ctx, p.VideoID, thumbURL); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_thumb_url", 24*time.Hour)
			return fmt.Errorf("failed to update thumb URL: %w", err)
		}

		// 9) Mark as processed
		if err := svc.UpdateStatus(ctx, p.VideoID, models.StatusProcessed); err != nil {
			_ = enq.SetStatus(ctx, p.JobID, "failed:update_status_processed", 24*time.Hour)
			return fmt.Errorf("failed to update final status: %w", err)
		}

		_ = enq.SetStatus(ctx, p.JobID, "done", 24*time.Hour)
		log.Printf("Video %d processed successfully", p.VideoID)
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

	// Initialize S3 client for worker
	s3Client, err := s3client.NewFromSSM(
		context.Background(),
		"/anb/s3/uploads-bucket",
		"/anb/s3/processed-bucket",
	)
	if err != nil {
		log.Fatal("Failed to initialize S3 client:", err)
	}

	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: redis},
		asynq.Config{Queues: map[string]int{"videos": 10}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(async.TypeProcessVideo, handleProcessVideo(svc, enq, s3Client))

	log.Println("worker: listening on redis", redis, "queue=videos")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
