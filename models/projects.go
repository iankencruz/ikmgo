package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID            int
	Title         string
	Slug          string
	Description   string
	CoverImageID  *int
	CoverImageURL *string
	MediaCount    int // ✅ Count of Media Items
	Published     bool
}

type ProjectModel struct {
	DB *pgxpool.Pool
}

// GetAllPublic returns all published projects with optional cover image

func (p *ProjectModel) GetAllPublic() ([]*Project, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT pr.id, pr.title, pr.slug, pr.description, pr.cover_image_id, m.thumbnail_url
		FROM projects pr
		LEFT JOIN media m ON pr.cover_image_id = m.id
		WHERE pr.published = TRUE
		ORDER BY pr.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.Title, &project.Slug, &project.Description, &project.CoverImageID, &project.CoverImageURL)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func (p *ProjectModel) SetPublished(id int, published bool) error {
	res, err := p.DB.Exec(context.Background(),
		"UPDATE projects SET published = $1 WHERE id = $2", published, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no project found with ID %d", id)
	}
	return nil
}

// GetByID returns a specific project by its ID
func (p *ProjectModel) GetByID(id int) (*Project, error) {
	var project Project
	err := p.DB.QueryRow(context.Background(), `
		SELECT
		  pr.id,
		  pr.title,
		  pr.slug,
		  pr.description,
		  pr.published,
		  pr.cover_image_id,
		  m.thumbnail_url,
		  (SELECT COUNT(*) FROM project_media WHERE project_id = pr.id) as media_count
		FROM projects pr
		LEFT JOIN media m ON pr.cover_image_id = m.id
		WHERE pr.id = $1

	`, id).Scan(
		&project.ID,
		&project.Title,
		&project.Slug,
		&project.Description,
		&project.Published,
		&project.CoverImageID,
		&project.CoverImageURL,
		&project.MediaCount,
	)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetMedia returns media linked to the project via the join table

func (p *ProjectModel) GetMediaPaginated(projectID, limit, offset int) ([]*Media, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT m.id, m.file_name, m.thumbnail_url, m.full_url, m.mime_type, m.embed_url
		FROM project_media pm
		JOIN media m ON pm.media_id = m.id
		WHERE pm.project_id = $1
		ORDER BY pm.position ASC
		LIMIT $2 OFFSET $3
	`, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*Media
	for rows.Next() {
		var m Media
		var embed pgtype.Text
		var mime pgtype.Text

		if err := rows.Scan(&m.ID, &m.FileName, &m.ThumbnailURL, &m.FullURL, &mime, &embed); err != nil {
			return nil, err
		}

		if embed.Valid {
			m.EmbedURL = &embed.String
		}

		if mime.Valid {
			m.MimeType = &mime.String
		}

		media = append(media, &m)
	}

	return media, nil
}

func (p *ProjectModel) GetAll() ([]map[string]interface{}, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT pr.id, pr.title, pr.slug, pr.description, pr.cover_image_id, pr.published, m.thumbnail_url
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
		var title, description, slug string
		var coverImageID *int
		var coverImageURL *string
		var published bool

		if err := rows.Scan(&id, &title, &slug, &description, &coverImageID, &published, &coverImageURL); err != nil {
			return nil, err
		}

		projects = append(projects, map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"Slug":          slug,
			"Description":   description,
			"CoverImageID":  coverImageID,
			"CoverImageURL": coverImageURL,
			"Published":     published,
		})
	}
	return projects, nil
}

func (p *ProjectModel) Create(title, description, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	_, err := p.DB.Exec(context.Background(),
		`INSERT INTO projects (title, description, slug) VALUES ($1, $2, $3)`,
		title, description, slug,
	)
	return err
}

func (p *ProjectModel) SetCoverImage(projectID, mediaID int) error {
	res, err := p.DB.Exec(context.Background(),
		`UPDATE projects SET cover_image_id = $1 WHERE id = $2`,
		mediaID, projectID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no project found with ID %d", projectID)
	}
	return nil
}

func (m *ProjectModel) UpdateBasicInfo(id int, title, description, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return fmt.Errorf("slug cannot be empty")
	}

	res, err := m.DB.Exec(context.Background(), `
		UPDATE projects SET title=$1, description=$2, slug=$3 WHERE id=$4
	`, title, description, slug, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no project found with ID %d", id)
	}
	return nil
}

// Get Count of Projects
func (p *ProjectModel) Count() (int, error) {
	var count int
	err := p.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM projects
	`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get 5 latest projects
func (p *ProjectModel) GetLatest(limit int) ([]map[string]interface{}, error) {
	rows, err := p.DB.Query(context.Background(), `
		SELECT pr.id, pr.title, pr.slug, pr.description, pr.cover_image_id, m.thumbnail_url
		FROM projects pr
		LEFT JOIN media m ON pr.cover_image_id = m.id
		WHERE pr.published = TRUE
		ORDER BY pr.id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var id int
		var title, description, slug string
		var coverImageID *int
		var coverImageURL *string

		if err := rows.Scan(&id, &title, &slug, &description, &coverImageID, &coverImageURL); err != nil {
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

func (p *ProjectModel) GetBySlug(slug string) (*Project, error) {
	var project Project
	err := p.DB.QueryRow(context.Background(), `
		SELECT id, title, slug, description, cover_image_id, published
		FROM projects WHERE slug = $1`, slug).Scan(
		&project.ID, &project.Title, &project.Slug, &project.Description, &project.CoverImageID, &project.Published,
	)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (p *ProjectModel) Delete(id int) error {
	res, err := p.DB.Exec(context.Background(),
		`DELETE FROM projects WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no project found with ID %d", id)
	}
	return nil
}
