package auth

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"sync"

	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
	"github.com/denwong47/pigeon-hole/pkg/tokens"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

// AuthManager is a struct that represents a list of users.
type AuthManager struct {
	Name      string                 `json:"name" doc:"The name of the list."`
	Timestamp time.Time              `json:"timestamp" doc:"The time this list was updated."`
	Users     map[string]*users.User `json:"users" doc:"The list of users."`
	Tokens    tokens.TokenManager    `json:"-"`
	Options   users.UserOptions      `json:"-"`
	lock      sync.RWMutex
}

// NewAuthManger creates a new AuthManager with an empty list of users.
func NewAuthManager(name string, options users.UserOptions) *AuthManager {
	return &AuthManager{
		Name:      name,
		Timestamp: time.Now().UTC(),
		Users:     make(map[string]*users.User, 0),
		Tokens:    tokens.NewTokenManager(options.TokenExpiration),
		Options:   options,
	}
}

// UpdateTimestamp updates the timestamp of the AuthManager.
func (ul *AuthManager) UpdateTimestamp() {
	ul.Timestamp = time.Now().UTC()
}

// Length returns the number of users in the list.
func (ul *AuthManager) Length() int {
	return len(ul.Users)
}

// ImportFrom reads the user list from a file.
func ImportFrom(path string, options users.UserOptions) (*AuthManager, error) {
	buffer, err := os.ReadFile(path)

	if err != nil {
		return &AuthManager{}, errorMessages.ErrUserFileNotFound
	}

	var ul AuthManager
	if err := json.Unmarshal(buffer, &ul); err != nil {
		return &AuthManager{}, err
	}

	// Reinitialize the options, which would not be serialized.
	ul.Options = options
	// Reinitialize the lock, which would not be serialized.
	ul.lock = sync.RWMutex{}
	// Reinitialize the token manager, which would not be serialized.
	ul.Tokens = tokens.NewTokenManager(ul.Options.TokenExpiration)

	return &ul, nil
}

// ImportFromOrNew reads the user list from a file, or creates a new one if the file does not exist.
func ImportFromOrNew(path string, name string, options users.UserOptions) (*AuthManager, error) {
	if authManager, err := ImportFrom(path, options); err != nil {
		if errorMessages.Matches(err, errorMessages.ErrUserFileNotFound) {
			// The file does not exist, create a new user list
			return NewAuthManager(name, options), nil
		} else {
			// Something else went wrong, return the error
			return &AuthManager{}, err
		}
	} else {
		// The file was read successfully
		return authManager, nil
	}
}

// ExportTo writes the user list to a file.
func (ul *AuthManager) ExportTo(path string) error {
	ul.lock.RLock()
	defer ul.lock.RUnlock()

	ul.UpdateTimestamp()

	buffer, err := json.MarshalIndent(ul, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("Writing %d bytes of user data to %s...\n", len(buffer), path)
	// Make sure the file is only readable by the current user
	os.WriteFile(path, buffer, 0700)

	return nil
}

// Add the user to the list.
func (ul *AuthManager) AddUser(user *users.User) (*users.User, error) {
	ul.lock.Lock()
	defer ul.lock.Unlock()

	if existing, ok := ul.Users[user.Email]; ok {
		return existing, errorMessages.ErrUserAlreadyExists
	}

	ul.Users[user.Email] = user

	return user, nil
}

// Remove the user from the list.
func (ul *AuthManager) RemoveUser(email string) (*users.User, error) {
	ul.lock.Lock()
	defer ul.lock.Unlock()
	email = strings.ToLower(email)

	user, ok := ul.Users[email]

	if !ok {
		return &users.User{}, errorMessages.ErrUserNotFound
	}

	delete(ul.Users, email)

	return user, nil
}

// Get the user from the list.
func (ul *AuthManager) GetUser(email string) (*users.User, error) {
	ul.lock.RLock()
	defer ul.lock.RUnlock()
	email = strings.ToLower(email)

	if user, ok := ul.Users[email]; !ok {
		return nil, errorMessages.ErrUserNotFound
	} else {
		return user, nil
	}
}
