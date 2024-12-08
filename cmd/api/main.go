package main

import (
	"context"
	"html/template"
	"ikm/internal/handlers"
	"ikm/internal/models"
	"ikm/internal/render"
	"ikm/internal/session"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type application struct {
	logger         *log.Logger
	templateCache  map[string]*template.Template
	db             *pgxpool.Pool
	s3Client       *s3.S3
	galleryModel   *models.GalleryModel
	mediaModel     *models.MediaModel
	userModel      *models.UserModel
	sessionManager *session.Manager
}

func main() {
	ctx := context.Background()

	curDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	// Load environment variables from the .env file
	err = godotenv.Load(curDir + "/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Initialize database
	db, err := openDB(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Initialize template cache
	templateCache, err := render.NewTemplateCache()
	if err != nil {
		logger.Fatal(err)
	}

	// Log the contents of the template cache
	for name := range templateCache {
		logger.Printf("Template cached: %s", name)
	}

	// Initialize S3 client
	s3Client, err := initS3Client()
	if err != nil {
		logger.Fatal(err)
	}

	// Initialize media model
	mediaModel := models.NewMediaModel(db, s3Client, os.Getenv("S3_BUCKET"))

	// Initialize gallery model
	galleryModel := models.NewGalleryModel(db)

	// Initilize user model
	userModel := &models.UserModel{DB: db}

	sessionManager := session.NewManager()

	// Initialize application
	app := &application{
		logger:         logger,
		templateCache:  templateCache,
		db:             db,
		s3Client:       s3Client,
		mediaModel:     mediaModel,
		galleryModel:   galleryModel,
		userModel:      userModel,
		sessionManager: sessionManager,
	}

	// Create handlers instance
	handlers := handlers.New(app.logger, app.templateCache, app.db, app.galleryModel, app.mediaModel, app.userModel, sessionManager)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mainRouter(handlers, sessionManager),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting server on %s", srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(ctx context.Context) (*pgxpool.Pool, error) {
	// Build the database URL from environment variables
	dbURL := "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME")

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	// Set pool configuration
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 15 * time.Minute

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func initS3Client() (*s3.S3, error) {
	sess, err := awsSession.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("S3_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
		Endpoint:    aws.String(os.Getenv("S3_ENDPOINT")),
	})
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}
