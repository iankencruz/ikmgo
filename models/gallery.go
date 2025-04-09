package models

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Gallery struct {
	ID            int
	Title         string
	Slug          string
	Featured      bool
	Description   string
	CoverImageID  *int    // ✅ Cover Image ID (Nullable)
	CoverImageURL *string // ✅ Stores image URL (Joined from media)
	MediaCount    int     // ✅ Count of Media Items
	Published     bool
}

type GalleryModel struct {
	DB *pgxpool.Pool
}

// GetByTitle retrieves a single gallery by its title

func (g *GalleryModel) GetByTitle(title string) (*Gallery, []*Media, error) {
	var gallery Gallery

	// Fetch the gallery by title
	err := g.DB.QueryRow(context.Background(),
		"SELECT id, title, slug, description FROM galleries WHERE title=$1", title).
		Scan(&gallery.ID, &gallery.Title, &gallery.Slug, &gallery.Description)

	if err != nil {
		return nil, nil, err
	}

	// Fetch associated images with mime_type safely set
	rows, err := g.DB.Query(context.Background(),
		`SELECT m.id, m.file_name, m.thumbnail_url, m.full_url,
		        COALESCE(m.mime_type, '') AS mime_type,
		        gm.position
		 FROM media m 
		 JOIN gallery_media gm ON m.id = gm.media_id
		 WHERE gm.gallery_id = $1
		 ORDER BY gm.position ASC`, gallery.ID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var images []*Media
	for rows.Next() {
		var media Media
		err := rows.Scan(
			&media.ID,
			&media.FileName,
			&media.ThumbnailURL,
			&media.FullURL,
			&media.MimeType, // now safe
			&media.Position,
		)
		if err != nil {
			return nil, nil, err
		}
		images = append(images, &media)
	}

	return &gallery, images, nil
}

// SetFeatured updates the featured gallery
func (g *GalleryModel) SetFeatured(id int) error {
	// Unset the previous featured gallery
	_, err := g.DB.Exec(context.Background(), "UPDATE galleries SET featured = FALSE WHERE featured = TRUE")
	if err != nil {
		return err
	}

	// Set the new featured gallery
	res, err := g.DB.Exec(context.Background(), "UPDATE galleries SET featured = TRUE WHERE id = $1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no gallery found with ID %d", id)
	}
	return nil
}

// Create adds a new gallery to the database
func (g *GalleryModel) Create(title, description, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	_, err := g.DB.Exec(context.Background(), "INSERT INTO galleries (title, description, slug) VALUES ($1, $2, $3)", title, description, slug)
	return err
}

func (g *GalleryModel) GetAllPublic() ([]map[string]interface{}, error) {
	rows, err := g.DB.Query(context.Background(),
		`SELECT g.id, g.title, g.slug, g.description, g.cover_image_id, m.full_url AS cover_image_url,
                (SELECT COUNT(*) FROM gallery_media WHERE gallery_media.gallery_id = g.id) AS media_count
         FROM galleries g
         LEFT JOIN media m ON g.cover_image_id = m.id
         WHERE g.published = TRUE
         ORDER BY g.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galleries []map[string]interface{}
	for rows.Next() {
		var id int
		var description string
		var slug string
		var title string
		var coverImageID *int
		var coverImageURL *string
		var mediaCount int

		err := rows.Scan(&id, &title, &slug, &description, &coverImageID, &coverImageURL, &mediaCount)
		if err != nil {
			return nil, err
		}

		gallery := map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"Slug":          slug,
			"CoverImageID":  coverImageID,
			"CoverImageURL": coverImageURL,
			"MediaCount":    mediaCount,
		}
		galleries = append(galleries, gallery)
	}
	return galleries, nil
}

// GetAll fetches all galleries

func (g *GalleryModel) GetAll() ([]map[string]interface{}, error) {
	rows, err := g.DB.Query(context.Background(),
		`SELECT g.id, g.title, g.slug, g.description, g.published, m.full_url AS cover_image_url,
	       (SELECT COUNT(*) FROM gallery_media WHERE gallery_media.gallery_id = g.id) AS media_count
		FROM galleries g
		LEFT JOIN media m ON g.cover_image_id = m.id
		ORDER BY g.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galleries []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var slug string
		var published bool
		var coverImageURL *string
		var mediaCount int
		var description string

		err := rows.Scan(&id, &title, &slug, &description, &published, &coverImageURL, &mediaCount)
		if err != nil {
			return nil, err
		}

		// Store each gallery as a map (key-value)
		gallery := map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"CoverImageURL": coverImageURL,
			"MediaCount":    mediaCount, // ✅ Media count included
			"Published":     published,  // ✅ Include Published field
		}
		galleries = append(galleries, gallery)
	}
	return galleries, nil
}

