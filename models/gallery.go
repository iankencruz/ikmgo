package models

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Gallery struct {
	ID            int
	Title         string
	Featured      bool
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
		"SELECT id, title FROM galleries WHERE title=$1", title).
		Scan(&gallery.ID, &gallery.Title)

	if err != nil {
		return nil, nil, err
	}

	// Fetch associated images sorted by position in ASCENDING order
	rows, err := g.DB.Query(context.Background(),
		"SELECT id, file_name, thumbnail_url,  full_url, position FROM media WHERE gallery_id = $1 ORDER BY position ASC", gallery.ID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var images []*Media
	for rows.Next() {
		var media Media
		err := rows.Scan(&media.ID, &media.FileName, &media.ThumbnailURL, &media.FullURL, &media.Position)
		if err != nil {
			return nil, nil, err
		}
		images = append(images, &media)
	}

	return &gallery, images, nil
}

// GetFeatured retrieves the gallery marked as featured
func (g *GalleryModel) GetFeatured() (*Gallery, []*Media, error) {
	var gallery Gallery

	// Fetch the featured gallery
	err := g.DB.QueryRow(context.Background(),
		"SELECT id, title FROM galleries WHERE featured = TRUE LIMIT 1").
		Scan(&gallery.ID, &gallery.Title)

	if err != nil {
		return nil, nil, err
	}

	// Fetch associated images including their URL
	rows, err := g.DB.Query(context.Background(),
		"SELECT id, file_name, thumbnail_url,  full_url, position FROM media WHERE gallery_id = $1 ORDER BY id ASC", gallery.ID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var images []*Media
	for rows.Next() {
		var media Media
		err := rows.Scan(&media.ID, &media.FileName, &media.ThumbnailURL, media.FullURL, media.Position) // ✅ Fetch media URL
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
	_, err = g.DB.Exec(context.Background(), "UPDATE galleries SET featured = TRUE WHERE id = $1", id)
	return err
}

// Create adds a new gallery to the database
func (g *GalleryModel) Create(title string) error {
	_, err := g.DB.Exec(context.Background(), "INSERT INTO galleries (title) VALUES ($1)", title)
	return err
}

func (g *GalleryModel) GetAllPublic() ([]map[string]interface{}, error) {
	rows, err := g.DB.Query(context.Background(),
		`SELECT g.id, g.title, g.cover_image_id, m.full_url AS cover_image_url,
                (SELECT COUNT(*) FROM media WHERE media.gallery_id = g.id) AS media_count
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
		var title string
		var coverImageID *int
		var coverImageURL *string
		var mediaCount int

		err := rows.Scan(&id, &title, &coverImageID, &coverImageURL, &mediaCount)
		if err != nil {
			return nil, err
		}

		gallery := map[string]interface{}{
			"ID":            id,
			"Title":         title,
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
		`SELECT g.id, g.title, g.cover_image_id, g.published, m.full_url AS cover_image_url,
                (SELECT COUNT(*) FROM media WHERE media.gallery_id = g.id) AS media_count
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
		var coverImageID *int
		var coverImageURL *string
		var mediaCount int
		var published bool // ✅ Store as a boolean

		err := rows.Scan(&id, &title, &coverImageID, &published, &coverImageURL, &mediaCount)
		if err != nil {
			return nil, err
		}

		// Store each gallery as a map (key-value)
		gallery := map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"CoverImageID":  coverImageID,
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
	_, err := g.DB.Exec(context.Background(),
		"UPDATE galleries SET cover_image_id = $1 WHERE id = $2", mediaID, galleryID)
	return err
}

// GetByID retrieves a single gallery by its ID

func (g *GalleryModel) GetByID(id int) (*Gallery, error) {
	var gallery Gallery
	err := g.DB.QueryRow(context.Background(), `
		SELECT g.id, g.title, g.cover_image_id, m.thumbnail_url
		FROM galleries g
		LEFT JOIN media m ON g.cover_image_id = m.id
		WHERE g.id = $1`, id).
		Scan(&gallery.ID, &gallery.Title, &gallery.CoverImageID, &gallery.CoverImageURL)

	if err != nil {
		log.Printf("⚠️ Scan fallback due to broken cover_image_id: %v", err)

		// fallback query without the join
		err = g.DB.QueryRow(context.Background(),
			`SELECT id, title, cover_image_id FROM galleries WHERE id = $1`, id).
			Scan(&gallery.ID, &gallery.Title, &gallery.CoverImageID)

		// set to nil manually
		gallery.CoverImageURL = nil

		if err != nil {
			return nil, err
		}
	}

	return &gallery, nil
}

// Update updates the title of an existing gallery
func (g *GalleryModel) Update(id int, title string) error {
	_, err := g.DB.Exec(context.Background(), "UPDATE galleries SET title=$1 WHERE id=$2", title, id)
	return err
}

// Delete removes a gallery from the database
func (g *GalleryModel) Delete(id string) error {
	_, err := g.DB.Exec(context.Background(), "DELETE FROM galleries WHERE id=$1", id)
	return err
}

func (g *GalleryModel) SetPublished(id int, published bool) error {
	_, err := g.DB.Exec(context.Background(), "UPDATE galleries SET published=$1 WHERE id=$2", published, id)
	return err
}
