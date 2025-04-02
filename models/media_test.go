package models

import (
	"testing"
)

func TestMediaModel_InsertAndFetch(t *testing.T) {
	db := setupTestDB(t)
	model := &MediaModel{DB: db}

	id, err := model.InsertAndReturnID("image.jpg", "https://full.url", "https://thumb.url")
	if err != nil {
		t.Fatalf("InsertAndReturnID failed: %v", err)
	}

	t.Run("get by ID", func(t *testing.T) {
		media, err := model.GetByID(id)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if media.FileName != "image.jpg" {
			t.Errorf("unexpected filename: %s", media.FileName)
		}
	})
}
