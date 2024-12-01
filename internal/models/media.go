// ./internal/models/media.go
package models

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	CreatedAt time.Time
	ImageURL  string
	ID        int64
	GalleryID int64
}

type MediaModel struct {
	DB       *pgxpool.Pool
	S3       *s3.S3
	S3Bucket string
}

// NewMediaModel creates and returns a new instance of MediaModel
func NewMediaModel(db *pgxpool.Pool, s3Client *s3.S3, s3Bucket string) *MediaModel {
	return &MediaModel{
		DB:       db,
		S3:       s3Client,
		S3Bucket: s3Bucket,
	}
}

func (m *MediaModel) UploadMedia(galleryID int64, imagePath, s3Key string) (int64, error) {
	// Open the file for reading
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open file %s: %v", imagePath, err)
	}
	defer file.Close()

	// Upload image to S3 using an uploader
	uploader := s3manager.NewUploaderWithClient(m.S3)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: &m.S3Bucket,
		Key:    &s3Key,
		Body:   file,
	})
	if err != nil {
		return 0, fmt.Errorf("unable to upload file to S3: %v", err)
	}

	// Save media information to the database using pgxpool
	query := `
		INSERT INTO media (gallery_id, image_url, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`

	var mediaID int64
	err = m.DB.QueryRow(context.Background(), query, galleryID, s3Key).Scan(&mediaID)
	if err != nil {
		return 0, fmt.Errorf("unable to save media information to database: %v", err)
	}

	return mediaID, nil
}
