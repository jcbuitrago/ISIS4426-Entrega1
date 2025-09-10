package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"ISIS4426-Entrega1/app/models"
)

// ----- Fake Repo (mock manual) -----

type fakeVideoRepo struct {
	// inputs capturados
	gotCreate       *models.Video
	gotGetByID      int
	gotList         struct{ limit, offset int }
	gotListByUser   struct{ userID, limit, offset int }
	gotDelete       int
	gotUpdateStatus struct {
		id     int
		status models.VideoStatus
		at     time.Time
	}
	gotUpdateProcessedURL struct {
		id  int
		url string
		at  time.Time
	}
	gotUpdateThumbURL struct {
		id  int
		url string
		at  time.Time
	}

	// valores de retorno configurables
	retCreate             models.Video
	retGetByID            *models.Video
	retList               []models.Video
	retListByUser         []models.Video
	errCreate             error
	errGetByID            error
	errList               error
	errListByUser         error
	errDelete             error
	errUpdateStatus       error
	errUpdateProcessedURL error
	errUpdateThumbURL     error
}

func (f *fakeVideoRepo) Create(v models.Video) (models.Video, error) {
	f.gotCreate = &v
	return f.retCreate, f.errCreate
}
func (f *fakeVideoRepo) GetByID(ctx context.Context, id int) (*models.Video, error) {
	f.gotGetByID = id
	return f.retGetByID, f.errGetByID
}
func (f *fakeVideoRepo) List(ctx context.Context, limit, offset int) ([]models.Video, error) {
	f.gotList.limit, f.gotList.offset = limit, offset
	return f.retList, f.errList
}
func (f *fakeVideoRepo) ListByUser(ctx context.Context, userID, limit, offset int) ([]models.Video, error) {
	f.gotListByUser.userID, f.gotListByUser.limit, f.gotListByUser.offset = userID, limit, offset
	return f.retListByUser, f.errListByUser
}
func (f *fakeVideoRepo) Delete(ctx context.Context, id int) error {
	f.gotDelete = id
	return f.errDelete
}
func (f *fakeVideoRepo) UpdateStatus(ctx context.Context, id int, status models.VideoStatus, updatedAt time.Time) error {
	f.gotUpdateStatus = struct {
		id     int
		status models.VideoStatus
		at     time.Time
	}{id: id, status: status, at: updatedAt}
	return f.errUpdateStatus
}
func (f *fakeVideoRepo) UpdateProcessedURL(ctx context.Context, id int, url string, updatedAt time.Time) error {
	f.gotUpdateProcessedURL = struct {
		id  int
		url string
		at  time.Time
	}{id: id, url: url, at: updatedAt}
	return f.errUpdateProcessedURL
}
func (f *fakeVideoRepo) UpdateThumbURL(ctx context.Context, id int, url string, updatedAt time.Time) error {
	f.gotUpdateThumbURL = struct {
		id  int
		url string
		at  time.Time
	}{id: id, url: url, at: updatedAt}
	return f.errUpdateThumbURL
}

// ----- Tests -----

func TestVideoService_Create_Success(t *testing.T) {
	f := &fakeVideoRepo{
		retCreate: models.Video{VideoID: 42, Title: "Tiro de 3", OriginURL: "/data/uploads/a.mp4", Status: models.StatusUploaded, UserID: 7},
	}
	s := NewVideoService(f)

	got, err := s.Create(7, "Tiro de 3", "/data/uploads/a.mp4")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Validaciones sobre lo que se envía al repo
	if f.gotCreate == nil {
		t.Fatalf("repo.Create no fue llamado")
	}
	if f.gotCreate.Title != "Tiro de 3" {
		t.Errorf("Title = %q; want %q", f.gotCreate.Title, "Tiro de 3")
	}
	if f.gotCreate.OriginURL != "/data/uploads/a.mp4" {
		t.Errorf("OriginURL = %q; want %q", f.gotCreate.OriginURL, "/data/uploads/a.mp4")
	}
	if f.gotCreate.UserID != 7 {
		t.Errorf("UserID = %d; want %d", f.gotCreate.UserID, 7)
	}
	if f.gotCreate.Status != models.StatusUploaded {
		t.Errorf("Status = %s; want %s", f.gotCreate.Status, models.StatusUploaded)
	}
	if f.gotCreate.UploadedAt.IsZero() {
		t.Error("UploadedAt debe ser no-cero")
	}
	if !f.gotCreate.ProcessedAt.IsZero() {
		t.Error("ProcessedAt debe ser cero al crear")
	}

	// Lo que regresa el servicio
	if got.VideoID != 42 {
		t.Errorf("VideoID = %d; want 42", got.VideoID)
	}
}

