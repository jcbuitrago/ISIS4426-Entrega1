package async

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

const TypeProcessVideo = "video:process"

// VideoProcessingPayload represents the message structure for SQS
type VideoProcessingPayload struct {
	JobID     string `json:"job_id"`
	VideoID   int    `json:"video_id"`
	InputPath string `json:"input_path"` // S3 key
	Title     string `json:"title"`
	UserID    int    `json:"user_id"`
}

// SQSEnqueuer replaces the Asynq-based Enqueuer
type SQSEnqueuer struct {
	sqsClient *sqs.Client
	queueURL  string
	db        *sql.DB
}

// NewSQSEnqueuer create a new SQS-based job enqueuer
func NewSQSEnqueuer(ctx context.Context, queueURL string, db *sql.DB) (*SQSEnqueuer, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &SQSEnqueuer{
		sqsClient: sqs.NewFromConfig(cfg),
		queueURL:  queueURL,
		db:        db,
	}, nil
}

// EnqueueVideoProcessing sends a video processing job to SQS
func (e *SQSEnqueuer) EnqueueVideoProcessing(
	ctx context.Context,
	videoID int,
	userID int,
	title string,
	inputPath string,
) (string, error) {
	jobID := uuid.New().String()

	payload := VideoProcessingPayload{
		JobID:     jobID,
		VideoID:   videoID,
		UserID:    userID,
		Title:     title,
		InputPath: inputPath,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SQS message body: %w", err)
	}

	// Sends message to SQS
	_, err = e.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(e.queueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"JobType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(TypeProcessVideo),
			},
			"VideoID": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(fmt.Sprintf("%d", videoID)),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("Failed to send SQS message: %w", err)
	}

	// Store job status in DB
	if err := e.createJobStatus(ctx, jobID, "queued"); err != nil {
		return "", fmt.Errorf("failed to store job status: %w", err)
	}

	return jobID, nil
}

// SetStatus updates the job status in PostgreSQL
func (e *SQSEnqueuer) SetStatus(ctx context.Context, jobID string, status string, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)

	const query = `
		UPDATE job_status
		SET status = $1, updated_at = NOW(), expires_at = $2
		WHERE job_id = $3
	`

	_, err := e.db.ExecContext(ctx, query, status, expiresAt, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// GetStatus retrieces the job status from postgresql
func (e *SQSEnqueuer) GetStatus(ctx context.Context, jobID string) (string, error) {
	const query = `
		SELECT status
		FROM job_status
		WHERE job_id = $1 AND expires_at > NOW()
	`
	var status string
	err := e.db.QueryRowContext(ctx, query, jobID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("job not found or expired")
		}
		return "", fmt.Errorf("failed to get job status: %w", err)
	}

	return status, nil
}

// Ping check if SQS is accessible (for health checks)
func (e *SQSEnqueuer) Ping(ctx context.Context) error {
	_, err := e.sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(e.queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
		},
	})

	return err
}

// Close is a no-op for SQS (no persistent connection)
func (e *SQSEnqueuer) Close() error {
	return nil
}

// Helper: createJobStatus inserts initial job status
func (e *SQSEnqueuer) createJobStatus(ctx context.Context, jobID string, status string) error {
	const query = `
		INSERT INTO job_status (job_id,status, created_at, updated_at, expires_at)
		VALUES ($1, $2, NOW(), NOW(), NOW() + INTERVAL '24 hours')
	`
	_, err := e.db.ExecContext(ctx, query, jobID, status)
	if err != nil {
		return fmt.Errorf("failed to insert job instatus: %w", err)
	}

	return nil
}

// Helper: Cleanup expired job statuses (runs periodically)
func (e *SQSEnqueuer) cleanupExpiredJobStatuses(ctx context.Context) error {
	const query = `DELETE FROM job_status WHERE expires_at < NOW()`
	_, err := e.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error while cleaning up expired job statuses: %w", err)
	}

	return nil
}

func (e *SQSEnqueuer) CleanupExpiredStatuses(ctx context.Context) error {
	return e.cleanupExpiredJobStatuses(ctx)
}
