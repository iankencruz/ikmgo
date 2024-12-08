// internal/models/user.go
package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int64
	Name           string
	Email          string
	HashedPassword string
	Role           string
	CreatedAt      time.Time
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m *UserModel) Insert(name, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// Use named parameters to insert the user
	args := pgx.NamedArgs{
		"name":            name,
		"email":           email,
		"hashed_password": string(hashedPassword),
		"role":            role,
	}

	query := `
        INSERT INTO users (name, email, hashed_password, role, created_at)
        VALUES (@name, @email, @hashed_password, @role, NOW())
    `
	_, err = m.DB.Exec(context.Background(), query, args)
	return err
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
        SELECT id, name, email, hashed_password, role, created_at 
        FROM users 
        WHERE email = @email
    `

	// Use named parameters to get the user by email
	args := pgx.NamedArgs{
		"email": email,
	}

	user := &User{}
	err := m.DB.QueryRow(context.Background(), query, args).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *UserModel) GetByID(id int64) (*User, error) {
	query := `
			SELECT id, name, email, role
			FROM users
			WHERE id = $1
	`
	row := m.DB.QueryRow(context.Background(), query, id)

	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role)
	if err != nil {
		return nil, err
	}

	return user, nil
}
