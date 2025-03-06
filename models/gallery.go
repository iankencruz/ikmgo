package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Gallery struct {
	ID    int
	Title string
}

type GalleryModel struct {
	DB *pgxpool.Pool
}

// Create adds a new gallery to the database
func (g *GalleryModel) Create(title string) error {
	_, err := g.DB.Exec(context.Background(), "INSERT INTO galleries (title) VALUES ($1)", title)
	return err
}

// GetAll fetches all galleries
func (g *GalleryModel) GetAll() ([]map[string]interface{}, error) {
	rows, err := g.DB.Query(context.Background(),
		`SELECT g.id, g.title, 
                (SELECT COUNT(*) FROM media WHERE media.gallery_id = g.id) AS media_count
         FROM galleries g
         ORDER BY g.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galleries []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var mediaCount int

		err := rows.Scan(&id, &title, &mediaCount)
		if err != nil {
			return nil, err
		}

		// Store each gallery as a map (key-value)
		gallery := map[string]interface{}{
			"ID":         id,
			"Title":      title,
			"MediaCount": mediaCount, // ✅ Includes media count
		}
		galleries = append(galleries, gallery)
	}
	return galleries, nil
}

// GetByID retrieves a single gallery by its ID
func (g *GalleryModel) GetByID(id int) (*Gallery, error) {
	var gallery Gallery
	err := g.DB.QueryRow(context.Background(), "SELECT id, title FROM galleries WHERE id=$1", id).
		Scan(&gallery.ID, &gallery.Title)

	if err != nil {
		return nil, err
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
