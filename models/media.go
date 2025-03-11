package models

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	ID        int
	FileName  string
	URL       string
	GalleryID int
	Position  int
}

type MediaModel struct {
	DB *pgxpool.Pool
}

// Insert adds a new media file and links it to a gallery
func (m *MediaModel) Insert(fileName, fileURL string, galleryID, position int) error {
	_, err := m.DB.Exec(context.Background(),
		"INSERT INTO media (file_name, url, gallery_id, position) VALUES ($1, $2, $3, $4)",
		fileName, fileURL, galleryID, position)

	return err
}

// GetByGalleryID retrieves all media files for a specific gallery

func (m *MediaModel) GetByGalleryID(galleryID int) ([]*Media, error) {
	rows, err := m.DB.Query(context.Background(),
		"SELECT id, file_name, url, position FROM media WHERE gallery_id=$1 ORDER BY position DESC", galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		m := &Media{}
		err := rows.Scan(&m.ID, &m.FileName, &m.URL, &m.Position)
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
		"SELECT id, file_name, url, gallery_id FROM media WHERE id=$1", id).
		Scan(&media.ID, &media.FileName, &media.URL, &media.GalleryID)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// Delete removes a media file from the database
func (m *MediaModel) Delete(id int) error {
	_, err := m.DB.Exec(context.Background(), "DELETE FROM media WHERE id=$1", id)
	return err
}

// DeleteByGalleryID deletes all media files associated with a specific gallery
func (m *MediaModel) DeleteByGalleryID(galleryID int) error {
	_, err := m.DB.Exec(context.Background(),
		"DELETE FROM media WHERE gallery_id=$1", galleryID)
	return err
}

func (m *MediaModel) GetAll() ([]map[string]interface{}, error) {
	rows, err := m.DB.Query(context.Background(),
		`SELECT media.id, media.file_name, media.url, galleries.title 
		 FROM media 
		 JOIN galleries ON media.gallery_id = galleries.id 
		 ORDER BY media.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []map[string]interface{} // ✅ Return as a slice of maps
	for rows.Next() {
		var id int
		var fileName, url, galleryTitle string

		err := rows.Scan(&id, &fileName, &url, &galleryTitle)
		if err != nil {
			return nil, err
		}

		// Store each media record as a map (key-value)
		mediaItem := map[string]interface{}{
			"ID":           id,
			"FileName":     fileName,
			"URL":          url,
			"GalleryTitle": galleryTitle, // ✅ No change to Media struct
		}
		media = append(media, mediaItem)
	}
	return media, nil
}

func (m *MediaModel) UpdatePosition(mediaID, newPosition, galleryID int) error {
	_, err := m.DB.Exec(context.Background(),
		"UPDATE media SET position = $1 WHERE id = $2 AND gallery_id = $3", newPosition, mediaID, galleryID)
	return err
}

func (m *MediaModel) ReorderPositions(galleryID, mediaID, newPosition int) error {
	tx, err := m.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// ✅ Fetch current position
	var currentPosition int
	err = tx.QueryRow(context.Background(),
		"SELECT position FROM media WHERE id = $1 AND gallery_id = $2", mediaID, galleryID).
		Scan(&currentPosition)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("❌ Media ID %d not found in gallery %d", mediaID, galleryID)
			return nil // ✅ Exit gracefully instead of failing
		}
		return err
	}

	// If position hasn't changed, do nothing
	if currentPosition == newPosition {
		return nil
	}

	// ✅ Adjust positions dynamically
	if currentPosition < newPosition {
		_, err = tx.Exec(context.Background(),
			"UPDATE media SET position = position - 1 WHERE gallery_id = $1 AND position > $2 AND position <= $3",
			galleryID, currentPosition, newPosition)
	} else {
		_, err = tx.Exec(context.Background(),
			"UPDATE media SET position = position + 1 WHERE gallery_id = $1 AND position >= $3 AND position < $2",
			galleryID, newPosition, currentPosition)
	}

	if err != nil {
		return err
	}

	// ✅ Set new position
	_, err = tx.Exec(context.Background(),
		"UPDATE media SET position = $1 WHERE id = $2 AND gallery_id = $3", newPosition, mediaID, galleryID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background()) // ✅ Apply changes
}

func (m *MediaModel) MediaExists(mediaID, galleryID int) (bool, error) {
	var exists bool
	err := m.DB.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM media WHERE id = $1 AND gallery_id = $2)",
		mediaID, galleryID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m *MediaModel) GetNextPosition(galleryID int) (int, error) {
	var maxPosition int

	err := m.DB.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(position), -1) FROM media WHERE gallery_id = $1", galleryID).
		Scan(&maxPosition)

	if err != nil {
		return 0, err
	}

	return maxPosition + 1, nil // ✅ Assigns the next available position
}
