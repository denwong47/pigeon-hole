package tokens

import (
	"log"
	"sync"
	"time"

	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

// Token Manager. This is a singleton that manages the tokens.
type TokenManager struct {
	tokens     map[string]TokenData
	lock       sync.RWMutex
	expiration time.Duration
}

type TokenData struct {
	Token  string
	User   *users.User
	Expiry time.Time
}

// NewTokenManager creates a new TokenManager.
func NewTokenManager(expiration time.Duration) TokenManager {
	return TokenManager{
		tokens:     make(map[string]TokenData, 0),
		expiration: expiration,
	}
}

// CreateToken creates a new token that points to a user.
func (tm *TokenManager) CreateToken(user *users.User) (*TokenData, error) {
	token, err := GenerateToken(TokenLength)

	if err != nil {
		return &TokenData{}, err
	}

	tm.lock.Lock()
	defer tm.lock.Unlock()

	tokenData := TokenData{
		Token:  token,
		User:   user,
		Expiry: time.Now().Add(tm.expiration),
	}
	tm.tokens[token] = tokenData

	return &tokenData, nil
}

// GetUser retrieves the `TokenData“ from the token.
func (tm *TokenManager) GetToken(token string) (*TokenData, error) {
	tm.lock.RLock()
	defer tm.lock.RUnlock()

	if data, ok := tm.tokens[token]; ok {
		log.Printf("Token %s found for user %s, expiring at %s.\n", token, data.User.Name, data.Expiry)
		if data.Expiry.Before(time.Now()) {
			return &TokenData{}, errorMessages.ErrTokenExpired
		}
		return &data, nil
	}

	return &TokenData{}, errorMessages.ErrTokenInvalid
}

// DeleteToken removes the token from the manager.
func (tm *TokenManager) DeleteToken(token string) error {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	if _, ok := tm.tokens[token]; ok {
		delete(tm.tokens, token)
		return nil
	}

	return errorMessages.ErrTokenInvalid
}
