package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Contact struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Subject   string
	Message   string
	CreatedAt time.Time
}

type ContactModel struct {
	DB *pgxpool.Pool
}

func (m *ContactModel) Insert(firstName, lastName, email, subject, message string) error {
	_, err := m.DB.Exec(context.Background(),
		`INSERT INTO contacts (first_name, last_name, email, subject, message) 
		 VALUES ($1, $2, $3, $4, $5)`,
		firstName, lastName, email, subject, message,
	)
	return err
}

func (m *ContactModel) GetAll() ([]*Contact, error) {
	rows, err := m.DB.Query(context.Background(),
		`SELECT id, first_name, last_name, email, subject, message, created_at
		 FROM contacts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		c := &Contact{}
		err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Subject, &c.Message, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}
