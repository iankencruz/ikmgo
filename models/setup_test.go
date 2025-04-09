package models

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if err := godotenv.Load("../.env"); err != nil {
		t.Log("⚠️  .env file not found, assuming env vars are set")
	}

	dbURL := os.Getenv("DB_URL")
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	_, err = db.Exec(context.Background(), `
        TRUNCATE users, media, gallery_media, project_media RESTART IDENTITY CASCADE;
    `)
	if err != nil {
		t.Fatalf("❌ Failed to truncate test tables: %v", err)
	}

	return db
}