func TestVideoService_Create_Validation(t *testing.T) {
	s := NewVideoService(&fakeVideoRepo{})

	_, err := s.Create(1, "", "/x.mp4")
	if !errors.Is(err, ErrInvalidTitle) {
		t.Errorf("esperaba ErrInvalidTitle, got %v", err)
	}

	_, err = s.Create(1, "ok", "  ")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("esperaba ErrInvalidURL, got %v", err)
	}
}

func TestVideoService_UpdateStatus_PassesParams(t *testing.T) {
	f := &fakeVideoRepo{}
	s := NewVideoService(f)

	ctx := context.TODO()
	if err := s.UpdateStatus(ctx, 10, models.StatusProcessing); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if f.gotUpdateStatus.id != 10 {
		t.Errorf("id = %d; want 10", f.gotUpdateStatus.id)
	}
	if f.gotUpdateStatus.status != models.StatusProcessing {
		t.Errorf("status = %s; want %s", f.gotUpdateStatus.status, models.StatusProcessing)
	}
	if f.gotUpdateStatus.at.IsZero() {
		t.Error("updatedAt no debería ser cero")
	}
}

func TestVideoService_List_NormalizesBounds(t *testing.T) {
	f := &fakeVideoRepo{retList: []models.Video{}}
	s := NewVideoService(f)

	ctx := context.TODO()

	// limit <= 0 → 20, offset < 0 → 0
	_, _ = s.List(ctx, -5, -1)
	if f.gotList.limit != 20 {
		t.Errorf("limit = %d; want 20", f.gotList.limit)
	}
	if f.gotList.offset != 0 {
		t.Errorf("offset = %d; want 0", f.gotList.offset)
	}

	// limit > 100 → 20
	_, _ = s.List(ctx, 1000, 3)
	if f.gotList.limit != 20 {
		t.Errorf("limit = %d; want 20 (capped)", f.gotList.limit)
	}
	if f.gotList.offset != 3 {
		t.Errorf("offset = %d; want 3", f.gotList.offset)
	}
}

func TestVideoService_ListByUser_NormalizesBoundsAndUser(t *testing.T) {
	f := &fakeVideoRepo{retListByUser: []models.Video{}}
	s := NewVideoService(f)

	ctx := context.TODO()

	_, _ = s.ListByUser(ctx, 9, 0, -10)
	if f.gotListByUser.userID != 9 {
		t.Errorf("userID = %d; want 9", f.gotListByUser.userID)
	}
	if f.gotListByUser.limit != 20 {
		t.Errorf("limit = %d; want 20", f.gotListByUser.limit)
	}
	if f.gotListByUser.offset != 0 {
		t.Errorf("offset = %d; want 0", f.gotListByUser.offset)
	}
}

func TestVideoService_GetByID_PassesThrough(t *testing.T) {
	want := &models.Video{VideoID: 5, Title: "X"}
	f := &fakeVideoRepo{retGetByID: want}
	s := NewVideoService(f)

	got, err := s.GetByID(context.TODO(), 5)
	if err != nil {
		t.Fatalf("GetByID error = %v", err)
	}
	if got.VideoID != 5 {
		t.Errorf("VideoID = %d; want 5", got.VideoID)
	}
	if f.gotGetByID != 5 {
		t.Errorf("repo got id = %d; want 5", f.gotGetByID)
	}
}

