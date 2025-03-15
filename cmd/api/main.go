package main

import (
	"context"
	"ikm/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Application struct {
	DB           *pgxpool.Pool
	UserModel    *models.UserModel
	GalleryModel *models.GalleryModel
	MediaModel   *models.MediaModel
	ContactModel *models.ContactModel

	// S3 configuration
	S3Client *minio.Client
	S3Bucket string
	S3Region string
}

func main() {

	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// S3 configuration
	// 1) Gather S3 config from environment
	s3Endpoint := os.Getenv("VULTR_S3_ENDPOINT")
	s3AccessKey := os.Getenv("VULTR_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("VULTR_S3_SECRET_KEY")
	s3Bucket := os.Getenv("VULTR_S3_BUCKET")
	s3Region := os.Getenv("VULTR_S3_REGION")

	// 2) Initialize MinIO S3 client
	s3Client, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3AccessKey, s3SecretKey, ""),
		Secure: true, // set to false if not using SSL
	})
	if err != nil {
		log.Fatalf("Unable to initialize S3 client: %v", err)
	}

	// Optionally ensure that the bucket exists; create if it doesn't:
	ctx := context.Background()
	exists, errBucketExists := s3Client.BucketExists(ctx, s3Bucket)
	if errBucketExists != nil {
		log.Fatalf("Error checking if S3 bucket exists: %v", errBucketExists)
	}
	if !exists {
		err = s3Client.MakeBucket(ctx, s3Bucket, minio.MakeBucketOptions{Region: s3Region})
		if err != nil {
			log.Fatalf("Unable to create S3 bucket: %v", err)
		}
		log.Printf("Created bucket %s\n", s3Bucket)
	}

	// Database connection
	dbURL := os.Getenv("DB_URL")
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Check if the connection is successful
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Unable to acquire a connection: %v", err)
	}
	defer conn.Release()

	var result int
	err = conn.QueryRow(context.Background(), "SELECT 1").Scan(&result)
	if err != nil || result != 1 {
		log.Fatalf("Database connection test failed: %v", err)
	}

	// Load all templates
	err = LoadTemplates()
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Initialize application struct
	app := &Application{
		DB:           dbPool,
		UserModel:    &models.UserModel{DB: dbPool},
		GalleryModel: &models.GalleryModel{DB: dbPool},
		MediaModel:   &models.MediaModel{DB: dbPool},
		ContactModel: &models.ContactModel{DB: dbPool},

		S3Client: s3Client,
		S3Bucket: s3Bucket,
	}

	// DebugRoutes(app.routes())

	// Start server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("Server running on port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
