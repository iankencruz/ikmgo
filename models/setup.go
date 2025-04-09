package models

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTablesIfNotExist(db *pgxpool.Pool) error {
	ctx := context.Background()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			fname TEXT NOT NULL,
			lname TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS media (
			id SERIAL PRIMARY KEY,
			file_name TEXT NOT NULL,
			full_url TEXT NOT NULL,
			thumbnail_url TEXT,
			embed_url TEXT,
			mime_type TEXT,
			position INTEGER DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS projects (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			slug TEXT UNIQUE NOT NULL,
			published BOOLEAN DEFAULT FALSE,
			cover_image_id INTEGER REFERENCES media(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS galleries (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			slug TEXT UNIQUE NOT NULL,
			published BOOLEAN DEFAULT FALSE,
			featured BOOLEAN DEFAULT FALSE,
			cover_image_id INTEGER REFERENCES media(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS gallery_media (
			gallery_id INTEGER REFERENCES galleries(id) ON DELETE CASCADE,
			media_id INTEGER REFERENCES media(id) ON DELETE CASCADE,
			position INTEGER DEFAULT 0,
			PRIMARY KEY (gallery_id, media_id)
		);`,

		`CREATE TABLE IF NOT EXISTS project_media (
			project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
			media_id INTEGER REFERENCES media(id) ON DELETE CASCADE,
			position INTEGER DEFAULT 0,
			PRIMARY KEY (project_id, media_id)
		);`,

		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS contacts (
			id SERIAL PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL,
			subject TEXT,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(ctx, stmt)
		if err != nil {
			log.Printf("❌ Failed to create table: %v", err)
			return err
		}
	}

	log.Println("✅ All tables checked/created successfully.")
	return nil
}

func EnsureAdminUserExists(userModel *UserModel) error {
	const (
		defaultFname = "Admin"
		defaultLname = "User"
	)

	log.Println("🔍 Checking for existing users...")
	users, err := userModel.GetAll()
	if err != nil {
		return err
	}

	if len(users) == 0 {
		log.Println("⚠️ No users found. Creating default admin user...")

		err := userModel.Create(defaultFname, defaultLname, os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASS"))
		if err != nil {
			log.Printf("❌ Failed to create default admin user: %v", err)
			return err
		}

		log.Print(os.Getenv("ADMIN_EMAIL"))
		log.Print("✅ Default admin user created")
	} else {
		log.Printf("✅ %d user(s) already exist. Skipping admin bootstrap.", len(users))
	}

	return nil
}
