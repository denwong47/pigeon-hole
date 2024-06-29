package users

import (
	"strings"

	"github.com/google/uuid"
)

// User is the struct that represents a user in the system.
type User struct {
	Uuid       uuid.UUID  `json:"uuid" doc:"The unique identifier for the user. Currently unused."`
	Name       string     `json:"name" doc:"The name of the user."`
	Email      string     `json:"email" doc:"The email of the user. No emails will be sent; this is used as an unique identifier only."`
	HashedPass []byte     `json:"hashedPass" doc:"The hashed password of the user."`
	Privileges Privileges `json:"privileges" doc:"The privileges that this user has on objects."`
}

// New creates a new user with a new UUID and the specified privileges.
// This new user will not have a password set, and thus cannot be authenticated.
func NewUser(name string, email string, privileges Privileges) User {
	return UserWithUuid(
		uuid.New(),
		name,
		email,
		privileges,
	)
}

// Generate a user with the specified UUID.
// This new user will not have a password set, and thus cannot be authenticated.
func UserWithUuid(uuid uuid.UUID, name string, email string, privileges Privileges) User {
	return User{
		Uuid:       uuid,
		Name:       name,
		Email:      strings.ToLower(email),
		Privileges: privileges,
	}
}

// Returns `true` if the User is allowed to read objects.
func (u *User) CanSelect(isOwner bool) bool {
	return u.Privileges.All.Select || (isOwner && u.Privileges.Owned.Select)
}

// Returns `true` if the User is allowed insert objects.
// This supports the `isOwner` flag, but typically only the `Owned` privileges are
// since there is no mechanism for a user to insert data other than their own.
func (u *User) CanInsert(isOwner bool) bool {
	return u.Privileges.All.Insert || (isOwner && u.Privileges.Owned.Insert)
}

// Returns `true` if the User is allowed to update objects.
func (u *User) CanUpdate(isOwner bool) bool {
	return u.Privileges.All.Update || (isOwner && u.Privileges.Owned.Update)
}

// Returns `true` if the User is allowed to delete objects.
func (u *User) CanDelete(isOwner bool) bool {
	return u.Privileges.All.Delete || (isOwner && u.Privileges.Owned.Delete)
}
