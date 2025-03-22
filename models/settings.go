package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Setting struct {
	Key   string
	Value string
}

type SettingsModel struct {
	DB *pgxpool.Pool
}

// Get a setting by key
func (s *SettingsModel) Get(key string) (string, error) {
	var value string
	err := s.DB.QueryRow(context.Background(), "SELECT value FROM settings WHERE key=$1", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// Update or insert a setting
func (s *SettingsModel) Set(key, value string) error {
	_, err := s.DB.Exec(context.Background(),
		"INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value",
		key, value)
	return err
}

// Get all settings as a map
func (s *SettingsModel) GetAll() (map[string]string, error) {
	rows, err := s.DB.Query(context.Background(), "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, nil
}
