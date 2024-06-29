package users

import (
	"crypto/sha256"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// Generate the Hashed password for a user.
func genHashedPass(uuid uuid.UUID, password string, options *UserOptions) []byte {
	hash := sha256.New()
	salted := strings.Join([]string{uuid.String(), password, options.Salt}, "|")
	hash.Write([]byte(salted))
	return hash.Sum(nil)
}

// Set the password for a user.
func (u User) SetPassword(password string, options *UserOptions) User {
	u.HashedPass = genHashedPass(u.Uuid, password, options)
	return u
}

// Check if the password is correct for a user.
func (u User) CheckPassword(password string, options *UserOptions) bool {
	return slices.Equal(u.HashedPass, genHashedPass(u.Uuid, password, options))
}
