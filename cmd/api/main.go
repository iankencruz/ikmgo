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
)

type Application struct {
	DB           *pgxpool.Pool
	UserModel    *models.UserModel
	GalleryModel *models.GalleryModel
	MediaModel   *models.MediaModel
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
	}

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
