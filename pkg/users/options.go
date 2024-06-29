package users

import "time"

// UserOptions is the struct that contains options for user related operations.
type UserOptions struct {
	Salt            string        `json:"salt" doc:"The salt used to hash the user's password."`
	TokenExpiration time.Duration `json:"tokenExpiration" doc:"The duration that a token is valid for."`
}
