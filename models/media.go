package models

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	ID           int
	FileName     string
	ThumbnailURL string
	FullURL      string
	MimeType     *string
	EmbedURL     *string
	Position     int
}

type MediaModel struct {
	DB *pgxpool.Pool
}

// --- Insert methods ---
func (m *MediaModel) InsertAndReturnID(fileName, fullURL, thumbURL string) (int, error) {

	var id int
	err := m.DB.QueryRow(context.Background(),
		`INSERT INTO media (file_name, full_url, thumbnail_url)
		 VALUES ($1, $2, $3) RETURNING id`,
		fileName, fullURL, thumbURL).Scan(&id)
	return id, err
}

func (m *MediaModel) UploadAndLinkProjectMedia(fileName, url, thumbURL string, projectID, position int) (int, error) {
	var mediaID int
	err := m.DB.QueryRow(context.Background(),
		`INSERT INTO media (file_name, full_url, thumbnail_url, position, project_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		fileName, url, thumbURL, position, projectID).Scan(&mediaID)
	if err != nil {
		return 0, err
	}

	_, err = m.DB.Exec(context.Background(),
		`INSERT INTO project_media (project_id, media_id)
		 VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		projectID, mediaID)

	return mediaID, err
}

func (m *MediaModel) InsertProjectMedia(projectID, mediaID int) error {
	query := `
		INSERT INTO project_media (project_id, media_id, position)
		VALUES ($1, $2, (
			SELECT COALESCE(MAX(position), 0) + 1 FROM project_media WHERE project_id = $1
		))
		ON CONFLICT DO NOTHING;
	`
	_, err := m.DB.Exec(context.Background(), query, projectID, mediaID)
	return err
}

func (m *MediaModel) InsertGalleryMedia(galleryID, mediaID int) error {
	query := `
		INSERT INTO gallery_media (gallery_id, media_id, position)
		VALUES ($1, $2, (
			SELECT COALESCE(MAX(position), 0) + 1 FROM gallery_media WHERE gallery_id = $1
		))
		ON CONFLICT DO NOTHING;
	`
	_, err := m.DB.Exec(context.Background(), query, galleryID, mediaID)
	return err
}

