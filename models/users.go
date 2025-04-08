package models

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Password  string
}
type UserModel struct {
	DB *pgxpool.Pool
}

// Create inserts a new user with a hashed password
func (u *UserModel) Create(fname, lname, email, password string) error {
	// Check if user already exists
	var exists bool
	err := u.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user
	_, err = u.DB.Exec(context.Background(),
		"INSERT INTO users (fname, lname, email, password) VALUES ($1, $2, $3, $4)",
		fname, lname, email, string(hashedPassword))
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
func (u *UserModel) GetUserByID(userID int) (*User, error) {
	var user User
	err := u.DB.QueryRow(context.Background(),
		"SELECT id, fname, lname, email FROM users WHERE id=$1", userID).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
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

func (u *UserModel) GetAll() ([]User, error) {
	rows, err := u.DB.Query(context.Background(), "SELECT id, fname, lname, email FROM users ORDER BY id ASC")
	if err != nil {
		log.Printf("❌ Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			return nil, err
		}
		users = append(users, user)
	}

	log.Printf("✅ Retrieved %d users", len(users))
	return users, nil
}

func (u *UserModel) Update(id int, fname, lname, email string) error {
	_, err := u.DB.Exec(context.Background(),
		"UPDATE users SET fname=$1, lname=$2, email=$3 WHERE id=$4",
		fname, lname, email, id)
	return err
}

// Get Count
func (u *UserModel) Count() (int, error) {
	var count int
	err := u.DB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("❌ Error counting users: %v", err)
		return 0, err
	}
	log.Printf("✅ Total users: %d", count)
	return count, nil
}
