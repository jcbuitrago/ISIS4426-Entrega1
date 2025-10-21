package services

import (
	"context"
	"errors"
	"time"

	"ISIS4426-Entrega1/app/models"
)

var (
	ErrInvalidTitle = errors.New("title is required")
	ErrInvalidURL   = errors.New("url is required")
)

type VideoRepo interface {
	Create(v models.Video) (models.Video, error)
	GetByID(ctx context.Context, id int) (*models.Video, error)
	List(ctx context.Context, limit, offset int) ([]models.Video, error)
	ListByUser(ctx context.Context, userID, limit, offset int) ([]models.Video, error)
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status models.VideoStatus, updatedAt time.Time) error
	UpdateProcessedURL(ctx context.Context, id int, url string, updatedAt time.Time) error
	UpdateThumbURL(ctx context.Context, id int, url string, updatedAt time.Time) error
}

type VideoService struct{ repo VideoRepo }

func NewVideoService(r VideoRepo) *VideoService { return &VideoService{repo: r} }

func (s *VideoService) Create(userID int, title, s3Key string) (models.Video, error) {
	// Store the S3 key in origin_url field for now
	// The full S3 URL will be generated when needed
	return s.repo.Create(models.Video{
		Title:     title,
		OriginURL: s3Key, // Store S3 key, not full URL
		Status:    models.StatusUploaded,
		UserID:    userID,
	})
}

func (s *VideoService) UpdateStatus(ctx context.Context, id int, st models.VideoStatus) error {
	return s.repo.UpdateStatus(ctx, id, st, time.Now())
}

func (s *VideoService) List(ctx context.Context, limit, offset int) ([]models.Video, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *VideoService) ListByUser(ctx context.Context, userID, limit, offset int) ([]models.Video, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *VideoService) GetByID(ctx context.Context, id int) (*models.Video, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *VideoService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *VideoService) UpdateProcessedURL(ctx context.Context, id int, url string) error {
	return s.repo.UpdateProcessedURL(ctx, id, url, time.Now())
}

func (s *VideoService) UpdateThumbURL(ctx context.Context, id int, url string) error {
	return s.repo.UpdateThumbURL(ctx, id, url, time.Now())
}
