package async

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

//const TypeProcessVideo = "video:process"

type ProcessVideoPayload struct {
	JobID     string `json:"job_id"`
	VideoID   int    `json:"video_id"`
	UserID    int    `json:"user_id"`
	Title     string `json:"title"`
	InputPath string `json:"input_path"`
}

type Enqueuer struct {
	Client *asynq.Client
	Redis  *redis.Client
}

func NewEnqueuer(redisAddr string) *Enqueuer {
	return &Enqueuer{
		Client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
		Redis:  redis.NewClient(&redis.Options{Addr: redisAddr}),
	}
}

func jobKey(jobID string) string { return fmt.Sprintf("jobs:%s:status", jobID) }

func (e *Enqueuer) SetStatus(ctx context.Context, jobID, status string, ttl time.Duration) error {
	return e.Redis.Set(ctx, jobKey(jobID), status, ttl).Err()
}

func (e *Enqueuer) GetStatus(ctx context.Context, jobID string) (string, error) {
	return e.Redis.Get(ctx, jobKey(jobID)).Result()
}

func (e *Enqueuer) EnqueueVideoProcessing(ctx context.Context, videoID, userID int, title, tmpPath string) (string, error) {
	jobID := uuid.NewString()
	b, _ := json.Marshal(ProcessVideoPayload{
		JobID: jobID, VideoID: videoID, UserID: userID, Title: title, InputPath: tmpPath,
	})
	task := asynq.NewTask(TypeProcessVideo, b)
	if _, err := e.Client.EnqueueContext(ctx, task, asynq.Queue("videos")); err != nil {
		return "", err
	}
	_ = e.SetStatus(ctx, jobID, "queued", 24*time.Hour)
	return jobID, nil
}
