package interfaces

import (
	"context"
	"fmt"
	"log"

	"github.com/danielgtaylor/huma/v2"

	"github.com/denwong47/pigeon-hole/pkg/auth"
	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
	keyValue "github.com/denwong47/pigeon-hole/pkg/key_value"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

// Add a user to the current active user list.
func AddUser(
	ctx context.Context,
	authManager *auth.AuthManager,
	input *AddUserRequest,
) (*AddUserResponse, error) {
	privileges, err := users.GetPrivilegesByType(input.Body.Type)

	if err != nil {
		return &AddUserResponse{}, huma.Error400BadRequest(fmt.Sprintf("Unknown user type `%s`.", input.Body.Type), err)
	}

	user := users.NewUser(input.Body.Name, input.Body.Email, privileges).SetPassword(input.Body.Password, &authManager.Options)

	if _, err := authManager.AddUser(&user); err != nil {
		return &AddUserResponse{}, huma.Error400BadRequest(fmt.Sprintf("Failed to add user '%s'.", input.Body.Name), err)
	} else if tokenData, err := authManager.Tokens.CreateToken(&user); err != nil {
		return &AddUserResponse{}, huma.Error500InternalServerError("Failed to create token for user.", err)
	} else {
		log.Printf("Added user '%s' with UUID `%s` and HashedPass `%s`, user list %s now has %d users.\n", user.Name, user.Uuid, user.HashedPass, authManager.Name, authManager.Length())
		return &AddUserResponse{Body: TokenResponseBody{
			Token:  tokenData.Token,
			Expiry: tokenData.Expiry,
		}}, nil
	}
}

// Remove a user from the current active user list.
func RemoveUser(
	ctx context.Context,
	authManager *auth.AuthManager,
	input *RemoveUserRequest,
) (*RemoveUserResponse, error) {
	if user, err := authManager.RemoveUser(input.Email); err != nil {
		return &RemoveUserResponse{}, huma.Error400BadRequest(fmt.Sprintf("Failed to remove user '%s'.", input.Email), err)
	} else {
		log.Printf("Removed user '%s' with UUID `%s`, user list %s now has %d users.\n", user.Name, user.Uuid, authManager.Name, authManager.Length())
		return &RemoveUserResponse{}, nil
	}
}

// For an existing user, login with password and return a token.
func LoginUser(
	ctx context.Context,
	authManager *auth.AuthManager,
	input *LoginUserRequest,
) (*LoginUserResponse, error) {
	user, err := authManager.GetUser(input.Body.Email)

	if err != nil {
		log.Printf("Failed to find user '%s'.\n", input.Body.Email)
		return &LoginUserResponse{}, huma.Error401Unauthorized(fmt.Sprintf("Failed to find user '%s'.", input.Body.Email), err)
	}

	if !user.CheckPassword(input.Body.Password, &authManager.Options) {
		log.Printf("User '%s' (%s) failed to authenticate, password mismatch.\n", user.Name, user.Email)
		return &LoginUserResponse{}, huma.Error401Unauthorized(fmt.Sprintf("Failed to authenticate user '%s'.", input.Body.Email), errorMessages.ErrAuthenticationFailed)
	}

	if tokenData, err := authManager.Tokens.CreateToken(user); err != nil {
		return &LoginUserResponse{}, huma.Error500InternalServerError("Failed to create token for user.", err)
	} else {
		log.Printf("User '%s' (%s) logged in successfully.\n", user.Name, user.Email)
		return &LoginUserResponse{Body: TokenResponseBody{
			Token:  tokenData.Token,
			Expiry: tokenData.Expiry,
		}}, nil
	}
}

// LogoutUser logs out the user by removing the token.
func LogoutUser(
	ctx context.Context,
	authManager *auth.AuthManager,
	input *LogoutUserRequest,
) (*LogoutUserResponse, error) {
	token, token_ok := GetTokenFromContext(ctx)
	user, user_ok := GetUserFromContext(ctx)
	if token_ok && user_ok {
		if err := authManager.Tokens.DeleteToken(token); err != nil {
			return &LogoutUserResponse{}, huma.Error500InternalServerError("Failed to delete token for user.", err)
		} else {
			log.Printf("User '%s' (%s) logged out successfully.\n", user.Name, user.Email)
			return &LogoutUserResponse{
				Body: struct{}{},
			}, nil
		}
	} else if token_ok {
		log.Printf("Cannot logout without token `%s`.\n", token)
		return &LogoutUserResponse{}, huma.Error401Unauthorized("Invalid token provided.", errorMessages.ErrTokenInvalid)
	} else {
		return &LogoutUserResponse{}, huma.Error401Unauthorized("Cannot logout without authorisation token.", errorMessages.ErrUnauthorized)
	}
}

// GetUserPermission returns the permission level of the user.
func GetUserPermission(
	ctx context.Context,
	authManager *auth.AuthManager,
	input *GetUserPermissionRequest,
) (*GetUserPermissionResponse, error) {
	_, token_ok := GetTokenFromContext(ctx)
	user, user_ok := GetUserFromContext(ctx)

	if !token_ok {
		return &GetUserPermissionResponse{}, huma.Error401Unauthorized("Cannot get permission without authorisation token.", errorMessages.ErrUnauthorized)
	}
	if !user_ok {
		return &GetUserPermissionResponse{}, huma.Error401Unauthorized("Cannot get permission without user context.", errorMessages.ErrUnauthorized)
	}

	if user.Email == "" {
		return &GetUserPermissionResponse{}, huma.Error401Unauthorized("Cannot get permission without valid user.", errorMessages.ErrUnauthorized)
	}

	return &GetUserPermissionResponse{
		Body: user.Privileges,
	}, nil
}

// GetKey retrieves the value for a given key.
func GetKey(
	ctx context.Context,
	authManager *auth.AuthManager,
	kvc *keyValue.KeyValueCache,
	input *GetKeyRequest,
) (*GetKeyResponse, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return &GetKeyResponse{}, huma.Error401Unauthorized("Authentication failed.", errorMessages.ErrUnauthorized)
	}

	if delivery, err := kvc.GetValue(input.Key, user); err != nil {
		if errorMessages.Matches(err, errorMessages.ErrKeyNotFound) {
			return &GetKeyResponse{}, huma.Error404NotFound(fmt.Sprintf("Failed to find key '%s'.", input.Key), err)
		} else if errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
			return &GetKeyResponse{}, huma.Error403Forbidden(fmt.Sprintf("User '%s' is not permitted to access key '%s'.", user.Name, input.Key), err)
		} else {
			return &GetKeyResponse{}, huma.Error400BadRequest(fmt.Sprintf("Cannot retrieve key '%s'.", input.Key), err)
		}
	} else {
		log.Printf("User '%s' (%s) retrieved key '%s'.\n", user.Name, user.Email, input.Key)
		return &GetKeyResponse{Body: delivery.Value}, nil
	}
}

