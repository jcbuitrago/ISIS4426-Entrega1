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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func run(cmd *exec.Cmd) error {
	log.Printf("[worker] exec start cmd=%q", cmd.String())
	clone := exec.Command(cmd.Path, cmd.Args[1:]...)
	clone.Env = cmd.Env
	clone.Dir = cmd.Dir

	log.Printf("[worker] Running: %s", clone.String())
	out, err := clone.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s: %w\noutput:\n%s", clone.String(), err, string(out))
	}
	return nil
}

func processVideo(ctx context.Context, p async.VideoProcessingPayload, svc *services.VideoService, status *async.SQSEnqueuer, s3Client *s3client.S3Client) error {
	_ = status.SetStatus(ctx, p.JobID, "processing:downloading", 24*time.Hour)

	// Create temporary work dir
	workDir := filepath.Join("/tmp", fmt.Sprintf("video_%d_%s", p.VideoID, p.JobID))
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:create_workdir", 24*time.Hour)
		return fmt.Errorf("create workdir: %w", err)
	}
	defer os.RemoveAll(workDir)

	originalFile := filepath.Join(workDir, "original.mp4")
	reader, err := s3Client.DownloadFromUploads(ctx, p.InputPath)
	if err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:download_original", 24*time.Hour)
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	outFile, err := os.Create(originalFile)
	if err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:create_temp_file", 24*time.Hour)
		return fmt.Errorf("create temp file: %w", err)
	}
	if _, err := io.Copy(outFile, reader); err != nil {
		outFile.Close()
		_ = status.SetStatus(ctx, p.JobID, "failed:save_temp_file", 24*time.Hour)
		return fmt.Errorf("save temp file: %w", err)
	}
	outFile.Close()

	_ = status.SetStatus(ctx, p.JobID, "processing:trim", 24*time.Hour)

	mainTrim := filepath.Join(workDir, "video_trim.mp4")
	scaledMain := filepath.Join(workDir, "video_720p.mp4")
	intro := "/assets/intro.mp4"
	outro := "/assets/outro.mp4"
	intro720 := filepath.Join(workDir, "intro_720.mp4")
	outro720 := filepath.Join(workDir, "outro_720.mp4")
	final := filepath.Join(workDir, "final.mp4")
	noaudio := filepath.Join(workDir, "final_noaudio.mp4")
	thumb := filepath.Join(workDir, "thumb.jpg")

	if err := run(trimTo30(originalFile, mainTrim)); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:trim", 24*time.Hour)
		return fmt.Errorf("trim: %w", err)
	}
	if err := run(extractThumbnail(originalFile, thumb)); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:thumb", 24*time.Hour)
		return fmt.Errorf("thumb: %w", err)
	}

	_ = status.SetStatus(ctx, p.JobID, "processing:scale", 24*time.Hour)
	if err := run(to720p16x9(mainTrim, scaledMain)); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:scale_main", 24*time.Hour)
		return fmt.Errorf("scale main: %w", err)
	}
	if err := run(to720p16x9(intro, intro720)); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:scale_intro", 24*time.Hour)
		return fmt.Errorf("scale intro: %w", err)
	}
	if err := run(to720p16x9(outro, outro720)); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:scale_outro", 24*time.Hour)
		return fmt.Errorf("scale outro: %w", err)
	}

	_ = status.SetStatus(ctx, p.JobID, "processing:concat", 24*time.Hour)
	if err := run(concatIntroMainOutro(intro720, scaledMain, outro720, final)); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:concat", 24*time.Hour)
		return fmt.Errorf("concat: %w", err)
	}

	cmdNoAudio := exec.Command("ffmpeg", "-y", "-i", final, "-an", "-c:v", "copy", noaudio)
	if err := run(cmdNoAudio); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:mute", 24*time.Hour)
		return fmt.Errorf("mute: %w", err)
	}

	_ = status.SetStatus(ctx, p.JobID, "processing:uploading", 24*time.Hour)
	processedKey := fmt.Sprintf("processed/%d_final_noaudio.mp4", p.VideoID)
	processedFile, err := os.Open(noaudio)
	if err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:open_processed", 24*time.Hour)
		return fmt.Errorf("open processed: %w", err)
	}
	defer processedFile.Close()
	if err := s3Client.UploadToProcessed(ctx, processedKey, processedFile); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:upload_processed", 24*time.Hour)
		return fmt.Errorf("upload processed: %w", err)
	}

	thumbKey := fmt.Sprintf("processed/%d_thumb.jpg", p.VideoID)
	thumbFile, err := os.Open(thumb)
	if err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:open_thumb", 24*time.Hour)
		return fmt.Errorf("open thumb: %w", err)
	}
	defer thumbFile.Close()
	if err := s3Client.UploadToProcessed(ctx, thumbKey, thumbFile); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:upload_thumb", 24*time.Hour)
		return fmt.Errorf("upload thumb: %w", err)
	}

	if err := svc.UpdateStatus(ctx, p.VideoID, models.StatusProcessing); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:update_status_processing", 24*time.Hour)
		return fmt.Errorf("update status processing: %w", err)
	}

	processedURL := s3Client.GetProcessedFileURL(processedKey)
	thumbURL := s3Client.GetProcessedFileURL(thumbKey)

	if err := svc.UpdateProcessedURL(ctx, p.VideoID, processedURL); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:update_processed_url", 24*time.Hour)
		return fmt.Errorf("update processed url: %w", err)
	}
	if err := svc.UpdateThumbURL(ctx, p.VideoID, thumbURL); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:update_thumb_url", 24*time.Hour)
		return fmt.Errorf("update thumb url: %w", err)
	}
	if err := svc.UpdateStatus(ctx, p.VideoID, models.StatusProcessed); err != nil {
		_ = status.SetStatus(ctx, p.JobID, "failed:update_status_processed", 24*time.Hour)
		return fmt.Errorf("update status processed: %w", err)
	}

	_ = status.SetStatus(ctx, p.JobID, "done", 24*time.Hour)
	log.Printf("Video %d processed successfully (job %s)", p.VideoID, p.JobID)
	return nil
}