// --- Get methods ---
func (m *MediaModel) GetAll() ([]*Media, error) {
	rows, err := m.DB.Query(context.Background(), `
		SELECT id, file_name, full_url, thumbnail_url, COALESCE(embed_url, '')
		FROM media
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		m := &Media{}
		err := rows.Scan(
			&m.ID,
			&m.FileName,
			&m.FullURL,
			&m.ThumbnailURL,
			&m.EmbedURL, // Now safe: always a non-nil string
		)
		if err != nil {
			return nil, err
		}
		media = append(media, m)
	}

	return media, nil
}

func (m *MediaModel) GetByGalleryID(galleryID int) ([]*Media, error) {
	rows, err := m.DB.Query(context.Background(),
		`SELECT m.id, m.file_name, m.full_url, m.thumbnail_url, m.embed_url, m.mime_type, gm.position
		 FROM gallery_media gm
		 JOIN media m ON gm.media_id = m.id
		 WHERE gm.gallery_id = $1
		 ORDER BY gm.position ASC`, galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		var m Media
		err := rows.Scan(&m.ID, &m.FileName, &m.FullURL, &m.ThumbnailURL, &m.EmbedURL, &m.MimeType, &m.Position)
		if err != nil {
			return nil, err
		}
		media = append(media, &m)
	}
	return media, nil
}

func (m *MediaModel) GetByID(id int) (*Media, error) {
	query := `SELECT id, file_name, full_url, thumbnail_url, mime_type, embed_url FROM media WHERE id = $1`
	row := m.DB.QueryRow(context.Background(), query, id)

	var media Media
	err := row.Scan(&media.ID, &media.FileName, &media.FullURL, &media.ThumbnailURL, &media.MimeType, &media.EmbedURL)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (m *MediaModel) GetByIDAndGallery(id, galleryID int) (*Media, error) {
	var media Media
	err := m.DB.QueryRow(context.Background(),
		`SELECT m.id, m.file_name, m.full_url, m.thumbnail_url, m.embed_url, m.mime_type, gm.position
		 FROM gallery_media gm
		 JOIN media m ON gm.media_id = m.id
		 WHERE m.id = $1 AND gm.gallery_id = $2`,
		id, galleryID).Scan(
		&media.ID, &media.FileName, &media.FullURL, &media.ThumbnailURL, &media.EmbedURL, &media.MimeType, &media.Position,
	)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// --- Position helpers ---
func (m *MediaModel) GetNextPosition(galleryID int) (int, error) {
	var pos int
	err := m.DB.QueryRow(context.Background(),
		`SELECT COALESCE(MAX(position), -1) + 1
		 FROM gallery_media
		 WHERE gallery_id = $1`, galleryID).Scan(&pos)
	return pos, err
}

func (m *MediaModel) GetNextProjectPosition(projectID int) (int, error) {
	var pos int
	err := m.DB.QueryRow(context.Background(), `
		SELECT COALESCE(MAX(position), 0) + 1 FROM project_media WHERE project_id=$1
	`, projectID).Scan(&pos)
	return pos, err
}

func (m *MediaModel) ReorderPositions(galleryID, mediaID, newPosition int) error {
	tx, err := m.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	var current int
	err = tx.QueryRow(context.Background(),
		`SELECT position FROM gallery_media
		 WHERE gallery_id = $1 AND media_id = $2`,
		galleryID, mediaID).Scan(&current)
	if err != nil {
		return err
	}

	if current == newPosition {
		return nil
	}

	if current < newPosition {
		_, err = tx.Exec(context.Background(),
			`UPDATE gallery_media
			 SET position = position - 1
			 WHERE gallery_id = $1 AND position > $2 AND position <= $3`,
			galleryID, current, newPosition)
	} else {
		_, err = tx.Exec(context.Background(),
			`UPDATE gallery_media
			 SET position = position + 1
			 WHERE gallery_id = $1 AND position >= $3 AND position < $2`,
			galleryID, current, newPosition)
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(),
		`UPDATE gallery_media
		 SET position = $1
		 WHERE gallery_id = $2 AND media_id = $3`,
		newPosition, galleryID, mediaID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (m *MediaModel) UpdatePositionsForProject(projectID int, mediaIDs []int) error {
	ctx := context.Background()
	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for position, mediaID := range mediaIDs {
		_, err := tx.Exec(ctx,
			`UPDATE media SET position = $1
			 WHERE id = $2 AND project_id = $3`,
			position, mediaID, projectID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (m *MediaModel) UpdatePositionsForGallery(galleryID int, mediaIDs []int) error {
	tx, err := m.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	for i, mediaID := range mediaIDs {
		_, err := tx.Exec(context.Background(), `
			UPDATE gallery_media
			SET position = $1
			WHERE gallery_id = $2 AND media_id = $3
		`, i, galleryID, mediaID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func (m *MediaModel) UpdatePositionsInBulk(galleryID int, mediaIDs []int) error {
	ctx := context.Background()
	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for position, mediaID := range mediaIDs {
		_, err := tx.Exec(ctx,
			`UPDATE gallery_media SET position = $1
			 WHERE gallery_id = $2 AND media_id = $3`,
			position, galleryID, mediaID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// --- Misc ---
func (m *MediaModel) MediaExists(mediaID, galleryID int) (bool, error) {
	var exists bool
	err := m.DB.QueryRow(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM gallery_media
			WHERE media_id = $1 AND gallery_id = $2
		)`, mediaID, galleryID).Scan(&exists)
	return exists, err
}

func (m *MediaModel) Delete(id int) error {
	_, err := m.DB.Exec(context.Background(), `DELETE FROM media WHERE id = $1`, id)
	return err
}

func (m *MediaModel) AttachToProject(projectID, mediaID, position int) error {
	_, err := m.DB.Exec(context.Background(), `
		INSERT INTO project_media (project_id, media_id, position)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, media_id) DO UPDATE
		SET position = EXCLUDED.position
	`, projectID, mediaID, position)
	return err
}

func (m *MediaModel) GetUnlinkedMedia(joinTable, foreignKey string, id int) ([]*Media, error) {
	query := fmt.Sprintf(`
		SELECT id, file_name, full_url, thumbnail_url, COALESCE(mime_type, ''), COALESCE(embed_url, '')
		FROM media
		WHERE id NOT IN (
			SELECT media_id FROM %s WHERE %s = $1
		)
		ORDER BY id DESC
	`, joinTable, foreignKey)

	// log.Printf("🕵️ GetUnlinkedMedia: Running query for %s.%s = %d", joinTable, foreignKey, id)

	rows, err := m.DB.Query(context.Background(), query, id)
	if err != nil {
		log.Printf("❌ Query failed in GetUnlinkedMedia: %v", err)
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		var m Media
		var mime, embed string

		if err := rows.Scan(&m.ID, &m.FileName, &m.FullURL, &m.ThumbnailURL, &mime, &embed); err != nil {
			log.Printf("❌ Scan failed for media row: %v", err)
			continue
		}

		if mime != "" {
			m.MimeType = &mime
		}
		if embed != "" {
			m.EmbedURL = &embed
		}

		// log.Printf("✅ Unlinked Media Found: ID=%d Name=%s", m.ID, m.FileName)
		media = append(media, &m)
	}

	log.Printf("📦 Total unlinked media found: %d", len(media))
	return media, nil
}

