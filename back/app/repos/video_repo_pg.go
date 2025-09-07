// back/app/repos/video_repo_pg.go
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
	INSERT INTO videos (title, url, status, user_id, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := r.DB.QueryRowContext(ctx, q, v.Title, v.OriginURL, v.Status, v.UserID, v.UploadedAt, v.ProcessedAt).
		Scan(&v.VideoID, &v.UploadedAt, &v.ProcessedAt)
	return v, err
}

func (r *VideoRepoPG) GetByID(id int) (models.Video, error) {
	const q = `SELECT id,title,url,status,user_id,created_at,updated_at FROM videos WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var v models.Video
	err := r.DB.QueryRowContext(ctx, q, id).
		Scan(&v.VideoID, &v.Title, &v.OriginURL, &v.Status, &v.UserID, &v.UploadedAt, &v.ProcessedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Video{}, ErrNotFound
	}
	return v, err
}

func (r *VideoRepoPG) List() ([]models.Video, error) {
	const q = `SELECT id,title,url,status,user_id,created_at,updated_at FROM videos ORDER BY id DESC`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Video
	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.VideoID, &v.Title, &v.OriginURL, &v.Status, &v.UserID, &v.UploadedAt, &v.ProcessedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VideoRepoPG) Delete(id int) error {
	const q = `DELETE FROM videos WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func (r *VideoRepoPG) UpdateStatus(id int, status models.VideoStatus, updatedAt time.Time) error {
	const q = `UPDATE videos SET status=$1, updated_at=$2 WHERE id=$3`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
