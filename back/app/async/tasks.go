package async

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const TypeProcessVideo = "video:process"

type ProcessVideoPayload struct {
	VideoID int    `json:"video_id"`
	SrcPath string `json:"src_path"`
	UserID  int    `json:"user_id"`
}

type Enqueuer struct{ Client *asynq.Client }

func NewEnqueuer(redisAddr string) *Enqueuer {
	return &Enqueuer{Client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})}
}

func (e *Enqueuer) EnqueueProcessVideo(ctx context.Context, p ProcessVideoPayload) (*asynq.TaskInfo, error) {
	b, _ := json.Marshal(p)
	t := asynq.NewTask(TypeProcessVideo, b, asynq.MaxRetry(5), asynq.Timeout(15*time.Minute))
	return e.Client.EnqueueContext(ctx, t, asynq.Queue("videos"))
}