func (m *MediaModel) UnlinkMediaFromGallery(galleryID, mediaID int) error {
	_, err := m.DB.Exec(context.Background(),
		`DELETE FROM gallery_media WHERE gallery_id = $1 AND media_id = $2`,
		galleryID, mediaID)
	return err
}

func (m *MediaModel) UnlinkMediaFromProject(projectID, mediaID int) error {
	_, err := m.DB.Exec(context.Background(),
		`DELETE FROM project_media WHERE project_id = $1 AND media_id = $2`,
		projectID, mediaID)
	return err
}

func (m *MediaModel) GetByIDUnsafe(id int) (*Media, error) {
	query := `
	SELECT id, file_name, full_url, thumbnail_url,
	       COALESCE(mime_type, '') AS mime_type,
	       COALESCE(embed_url, '') AS embed_url
	FROM media
	WHERE id = $1`

	var media Media

	err := m.DB.QueryRow(context.Background(), query, id).
		Scan(&media.ID, &media.FileName, &media.FullURL, &media.ThumbnailURL, &media.MimeType, &media.EmbedURL)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (m *MediaModel) GetPaginated(limit, offset int) ([]*Media, error) {

	rows, err := m.DB.Query(context.Background(), `
	SELECT id, file_name, thumbnail_url, full_url, mime_type, embed_url
	FROM media
	ORDER BY id DESC
	LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		log.Printf("❌ GetPaginated query error: %v", err) // ✅ this must be here
		return nil, err
	}

	defer rows.Close()

	var mediaList []*Media
	for rows.Next() {
		var media Media
		err := rows.Scan(
			&media.ID,
			&media.FileName,
			&media.ThumbnailURL,
			&media.FullURL,
			&media.MimeType,
			&media.EmbedURL,
		)
		if err != nil {
			log.Printf("❌ GetPaginated scan error: %v", err)
			return nil, err
		}
		mediaList = append(mediaList, &media)
	}

	return mediaList, nil
}

func (m *MediaModel) Count() (int, error) {
	var count int
	err := m.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM media`).Scan(&count)
	return count, err
}

// Get 5 latest media
func (m *MediaModel) GetLatest(limit int) ([]*Media, error) {
	rows, err := m.DB.Query(context.Background(), `
	SELECT id, file_name, thumbnail_url, full_url, mime_type, embed_url
	FROM media
	ORDER BY id DESC
	LIMIT $1
`, limit)
	if err != nil {
		log.Printf("❌ GetLatest query error: %v", err)
		return nil, err
	}

	defer rows.Close()

	var mediaList []*Media
	for rows.Next() {
		var media Media
		err := rows.Scan(
			&media.ID,
			&media.FileName,
			&media.ThumbnailURL,
			&media.FullURL,
			&media.MimeType,
			&media.EmbedURL,
		)
		if err != nil {
			log.Printf("❌ GetLatest scan error: %v", err)
			return nil, err
		}
		mediaList = append(mediaList, &media)
	}

	return mediaList, nil
}

func (m *MediaModel) GetUnlinkedMediaPaginated(joinTable, foreignKey string, id, limit, offset int) ([]*Media, int, error) {
	query := fmt.Sprintf(`
		SELECT id, file_name, full_url, thumbnail_url
		FROM media
		WHERE id NOT IN (
			SELECT media_id FROM %s WHERE %s = $1
		)
		ORDER BY id DESC
		LIMIT $2 OFFSET $3`, joinTable, foreignKey)

	rows, err := m.DB.Query(context.Background(), query, id, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		var m Media
		if err := rows.Scan(&m.ID, &m.FileName, &m.FullURL, &m.ThumbnailURL); err != nil {
			continue
		}
		media = append(media, &m)
	}

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM media
		WHERE id NOT IN (SELECT media_id FROM %s WHERE %s = $1)`, joinTable, foreignKey)
	err = m.DB.QueryRow(context.Background(), countQuery, id).Scan(&total)

	return media, total, err
}