// PutKey adds a new key to the cache.
func PutKey(
	ctx context.Context,
	authManager *auth.AuthManager,
	kvc *keyValue.KeyValueCache,
	input *PutKeyRequest,
) (*PutKeyResponse, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return &PutKeyResponse{}, huma.Error401Unauthorized("Authentication failed.", errorMessages.ErrUnauthorized)
	}

	if err := kvc.PutValue(input.Key, input.RawBody, user); err != nil {
		if errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
			return &PutKeyResponse{}, huma.Error403Forbidden(fmt.Sprintf("User '%s' is not permitted to add key '%s'.", user.Name, input.Key), err)
		} else if errorMessages.Matches(err, errorMessages.ErrKeyExists) {
			return &PutKeyResponse{}, huma.Error409Conflict(fmt.Sprintf("Key '%s' already exists.", input.Key), err)
		} else {
			return &PutKeyResponse{}, huma.Error400BadRequest(fmt.Sprintf("Cannot add key '%s'.", input.Key), err)
		}
	} else {
		log.Printf("User '%s' (%s) put or updated key '%s'.\n", user.Name, user.Email, input.Key)
		return &PutKeyResponse{
			Body: len(input.RawBody),
		}, nil
	}
}

// PutKey adds a new key to the cache.
func PatchKey(
	ctx context.Context,
	authManager *auth.AuthManager,
	kvc *keyValue.KeyValueCache,
	input *PutKeyRequest,
) (*PutKeyResponse, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return &PutKeyResponse{}, huma.Error401Unauthorized("Authentication failed.", errorMessages.ErrUnauthorized)
	}

	if err := kvc.UpdateValue(input.Key, input.RawBody, user); err != nil {
		if errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
			return &PutKeyResponse{}, huma.Error403Forbidden(fmt.Sprintf("User '%s' is not permitted to update key '%s'.", user.Name, input.Key), err)
		} else if errorMessages.Matches(err, errorMessages.ErrKeyNotFound) {
			return &PutKeyResponse{}, huma.Error404NotFound(fmt.Sprintf("Key '%s' does not exists.", input.Key), err)
		} else {
			return &PutKeyResponse{}, huma.Error400BadRequest(fmt.Sprintf("Cannot update key '%s'.", input.Key), err)
		}
	} else {
		log.Printf("User '%s' (%s) updated key '%s'.\n", user.Name, user.Email, input.Key)
		return &PutKeyResponse{
			Body: len(input.RawBody),
		}, nil
	}
}

