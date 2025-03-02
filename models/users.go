package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Email    string
	Password string
}

type UserModel struct {
	DB *pgxpool.Pool
}

// Create inserts a new user with a hashed password
func (u *UserModel) Create(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = u.DB.Exec(context.Background(),
		"INSERT INTO users (email, password) VALUES ($1, $2)", email, string(hashedPassword))
	return err
}

// Authenticate verifies user credentials and returns the user if successful
func (u *UserModel) Authenticate(email, password string) (*User, error) {
	var user User
	err := u.DB.QueryRow(context.Background(),
		"SELECT id, email, password FROM users WHERE email=$1", email).
		Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Compare stored hash with entered password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

// GetByID retrieves a user by their ID
func (u *UserModel) GetByID(id int) (*User, error) {
	var user User
	err := u.DB.QueryRow(context.Background(),
		"SELECT id, email FROM users WHERE id=$1", id).
		Scan(&user.ID, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdatePassword updates a user's password
func (u *UserModel) UpdatePassword(id int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = u.DB.Exec(context.Background(),
		"UPDATE users SET password=$1 WHERE id=$2", string(hashedPassword), id)
	return err
}

// Delete removes a user from the database
func (u *UserModel) Delete(id int) error {
	_, err := u.DB.Exec(context.Background(), "DELETE FROM users WHERE id=$1", id)
	return err
}
