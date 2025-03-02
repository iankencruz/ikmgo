package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	ID        int
	FileName  string
	GalleryID int
}

type MediaModel struct {
	DB *pgxpool.Pool
}

// Insert adds a new media file and links it to a gallery
func (m *MediaModel) Insert(fileName string, galleryID int) error {
	_, err := m.DB.Exec(context.Background(),
		"INSERT INTO media (file_name, gallery_id) VALUES ($1, $2)", fileName, galleryID)
	return err
}

// GetByGalleryID retrieves all media files for a specific gallery
func (m *MediaModel) GetByGalleryID(galleryID int) ([]*Media, error) {
	rows, err := m.DB.Query(context.Background(),
		"SELECT id, file_name FROM media WHERE gallery_id=$1 ORDER BY id DESC", galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		m := &Media{}
		err := rows.Scan(&m.ID, &m.FileName)
		if err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}

// GetByID retrieves a single media file by its ID
func (m *MediaModel) GetByID(id int) (*Media, error) {
	var media Media
	err := m.DB.QueryRow(context.Background(),
		"SELECT id, file_name, gallery_id FROM media WHERE id=$1", id).
		Scan(&media.ID, &media.FileName, &media.GalleryID)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// Delete removes a media file from the database
func (m *MediaModel) Delete(id int) error {
	_, err := m.DB.Exec(context.Background(),
		"DELETE FROM media WHERE id=$1", id)
	return err
}

// DeleteByGalleryID deletes all media files associated with a specific gallery
func (m *MediaModel) DeleteByGalleryID(galleryID int) error {
	_, err := m.DB.Exec(context.Background(),
		"DELETE FROM media WHERE gallery_id=$1", galleryID)
	return err
}
