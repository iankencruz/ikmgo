package models

import (
	"testing"
)

func TestProjectModel_Create(t *testing.T) {
	db := setupTestDB(t)
	model := &ProjectModel{DB: db}

	cases := []struct {
		name        string
		title       string
		description string
		slug        string
		wantErr     bool
	}{
		{
			name:        "✅ valid project",
			title:       "Project Alpha",
			description: "Intro project",
			slug:        "project-alpha",
			wantErr:     false,
		},
		{
			name:        "🟡 empty title",
			title:       "",
			description: "Missing title",
			slug:        "no-title",
			wantErr:     false,
		},
		{
			name:        "❌ missing slug",
			title:       "No Slug",
			description: "Missing slug",
			slug:        "",
			wantErr:     true,
		},
		{
			name:        "❌ duplicate slug",
			title:       "Duplicate Slug",
			description: "Second project",
			slug:        "project-alpha",
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.Create(tc.title, tc.description, tc.slug)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestProjectModel_Update(t *testing.T) {
	db := setupTestDB(t)
	model := &ProjectModel{DB: db}

	err := model.Create("Original", "Desc", "original-slug")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	project, _ := model.GetBySlug("original-slug")

	cases := []struct {
		name    string
		updated Project
		wantErr bool
	}{
		{
			name: "✅ successful update",
			updated: Project{
				Title:       "Updated",
				Description: "Updated desc",
				Slug:        "updated-slug",
			},
			wantErr: false,
		},
		{
			name: "❌ update with empty slug",
			updated: Project{
				Title:       "Updated",
				Description: "Updated desc",
				Slug:        "",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.UpdateBasicInfo(project.ID, tc.updated.Title, tc.updated.Description, tc.updated.Slug)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			p, err := model.GetByID(project.ID)
			if err != nil {
				t.Fatalf("GetByID failed: %v", err)
			}
			if p.Slug != tc.updated.Slug {
				t.Errorf("Expected slug %q, got %q", tc.updated.Slug, p.Slug)
			}
		})
	}
}

func TestProjectModel_SetPublished(t *testing.T) {
	db := setupTestDB(t)
	model := &ProjectModel{DB: db}

	err := model.Create("Publish Me", "desc", "publish-me")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	project, _ := model.GetBySlug("publish-me")

	cases := []struct {
		name      string
		id        int
		published bool
		wantErr   bool
	}{
		{
			name:      "✅ publish project",
			id:        project.ID,
			published: true,
			wantErr:   false,
		},
		{
			name:      "❌ invalid project ID",
			id:        9999,
			published: true,
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.SetPublished(tc.id, tc.published)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			p, _ := model.GetByID(tc.id)
			if p.Published != tc.published {
				t.Errorf("Expected published=%v, got %v", tc.published, p.Published)
			}
		})
	}
}

func TestProjectModel_Delete(t *testing.T) {
	db := setupTestDB(t)
	model := &ProjectModel{DB: db}

	err := model.Create("Delete Me", "desc", "delete-me")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	project, _ := model.GetBySlug("delete-me")

	cases := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "✅ delete existing project",
			id:      project.ID,
			wantErr: false,
		},
		{
			name:    "❌ delete invalid ID",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.Delete(tc.id)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestProjectModel_SetCoverImage(t *testing.T) {
	db := setupTestDB(t)
	model := &ProjectModel{DB: db}
	media := &MediaModel{DB: db}

	// Create a project
	err := model.Create("Cover Project", "desc", "cover-project")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	project, _ := model.GetBySlug("cover-project")

	// Insert media
	mediaID, err := media.InsertAndReturnID("cover.jpg", "full.jpg", "thumb.jpg")
	if err != nil {
		t.Fatalf("Insert media failed: %v", err)
	}

	cases := []struct {
		name      string
		projectID int
		mediaID   int
		wantErr   bool
	}{
		{
			name:      "✅ set valid cover image",
			projectID: project.ID,
			mediaID:   mediaID,
			wantErr:   false,
		},
		{
			name:      "❌ invalid project ID",
			projectID: 9999,
			mediaID:   mediaID,
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.SetCoverImage(tc.projectID, tc.mediaID)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			p, err := model.GetByID(tc.projectID)
			if err != nil {
				t.Fatalf("GetByID failed: %v", err)
			}

			if p.CoverImageID == nil || *p.CoverImageID != tc.mediaID {
				t.Errorf("Expected CoverImageID %d, got %v", tc.mediaID, p.CoverImageID)
			}
		})
	}
}