func main() {
	queueURL := getenv("SQS_QUEUE_URL", "")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL is required!")
	}
	region := getenv("AWS_REGION", "")
	if region == "" {
		log.Fatal("AWS_REGION is required!")
	}
	dsn := getenv("DB_DSN", "")
	if dsn == "" {
		log.Fatal("DB_DSN is required!")
	}

	db := repos.MustOpenPostgres(dsn)
	defer db.Close()

	repo := repos.NewVideoRepoPG(db)
	svc := services.NewVideoService(repo)

	// Initialize SQS client for worker
	statusStore, err := async.NewSQSEnqueuer(context.Background(), queueURL, db)
	if err != nil {
		log.Fatalf("Cannot initialize SQS enqueuer: %v", err)
	}
	defer statusStore.Close()
	log.Printf("[worker] SQS enqueuer initialized with url: %s", queueURL)

	go func(e *async.SQSEnqueuer) {
		t := time.NewTicker(6 * time.Hour)
		defer t.Stop()
		for range t.C {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := e.CleanupExpiredStatuses(ctx); err != nil {
				log.Printf("job_status cleanup error: %v", err)
			}
			cancel()
		}
	}(statusStore)

	// Initialize S3 client for worker
	s3Client, err := s3client.NewFromSSM(
		context.Background(),
		"/anb/s3/uploads-bucket",
		"/anb/s3/processed-bucket",
	)
	if err != nil {
		log.Fatal("Failed to initialize S3 client:", err)
	}
	log.Printf("[worker] startup s3_uploads_bucket=%s s3_processed_bucket=%s", s3Client.GetUploadsBucket(), s3Client.GetProcessedBucket())

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Cannot load AWS config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)

	log.Printf("Worker started. SQS queue=%s region=%s", queueURL, region)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		resp, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 5,
			WaitTimeSeconds:     20,
			MessageAttributeNames: []string{
				"All",
			},
		})
		cancel()
		if err != nil {
			log.Printf("Recieved error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if len(resp.Messages) == 0 {
			continue
		}

		for _, msg := range resp.Messages {
			jobTypeAttr, ok := msg.MessageAttributes["JobType"]
			if !ok || jobTypeAttr.StringValue == nil || *jobTypeAttr.StringValue != async.TypeProcessVideo {
				// Ignore unrelated messages
				continue
			}
			if msg.Body == nil {
				continue
			}

			var payload async.VideoProcessingPayload
			if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
				log.Printf("Unmarshal payload error: %v", err)
				continue
			}

			procCtx, procCancel := context.WithTimeout(context.Background(), 30*time.Minute)
			err := processVideo(procCtx, payload, svc, statusStore, s3Client)
			procCancel()
			if err != nil {
				log.Printf("Video processing Failed. Job %s failed: %v", payload.JobID, err)
				continue
			}

			if _, err := sqsClient.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				log.Printf("Delete message failed (job %s): %v", payload.JobID, err)
			}
		}
	}
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