// SetCoverImage updates the cover image of a gallery
func (g *GalleryModel) SetCoverImage(galleryID, mediaID int) error {
	res, err := g.DB.Exec(context.Background(),
		"UPDATE galleries SET cover_image_id = $1 WHERE id = $2", mediaID, galleryID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no gallery found with ID %d", galleryID)
	}
	return nil
}

// GetByID retrieves a single gallery by its ID

func (g *GalleryModel) GetByID(id int) (*Gallery, error) {
	var gallery Gallery
	err := g.DB.QueryRow(context.Background(), `
		SELECT g.id, g.title, g.slug, g.description, g.published, g.cover_image_id, m.full_url AS cover_image_url,
			   (SELECT COUNT(*) FROM gallery_media WHERE gallery_media.gallery_id = g.id) AS media_count
		FROM galleries g
		LEFT JOIN media m ON g.cover_image_id = m.id
		WHERE g.id = $1
		`, id).
		Scan(&gallery.ID, &gallery.Title, &gallery.Slug, &gallery.Description, &gallery.Published, &gallery.CoverImageID, &gallery.CoverImageURL, &gallery.MediaCount)

	if err != nil {
		log.Printf("⚠️ Scan fallback due to broken cover_image_id: %v", err)

		// fallback query without the join
		err = g.DB.QueryRow(context.Background(),
			`SELECT id, title, description, slug, published, cover_image_id FROM galleries WHERE id = $1`, id).
			Scan(&gallery.ID, &gallery.Title, &gallery.Description, &gallery.Slug, &gallery.Published, &gallery.CoverImageID)

		// set to nil manually
		gallery.CoverImageURL = nil

		if err != nil {
			return nil, err
		}
	}

	return &gallery, nil
}

// Update updates the title of an existing gallery
func (g *GalleryModel) Update(id int, title, description, slug string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}

	res, err := g.DB.Exec(context.Background(), "UPDATE galleries SET title=$1, description=$2, slug=$3 WHERE id=$4", title, description, slug, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no gallery found with ID %d", id)
	}
	return nil
}

// Delete removes a gallery from the database
func (g *GalleryModel) Delete(id int) error {
	result, err := g.DB.Exec(context.Background(), "DELETE FROM galleries WHERE id=$1", id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no gallery found with ID %d", id)
	}

	return nil
}

func (g *GalleryModel) SetPublished(id int, published bool) error {
	result, err := g.DB.Exec(context.Background(), "UPDATE galleries SET published=$1 WHERE id=$2", published, id)
	if err != nil {
		return fmt.Errorf("failed to set published status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected, gallery with ID %d may not exist", id)
	}

	return nil
}

// GetMedia returns all media linked to a gallery via the gallery_media join table

func (g *GalleryModel) GetMediaPaginated(galleryID, limit, offset int) ([]*Media, error) {
	rows, err := g.DB.Query(context.Background(), `
		SELECT m.id, m.file_name, m.thumbnail_url, m.full_url,
			   COALESCE(m.mime_type, '') AS mime_type,
			   gm.position
		FROM gallery_media gm
		JOIN media m ON gm.media_id = m.id
		WHERE gm.gallery_id = $1
		ORDER BY gm.position ASC
		LIMIT $2 OFFSET $3
	`, galleryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		var m Media
		if err := rows.Scan(&m.ID, &m.FileName, &m.ThumbnailURL, &m.FullURL, &m.MimeType, &m.Position); err != nil {
			return nil, err
		}
		media = append(media, &m)
	}

	return media, nil
}

