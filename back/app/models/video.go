package models

import "time"

type VideoStatus string

const (
	StatusUploaded   VideoStatus = "uploaded"
	StatusProcessing VideoStatus = "processing"
	StatusProcessed  VideoStatus = "processed"
	StatusFailed     VideoStatus = "failed"
)

type Video struct {
	VideoID      int         `json:"video_id"`
	Title        string      `json:"title,omitempty"`
	Status       VideoStatus `json:"status,omitempty"`
	UploadedAt   time.Time   `json:"uploaded_at,omitempty"`
	ProcessedAt  time.Time   `json:"processed_at,omitempty"`
	OriginURL    string      `json:"origin_url,omitempty"`
	ProcessedURL string      `json:"processed_url,omitempty"`
	ThumbURL     string      `json:"thumb_url,omitempty"`
	Votes        int         `json:"votes"`
	UserID       int         `json:"user_id,omitempty"`
}

type CreateVideoRequest struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}
