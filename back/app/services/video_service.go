package services

import (
	"ISIS4426-Entrega1/app/models"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidTitle = errors.New("title is required")
	ErrInvalidURL   = errors.New("url is required")
)

type VideoRepo interface {
	Create(v models.Video) (models.Video, error)
	GetByID(id int) (models.Video, error)
	List() ([]models.Video, error)
	Delete(id int) error
	UpdateStatus(id int, status models.VideoStatus, updatedAt time.Time) error
}

func (s *VideoService) UpdateStatus(id int, st models.VideoStatus) error {
	return s.repo.UpdateStatus(id, st, time.Now())
}

type VideoService struct{ repo VideoRepo }

func NewVideoService(r VideoRepo) *VideoService { return &VideoService{repo: r} }

func (s *VideoService) Create(userID int, title, url string) (models.Video, error) {
	if strings.TrimSpace(title) == "" {
		return models.Video{}, ErrInvalidTitle
	}
	if strings.TrimSpace(url) == "" {
		return models.Video{}, ErrInvalidURL
	}
	now := time.Now()
	v := models.Video{
		Title:       title,
		OriginURL:   url,
		Status:      models.StatusUploaded,
		UploadedAt:  now,
		ProcessedAt: now,
		UserID:      userID,
	}
	return s.repo.Create(v)
}

func (s *VideoService) GetByID(id int) (models.Video, error) { return s.repo.GetByID(id) }
func (s *VideoService) List() ([]models.Video, error)        { return s.repo.List() }
func (s *VideoService) Delete(id int) error                  { return s.repo.Delete(id) }
