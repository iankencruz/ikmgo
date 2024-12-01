// ./internal/models/gallery.go
package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Gallery struct {
	ID          int64
	Title       string
	Description string
	CoverImage  string
	CreatedAt   time.Time
}

type GalleryModel struct {
	DB *pgxpool.Pool
}

// NewGalleryModel creates and returns a new instance of GalleryModel
func NewGalleryModel(db *pgxpool.Pool) *GalleryModel {
	return &GalleryModel{
		DB: db,
	}
}

func (m *GalleryModel) CreateGallery(title, description, coverImage string) (int64, error) {
	query := `
		INSERT INTO galleries (title, description, cover_image, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`

	var galleryID int64
	err := m.DB.QueryRow(context.Background(), query, title, description, coverImage).Scan(&galleryID)
	if err != nil {
		return 0, err
	}

	return galleryID, nil
}
