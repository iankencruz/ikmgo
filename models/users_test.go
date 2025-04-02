package models

import (
	"testing"
)

func TestUserModel_CreateAndAuthenticate(t *testing.T) {
	db := setupTestDB(t)
	model := &UserModel{DB: db}

	cases := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{"create valid", "test@example.com", "password123", false},
		{"duplicate email", "test@example.com", "password123", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := model.Create("First", "Last", tc.email, tc.password)
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	t.Run("authenticate success", func(t *testing.T) {
		_, err := model.Authenticate("test@example.com", "password123")
		if err != nil {
			t.Errorf("Authenticate() failed: %v", err)
		}
	})

	t.Run("authenticate fail", func(t *testing.T) {
		_, err := model.Authenticate("test@example.com", "wrongpass")
		if err == nil {
			t.Error("expected failure for wrong password")
		}
	})
}
