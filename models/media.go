package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	ID           int
	FileName     string
	ThumbnailURL string
	FullURL      string
	GalleryID    sql.NullInt32
	Position     int
}

type MediaModel struct {
	DB *pgxpool.Pool
}

// Insert adds a new media file and links it to a gallery
func (m *MediaModel) Insert(
	filename string,
	url string,
	thumbURL string,
	galleryID int,
	position int,
) error {
	_, err := m.DB.Exec(context.Background(),
		`INSERT INTO media (
			file_name, full_url, thumbnail_url, gallery_id, position
		) VALUES ($1, $2, $3, $4, $5)`,
		filename, url, thumbURL, galleryID, position)

	return err
}

// GetByGalleryID retrieves all media files for a specific gallery

func (m *MediaModel) GetByGalleryID(galleryID int) ([]*Media, error) {
	rows, err := m.DB.Query(context.Background(),
		"SELECT id, file_name, full_url,  thumbnail_url, position FROM media WHERE gallery_id=$1 ORDER BY position ASC", galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		m := &Media{}
		err := rows.Scan(&m.ID, &m.FileName, &m.FullURL, &m.ThumbnailURL, &m.Position)
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

	query := `SELECT id, file_name, full_url, thumbnail_url, position, gallery_id FROM media WHERE id = $1`

	err := m.DB.QueryRow(context.Background(), query, id).Scan(
		&media.ID,
		&media.FileName,
		&media.FullURL,
		&media.ThumbnailURL,
		&media.Position,
		&media.GalleryID,
	)

	if err != nil {
		log.Printf("❌ GetByID failed: %v", err)
		return nil, err
	}

	return &media, nil
}

func (m *MediaModel) GetByIDAndGallery(id int, galleryID int) (*Media, error) {
	var media Media
	err := m.DB.QueryRow(context.Background(),
		`SELECT id, file_name, full_url, thumbnail_url, position, gallery_id
		 FROM media
		 WHERE id = $1 AND gallery_id = $2 order by position ASC`,
		id, galleryID).
		Scan(&media.ID, &media.FileName, &media.FullURL, &media.ThumbnailURL, &media.Position, &media.GalleryID)

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
		`SELECT media.id, media.file_name, media.thumbnail_url, full_url, galleries.title 
		 FROM media 
		 JOIN galleries ON media.gallery_id = galleries.id 
		 ORDER BY media.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []map[string]interface{} // ✅ Return as a slice of maps
	for rows.Next() {
		var id int
		var fileName, thumbnail_url, full_url, galleryTitle string

		err := rows.Scan(&id, &fileName, &thumbnail_url, &full_url, &galleryTitle)
		if err != nil {
			return nil, err
		}

		// Store each media record as a map (key-value)
		mediaItem := map[string]interface{}{
			"ID":           id,
			"FileName":     fileName,
			"ThumbnailURL": thumbnail_url,
			"FullURL":      full_url,
			"GalleryTitle": galleryTitle,
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

func (m *MediaModel) UpdatePositionsInBulk(galleryID int, mediaIDs []int) error {
	ctx := context.Background()
	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Step 1: Clear positions for all involved media to avoid constraint conflict
	ids := make([]interface{}, len(mediaIDs))
	placeholders := make([]string, len(mediaIDs))
	for i, id := range mediaIDs {
		ids[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	_, err = tx.Exec(ctx,
		fmt.Sprintf(
			`UPDATE media SET position = NULL
			 WHERE id IN (%s) AND gallery_id = $%d`,
			strings.Join(placeholders, ","), len(mediaIDs)+1,
		),
		append(ids, galleryID)...,
	)
	if err != nil {
		return fmt.Errorf("failed to nullify existing positions: %w", err)
	}

	// Step 2: Reassign positions without conflicts
	seen := make(map[int]bool)
	position := 0
	for _, mediaID := range mediaIDs {
		if seen[mediaID] {
			continue
		}
		seen[mediaID] = true

		_, err := tx.Exec(ctx,
			`UPDATE media SET position = $1 WHERE id = $2 AND gallery_id = $3`,
			position, mediaID, galleryID,
		)
		if err != nil {
			return fmt.Errorf("failed to update media %d: %w", mediaID, err)
		}
		position++
	}

	return tx.Commit(ctx)
}

func (m *MediaModel) UpdatePositionsForProject(projectID int, mediaIDs []int) error {
	ctx := context.Background()
	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear all positions first
	ids := make([]interface{}, len(mediaIDs))
	placeholders := make([]string, len(mediaIDs))
	for i, id := range mediaIDs {
		ids[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	_, err = tx.Exec(ctx,
		fmt.Sprintf(
			`UPDATE media SET position = NULL
			 WHERE id IN (%s) AND project_id = $%d`,
			strings.Join(placeholders, ","), len(mediaIDs)+1,
		),
		append(ids, projectID)...,
	)
	if err != nil {
		return fmt.Errorf("failed to clear media positions: %w", err)
	}

	// Reassign unique positions
	position := 0
	seen := make(map[int]bool)
	for _, mediaID := range mediaIDs {
		if seen[mediaID] {
			continue
		}
		seen[mediaID] = true

		_, err := tx.Exec(ctx,
			`UPDATE media SET position = $1 WHERE id = $2 AND project_id = $3`,
			position, mediaID, projectID)
		if err != nil {
			return fmt.Errorf("failed to update media %d: %w", mediaID, err)
		}
		position++
	}

	return tx.Commit(ctx)
}

func (m *MediaModel) GetNextProjectPosition(projectID int) (int, error) {
	var pos int
	err := m.DB.QueryRow(context.Background(),
		`SELECT COALESCE(MAX(position), -1) + 1 FROM media WHERE project_id = $1`,
		projectID).Scan(&pos)
	return pos, err
}

func (m *MediaModel) InsertProjectMedia(
	fileName, url, thumbURL string,
	projectID int,
	position int,
) (int, error) {
	var mediaID int

	err := m.DB.QueryRow(context.Background(),
		`INSERT INTO media (
			file_name, full_url, thumbnail_url, project_id, position
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		fileName, url, thumbURL, projectID, position).Scan(&mediaID)

	if err != nil {
		return 0, err
	}

	// ✅ Insert into project_media join table
	_, err = m.DB.Exec(context.Background(),
		`INSERT INTO project_media (project_id, media_id)
		 VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		projectID, mediaID)

	return mediaID, err
}
