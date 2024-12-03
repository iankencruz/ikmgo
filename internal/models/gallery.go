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
	Slug        string
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

func (m *GalleryModel) CreateGallery(title, slug, description, coverImage string) (int64, error) {
	query := `
		INSERT INTO galleries (title, slug, description, cover_image, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`

	var galleryID int64
	err := m.DB.QueryRow(context.Background(), query, title, slug, description, coverImage).Scan(&galleryID)
	if err != nil {
		return 0, err
	}

	return galleryID, nil
}

func (m *GalleryModel) GetGalleries() ([]*Gallery, error) {
	query := `
		SELECT id, title, slug, description, cover_image, created_at
		FROM galleries
		ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galleries []*Gallery
	for rows.Next() {
		g := &Gallery{}
		err := rows.Scan(&g.ID, &g.Title, &g.Slug, &g.Description, &g.CoverImage, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		galleries = append(galleries, g)
	}

	return galleries, nil
}
