package models

import "testing"

func TestGalleryModel_CRUD(t *testing.T) {
	db := setupTestDB(t)
	model := &GalleryModel{DB: db}

	cases := []struct {
		name        string
		title       string
		description string
		slug        string
		wantErr     bool
	}{
		{
			name:        "✅ valid gallery", // PASSING
			title:       "My Summer Gallery",
			description: "A lovely time at the beach",
			slug:        "summer-gallery",
			wantErr:     false,
		},
		{
			name:        "🟡 empty title", // OPTIONAL/EDGE
			title:       "",
			description: "Title is missing",
			slug:        "missing-title",
			wantErr:     false, // depends on your schema rules
		},
		{
			name:        "❌ missing slug", // FAILING
			title:       "Gallery With No Slug",
			description: "This should fail",
			slug:        "", // Slug should be required
			wantErr:     true,
		},
		{
			name:        "❌ duplicate slug",
			title:       "Another Nature Gallery",
			description: "Should conflict with earlier slug",
			slug:        "summer-gallery", // same as the first one
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.Create(tc.title, tc.description, tc.slug)

			if tc.wantErr {
				if err == nil {
					t.Errorf("❌ Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("❌ Unexpected error: %v", err)
				}
			}
		})
	}
}

// TESTING UPDATE FUNCTION
func TestGalleryModel_Update(t *testing.T) {
	db := setupTestDB(t)
	model := &GalleryModel{DB: db}

	cases := []struct {
		name    string  // descriptive name for t.Run()
		initial Gallery // initial gallery we insert
		updated Gallery // new values to apply in Update()
		wantErr bool    // do we expect Update() to fail?
	}{
		{
			name: "✅ successful update",
			initial: Gallery{
				Title:       "Initial Title",
				Description: "Initial Description",
				Slug:        "initial-slug",
			},
			updated: Gallery{
				Title:       "Updated Title",
				Description: "Updated Description",
				Slug:        "updated-slug",
			},
			wantErr: false,
		},
		{
			name: "❌ update with empty title",
			initial: Gallery{
				Title:       "Initial Title",
				Description: "Initial Description",
				Slug:        "initial-slug-2",
			},
			updated: Gallery{
				Title:       "",
				Description: "Updated Description",
				Slug:        "updated-slug-2",
			},
			wantErr: true,
		},
		{
			name: "❌ update with duplicate slug",
			initial: Gallery{
				Title:       "Initial Title",
				Description: "Initial Description",
				Slug:        "initial-slug-3",
			},
			updated: Gallery{
				Title:       "Another Title",
				Description: "Updated Description",
				Slug:        "initial-slug-2", // same as the initial slug
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Insert the initial gallery
			err := model.Create(tc.initial.Title, tc.initial.Description, tc.initial.Slug)
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}

			// Fetch the ID of the newly created gallery
			all, err := model.GetAll()
			if err != nil || len(all) == 0 {
				t.Fatal("Expected at least one gallery")
			}
			id := all[len(all)-1]["ID"].(int)

			// Update the gallery
			err = model.Update(id, tc.updated.Title, tc.updated.Description, tc.updated.Slug)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}

			// Fetch the updated gallery

			updated, err := model.GetByID(id)
			if err != nil {
				t.Fatalf("GetByID failed after update: %v", err)
			}

			if updated.Title != tc.updated.Title {
				t.Errorf("Expected title %q, got %q", tc.updated.Title, updated.Title)
			}
			if updated.Description != tc.updated.Description {
				t.Errorf("Expected description %q, got %q", tc.updated.Description, updated.Description)
			}
			if updated.Slug != tc.updated.Slug {
				t.Errorf("Expected slug %q, got %q", tc.updated.Slug, updated.Slug)
			}

		})
	}
}

