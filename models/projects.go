package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID            int
	Title         string
	Description   string
	CoverImageID  *int
	CoverImageURL *string
	Published     bool
}

type ProjectModel struct {
	DB *pgxpool.Pool
}

// GetAllPublic returns all published projects with optional cover image
func (p *ProjectModel) GetAllPublic() ([]map[string]interface{}, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT pr.id, pr.title, pr.description, pr.cover_image_id, m.thumbnail_url
		FROM projects pr
		LEFT JOIN media m ON pr.cover_image_id = m.id
		WHERE pr.published = TRUE
		ORDER BY pr.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var id int
		var title, description string
		var coverImageID *int
		var coverImageURL *string

		if err := rows.Scan(&id, &title, &description, &coverImageID, &coverImageURL); err != nil {
			return nil, err
		}

		projects = append(projects, map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"Description":   description,
			"CoverImageID":  coverImageID,
			"CoverImageURL": coverImageURL,
		})
	}

	return projects, nil
}

func (p *ProjectModel) SetPublished(id int, published bool) error {
	_, err := p.DB.Exec(context.Background(),
		`UPDATE projects SET published = $1 WHERE id = $2`,
		published, id,
	)
	return err
}

// GetByID returns a specific project by its ID
func (p *ProjectModel) GetByID(id int) (*Project, error) {
	var project Project
	err := p.DB.QueryRow(context.Background(), `
		SELECT pr.id, pr.title, pr.description, pr.cover_image_id, m.thumbnail_url
		FROM projects pr
		LEFT JOIN media m ON pr.cover_image_id = m.id
		WHERE pr.id = $1
	`, id).Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.CoverImageID,
		&project.CoverImageURL,
	)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetMedia returns media linked to the project via the join table
func (p *ProjectModel) GetMedia(projectID int) ([]*Media, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT m.id, m.file_name, m.thumbnail_url, m.full_url
		FROM project_media pm
		JOIN media m ON pm.media_id = m.id
		WHERE pm.project_id = $1
		ORDER BY pm.media_id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		var m Media
		if err := rows.Scan(&m.ID, &m.FileName, &m.ThumbnailURL, &m.FullURL); err != nil {
			return nil, err
		}
		media = append(media, &m)
	}

	return media, nil
}

func (p *ProjectModel) GetAll() ([]map[string]interface{}, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT pr.id, pr.title, pr.description, pr.cover_image_id, pr.published, m.thumbnail_url
		FROM projects pr
		LEFT JOIN media m ON pr.cover_image_id = m.id
		ORDER BY pr.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var id int
		var title, description string
		var coverImageID *int
		var coverImageURL *string
		var published bool

		if err := rows.Scan(&id, &title, &description, &coverImageID, &published, &coverImageURL); err != nil {
			return nil, err
		}

		projects = append(projects, map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"Description":   description,
			"CoverImageID":  coverImageID,
			"CoverImageURL": coverImageURL,
			"Published":     published,
		})
	}
	return projects, nil
}

func (p *ProjectModel) Create(title, description string) error {
	_, err := p.DB.Exec(context.Background(),
		`INSERT INTO projects (title, description) VALUES ($1, $2)`,
		title, description,
	)
	return err
}

func (p *ProjectModel) SetCoverImage(projectID, mediaID int) error {
	_, err := p.DB.Exec(context.Background(),
		`UPDATE projects SET cover_image_id = $1 WHERE id = $2`,
		mediaID, projectID,
	)
	return err
}
