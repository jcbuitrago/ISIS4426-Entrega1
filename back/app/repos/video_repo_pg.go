package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ISIS4426-Entrega1/app/models"
)

type VideoRepoPG struct{ DB *sql.DB }

func NewVideoRepoPG(db *sql.DB) *VideoRepoPG { return &VideoRepoPG{DB: db} }

var ErrNotFound = errors.New("video not found")

func (r *VideoRepoPG) Create(v models.Video) (models.Video, error) {
	const q = `
	INSERT INTO videos (title, status, uploaded_at, processed_at, origin_url, processed_url, votes, user_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	RETURNING id, uploaded_at, processed_at`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, q,
		v.Title, v.Status, v.UploadedAt, v.ProcessedAt, v.OriginURL, v.ProcessedURL, v.Votes, v.UserID,
	).Scan(&v.VideoID, &v.UploadedAt, &v.ProcessedAt)
	return v, err
}

func (r *VideoRepoPG) List(ctx context.Context, limit, offset int) ([]models.Video, error) {
	const q = `
	SELECT id, title, status, uploaded_at, processed_at, origin_url, processed_url, votes, user_id
	FROM videos
	ORDER BY id DESC
	LIMIT $1 OFFSET $2`
	rows, err := r.DB.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Video
	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.VideoID, &v.Title, &v.Status, &v.UploadedAt, &v.ProcessedAt,
			&v.OriginURL, &v.ProcessedURL, &v.Votes, &v.UserID); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VideoRepoPG) ListByUser(ctx context.Context, userID, limit, offset int) ([]models.Video, error) {
	const q = `
	SELECT id, title, status, uploaded_at, processed_at, origin_url, processed_url, votes, user_id
	FROM videos
	WHERE user_id = $1
	ORDER BY id DESC
	LIMIT $2 OFFSET $3`
	rows, err := r.DB.QueryContext(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Video
	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.VideoID, &v.Title, &v.Status, &v.UploadedAt, &v.ProcessedAt,
			&v.OriginURL, &v.ProcessedURL, &v.Votes, &v.UserID); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VideoRepoPG) GetByID(ctx context.Context, id int) (*models.Video, error) {
	const q = `
	SELECT id, title, status, uploaded_at, processed_at, origin_url, processed_url, votes, user_id
	FROM videos WHERE id = $1`
	var v models.Video
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&v.VideoID, &v.Title, &v.Status, &v.UploadedAt, &v.ProcessedAt,
		&v.OriginURL, &v.ProcessedURL, &v.Votes, &v.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VideoRepoPG) Delete(ctx context.Context, id int) error {
	const q = `DELETE FROM videos WHERE id=$1`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.DB.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *VideoRepoPG) UpdateStatus(ctx context.Context, id int, status models.VideoStatus, updatedAt time.Time) error {
	const q = `UPDATE videos SET status=$1, processed_at=$2 WHERE id=$3`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.DB.ExecContext(ctx, q, status, updatedAt, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *VideoRepoPG) UpdateProcessedURL(ctx context.Context, id int, url string, updatedAt time.Time) error {
	const q = `UPDATE videos SET processed_url=$1, processed_at=$2 WHERE id=$3`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.DB.ExecContext(ctx, q, url, updatedAt, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