// TESTING SETPUBLISH FUNCTION
func TestGalleryModel_SetPublish(t *testing.T) {
	db := setupTestDB(t)
	model := &GalleryModel{DB: db}

	// Insert a gallery to toggle
	err := model.Create("Publish Test", "Testing", "publish-test")
	if err != nil {
		t.Fatalf("failed to create gallery: %v", err)
	}

	// Fetch using slug

	gallery, err := model.GetBySlug("publish-test")
	if gallery == nil {
		t.Fatalf("failed to fetch gallery by slug")
	}

	// Test setting publish

	cases := []struct {
		name      string
		slug      string
		id        int
		published bool
		wantErr   bool
	}{
		{
			name:      "✅ set to published",
			slug:      gallery.Slug,
			id:        gallery.ID,
			published: true,
			wantErr:   false,
		},
		{
			name:      "✅ set to unpublished",
			slug:      gallery.Slug,
			published: false,
			id:        gallery.ID,
			wantErr:   false,
		},
		{
			name:      "❌ invalid ID",
			slug:      "invalid-slug",
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
					t.Errorf("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			if tc.slug == gallery.Slug {
				g, err := model.GetBySlug(tc.slug)
				if err != nil {
					t.Fatalf("GetBySlug failed: %v", err)
				}
				if g.Published != tc.published {
					t.Errorf("Expected published status %v, got %v", tc.published, g.Published)
				}
			}

		})
	}

}

// TESTING DELETE FUNCTION
func TestGalleryModel_Delete(t *testing.T) {
	db := setupTestDB(t)
	model := &GalleryModel{DB: db}

	// Insert a gallery to delete
	err := model.Create("Gallery to Delete", "This gallery will be deleted", "delete-me")
	if err != nil {
		t.Fatalf("failed to create gallery: %v", err)
	}

	// Fetch the ID of the newly created gallery
	gallery, err := model.GetBySlug("delete-me")
	if err != nil {
		t.Fatalf("failed to fetch gallery by slug: %v", err)
	}

	cases := []struct {
		name    string
		title   string
		slug    string
		id      int
		wantErr bool
	}{
		{
			name:    "✅ delete existing gallery",
			title:   "Gallery to Delete",
			slug:    "delete-me",
			id:      gallery.ID,
			wantErr: false,
		},
		{
			name:    "❌ delete non-existing gallery",
			title:   "Non-existing Gallery",
			slug:    "non-existing",
			id:      9999, // Assuming this ID doesn't exist
			wantErr: true,
		},
		{
			name:    "❌ delete gallery with invalid ID",
			title:   "Invalid ID Gallery",
			slug:    "invalid-id",
			id:      -1, // Invalid ID
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.Delete(tc.id)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !tc.wantErr {
				// Check if the gallery is actually deleted
				_, err := model.GetBySlug(tc.slug)
				if err == nil {
					t.Errorf("Gallery %q was not deleted, but it should have been", tc.slug)
				}
			}
		})
	}

}

// TESTING SETCOVERIMAGE FUNCTION
func TestGalleryModel_SetCoverImage(t *testing.T) {
	db := setupTestDB(t)
	model := &GalleryModel{DB: db}
	media := &MediaModel{DB: db}

	// Insert a gallery
	err := model.Create("Gallery", "desc", "slug-cover")
	if err != nil {
		t.Fatalf("create gallery failed: %v", err)
	}
	gallery, _ := model.GetBySlug("slug-cover")

	// Insert media
	mediaID, err := media.InsertAndReturnID("cover.jpg", "full.jpg", "thumb.jpg")
	if err != nil {
		t.Fatalf("insert media failed: %v", err)
	}

	cases := []struct {
		name      string
		galleryID int
		mediaID   int
		wantErr   bool
	}{
		{
			name:      "✅ set valid cover image",
			galleryID: gallery.ID,
			mediaID:   mediaID,
			wantErr:   false,
		},
		{
			name:      "❌ invalid gallery ID",
			galleryID: 9999,
			mediaID:   mediaID,
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.SetCoverImage(tc.galleryID, tc.mediaID)

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

			updated, _ := model.GetByID(tc.galleryID)
			if updated.CoverImageID == nil || *updated.CoverImageID != tc.mediaID {
				t.Errorf("Expected CoverImageID %d, got %v", tc.mediaID, updated.CoverImageID)
			}
		})
	}
}
