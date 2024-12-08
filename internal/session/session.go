package session

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// Manager struct to manage sessions
type Manager struct {
	store *sessions.CookieStore
}

// NewManager creates a new session manager with a secure cookie store
func NewManager() *Manager {
	// Replace with a secure, randomly generated key
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		panic("SESSION_KEY environment variable is not set!")
	}

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	return &Manager{store: store}
}

// Get retrieves a session by name
func (m *Manager) Get(r *http.Request, name string) (*sessions.Session, error) {
	return m.store.Get(r, name)
}

// Save saves the current session to the response
func (m *Manager) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return session.Save(r, w)
}

// Destroy deletes a session by name
func (m *Manager) Destroy(r *http.Request, w http.ResponseWriter, name string) error {
	session, err := m.store.Get(r, name)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1 // Marks the session as expired
	return session.Save(r, w)
}