// PostKey adds or updates a key in the cache.
func PostKey(
	ctx context.Context,
	authManager *auth.AuthManager,
	kvc *keyValue.KeyValueCache,
	input *PostKeyRequest,
) (*PostKeyResponse, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return &PostKeyResponse{}, huma.Error401Unauthorized("Authentication failed.", errorMessages.ErrUnauthorized)
	}

	if err := kvc.PutOrUpdateValue(input.Key, input.RawBody, user); err != nil {
		if errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
			return &PostKeyResponse{}, huma.Error403Forbidden(fmt.Sprintf("User '%s' is not permitted to add key '%s'.", user.Name, input.Key), err)
		} else {
			return &PostKeyResponse{}, huma.Error400BadRequest(fmt.Sprintf("Cannot add key '%s'.", input.Key), err)
		}
	} else {
		log.Printf("User '%s' (%s) put or updated key '%s'.\n", user.Name, user.Email, input.Key)
		return &PostKeyResponse{
			Body: len(input.RawBody),
		}, nil
	}
}

// DeleteKey removes a key from the cache.
func DeleteKey(
	ctx context.Context,
	authManager *auth.AuthManager,
	kvc *keyValue.KeyValueCache,
	input *DeleteKeyRequest,
) (*DeleteKeyResponse, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return &DeleteKeyResponse{}, huma.Error401Unauthorized("Authentication failed.", errorMessages.ErrUnauthorized)
	}

	if delivery, err := kvc.DeleteValue(input.Key, user); err != nil {
		if errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
			return &DeleteKeyResponse{}, huma.Error403Forbidden(fmt.Sprintf("User '%s' is not permitted to delete key '%s'.", user.Name, input.Key), err)
		} else if errorMessages.Matches(err, errorMessages.ErrKeyNotFound) {
			return &DeleteKeyResponse{}, huma.Error404NotFound(fmt.Sprintf("Failed to find key '%s'.", input.Key), err)
		} else {
			return &DeleteKeyResponse{}, huma.Error400BadRequest(fmt.Sprintf("Cannot delete key '%s'.", input.Key), err)
		}
	} else {
		log.Printf("User '%s' (%s) deleted key '%s'.\n", user.Name, user.Email, input.Key)
		return &DeleteKeyResponse{Body: delivery.Value}, nil
	}
}
