package models

import (
	"testing"
)

func TestMediaModel_InsertAndFetch(t *testing.T) {
	db := setupTestDB(t)
	model := &MediaModel{DB: db}

	cases := []struct {
		name     string
		fileName string
		fullURL  string
		thumbURL string
	}{
		{
			name:     "basic image insert",
			fileName: "image.jpg",
			fullURL:  "https://example.com/full.jpg",
			thumbURL: "https://example.com/thumb.jpg",
		},
		{
			name:     "png insert",
			fileName: "test.png",
			fullURL:  "https://cdn.com/image.png",
			thumbURL: "https://cdn.com/thumb.png",
		},
		{
			name:     "gif insert",
			fileName: "test.gif",
			fullURL:  "https://cdn.com/image.gif",
			thumbURL: "https://cdn.com/thumb.gif",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := model.InsertAndReturnID(tc.fileName, tc.fullURL, tc.thumbURL)
			if err != nil {
				t.Fatalf("InsertAndReturnID failed: %v", err)
			}

			media, err := model.GetByID(id)
			if err != nil {
				t.Fatalf("GetByID failed: %v", err)
			}

			if media.FileName != tc.fileName {
				t.Errorf("Expected filename %q, got %q", tc.fileName, media.FileName)
			}

			if media.FullURL != tc.fullURL {
				t.Errorf("Expected full URL %q, got %q", tc.fullURL, media.FullURL)
			}

			if media.ThumbnailURL != tc.thumbURL {
				t.Errorf("Expected thumbnail URL %q, got %q", tc.thumbURL, media.ThumbnailURL)
			}
		})
	}
}