// AttachMedia links a media item to a gallery with an optional position
func (g *GalleryModel) AttachMedia(galleryID, mediaID, position int) error {
	_, err := g.DB.Exec(context.Background(), `
		INSERT INTO gallery_media (gallery_id, media_id, position)
		VALUES ($1, $2, $3)
		ON CONFLICT (gallery_id, media_id) DO UPDATE
		SET position = EXCLUDED.position
	`, galleryID, mediaID, position)
	return err
}

func (g *GalleryModel) ReorderMedia(galleryID, mediaID, newPosition int) error {
	tx, err := g.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	var currentPosition int
	err = tx.QueryRow(context.Background(), `
		SELECT position FROM gallery_media
		WHERE gallery_id = $1 AND media_id = $2
	`, galleryID, mediaID).Scan(&currentPosition)

	if err != nil {
		return fmt.Errorf("failed to fetch current position: %w", err)
	}

	// No position change
	if currentPosition == newPosition {
		return nil
	}

	// Move everything else to make room
	if currentPosition < newPosition {
		_, err = tx.Exec(context.Background(), `
			UPDATE gallery_media
			SET position = position - 1
			WHERE gallery_id = $1 AND position > $2 AND position <= $3
		`, galleryID, currentPosition, newPosition)
	} else {
		_, err = tx.Exec(context.Background(), `
			UPDATE gallery_media
			SET position = position + 1
			WHERE gallery_id = $1 AND position >= $3 AND position < $2
		`, galleryID, newPosition, currentPosition)
	}
	if err != nil {
		return err
	}

	// Move the target
	_, err = tx.Exec(context.Background(), `
		UPDATE gallery_media
		SET position = $1
		WHERE gallery_id = $2 AND media_id = $3
	`, newPosition, galleryID, mediaID)

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (g *GalleryModel) GetNextPosition(galleryID int) (int, error) {
	var maxPosition int

	err := g.DB.QueryRow(context.Background(), `
		SELECT COALESCE(MAX(position), -1)
		FROM gallery_media
		WHERE gallery_id = $1
	`, galleryID).Scan(&maxPosition)

	if err != nil {
		return 0, err
	}

	return maxPosition + 1, nil
}

// Get Count of galleries in galleries table
func (g *GalleryModel) Count() (int, error) {
	var count int
	err := g.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM galleries
	`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// get 5 latest galleries
func (g *GalleryModel) GetLatest(limit int) ([]*Gallery, error) {
	rows, err := g.DB.Query(context.Background(), `
		SELECT id, title, slug, description, cover_image_id, featured
		FROM galleries
		WHERE published = TRUE
		ORDER BY id DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galleries []*Gallery
	for rows.Next() {
		var gallery Gallery
		err := rows.Scan(&gallery.ID, &gallery.Title, &gallery.Slug, &gallery.Description, &gallery.CoverImageID, &gallery.Featured)
		if err != nil {
			return nil, err
		}
		galleries = append(galleries, &gallery)
	}

	return galleries, nil
}

// Get MediaCount of a gallery
func (g *GalleryModel) GetMediaCount(galleryID int) (int, error) {
	var count int
	err := g.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM gallery_media WHERE gallery_id = $1`, galleryID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (g *GalleryModel) GetBySlug(slug string) (*Gallery, error) {
	var gallery Gallery
	err := g.DB.QueryRow(context.Background(), `
		SELECT id, title, slug, description, cover_image_id, published
		FROM galleries WHERE slug = $1`, slug).Scan(
		&gallery.ID, &gallery.Title, &gallery.Slug, &gallery.Description, &gallery.CoverImageID, &gallery.Published,
	)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}