func TestVideoService_Delete_PassesThrough(t *testing.T) {
	f := &fakeVideoRepo{}
	s := NewVideoService(f)

	if err := s.Delete(context.TODO(), 77); err != nil {
		t.Fatalf("Delete error = %v", err)
	}
	if f.gotDelete != 77 {
		t.Errorf("repo got id = %d; want 77", f.gotDelete)
	}
}

func TestVideoService_UpdateProcessedURL_PassesParams(t *testing.T) {
	f := &fakeVideoRepo{}
	s := NewVideoService(f)

	err := s.UpdateProcessedURL(context.TODO(), 11, "http://api/static/processed/a.mp4")
	if err != nil {
		t.Fatalf("UpdateProcessedURL error = %v", err)
	}
	if f.gotUpdateProcessedURL.id != 11 {
		t.Errorf("id = %d; want 11", f.gotUpdateProcessedURL.id)
	}
	if f.gotUpdateProcessedURL.url != "http://api/static/processed/a.mp4" {
		t.Errorf("url = %q; want processed url", f.gotUpdateProcessedURL.url)
	}
	if f.gotUpdateProcessedURL.at.IsZero() {
		t.Error("updatedAt no debería ser cero")
	}
}

func TestVideoService_UpdateThumbURL_PassesParams(t *testing.T) {
	f := &fakeVideoRepo{}
	s := NewVideoService(f)

	err := s.UpdateThumbURL(context.TODO(), 12, "http://api/static/thumbs/a.jpg")
	if err != nil {
		t.Fatalf("UpdateThumbURL error = %v", err)
	}
	if f.gotUpdateThumbURL.id != 12 {
		t.Errorf("id = %d; want 12", f.gotUpdateThumbURL.id)
	}
	if f.gotUpdateThumbURL.url != "http://api/static/thumbs/a.jpg" {
		t.Errorf("url = %q; want thumb url", f.gotUpdateThumbURL.url)
	}
	if f.gotUpdateThumbURL.at.IsZero() {
		t.Error("updatedAt no debería ser cero")
	}
}

// (Opcional) errores propagados desde el repo:
func TestVideoService_RepoErrorsPropagate(t *testing.T) {
	repoErr := errors.New("repo boom")
	f := &fakeVideoRepo{
		errList:               repoErr,
		errGetByID:            repoErr,
		errDelete:             repoErr,
		errUpdateStatus:       repoErr,
		errUpdateProcessedURL: repoErr,
		errUpdateThumbURL:     repoErr,
	}
	s := NewVideoService(f)
	ctx := context.TODO()

	if _, err := s.List(ctx, 10, 0); !errors.Is(err, repoErr) {
		t.Errorf("List err = %v; want %v", err, repoErr)
	}
	if _, err := s.GetByID(ctx, 1); !errors.Is(err, repoErr) {
		t.Errorf("GetByID err = %v; want %v", err, repoErr)
	}
	if err := s.Delete(ctx, 1); !errors.Is(err, repoErr) {
		t.Errorf("Delete err = %v; want %v", err, repoErr)
	}
	if err := s.UpdateStatus(ctx, 1, models.StatusProcessing); !errors.Is(err, repoErr) {
		t.Errorf("UpdateStatus err = %v; want %v", err, repoErr)
	}
	if err := s.UpdateProcessedURL(ctx, 1, "u"); !errors.Is(err, repoErr) {
		t.Errorf("UpdateProcessedURL err = %v; want %v", err, repoErr)
	}
	if err := s.UpdateThumbURL(ctx, 1, "t"); !errors.Is(err, repoErr) {
		t.Errorf("UpdateThumbURL err = %v; want %v", err, repoErr)
	}
}
