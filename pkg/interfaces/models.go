package interfaces

import (
	"context"
	"time"

	"github.com/denwong47/pigeon-hole/pkg/auth"
	keyValue "github.com/denwong47/pigeon-hole/pkg/key_value"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

// Short Hand for the EndpointHandler function signature.
type EndpointHandler[T, R any] func(ctx context.Context, input *T) (*R, error)

// Short Hand for an EndpointHandler that uses an AuthManager.
type EndpointHandlerWithAuthManager[T, R any] func(ctx context.Context, authManager *auth.AuthManager, input *T) (*R, error)

// Short Hand for an EndpointHandler that uses a KeyValueCache as well as an AuthManager.
type EndpointHandlerWithKeyValueCache[T, R any] func(ctx context.Context, authManager *auth.AuthManager, keyValueCache *keyValue.KeyValueCache, input *T) (*R, error)

// KeyRequest is the request object for all endpoints that takes a single key as input.
type KeyRequest struct {
	Authorization string `header:"Authorization" doc:"The bearer token for authorization. No other form of authorization is supported." example:"Bearer token"`
	Key           string `path:"key" maxLength:"1024" example:"myObjectKey" doc:"The object key of the desired delivery. Obtain this from the sender."`
}

// AuthorizationToken extracts the token from the Authorization header.
func (k *KeyRequest) AuthorizationToken() (string, error) {
	return ParseBearerAuthorization(k.Authorization)
}

// DeliveryResponse is the response object for the delivery endpoint.
type DeliveryResponse struct {
	Body keyValue.KeyValueDelivery `json:"body" doc:"The object that was delivered."`
}

type AddUserRequest struct {
	Body struct {
		Name     string `json:"name" doc:"The name of the user to add." required:"true" minLength:"1" maxLength:"1024"`
		Email    string `json:"email" format:"email" doc:"The email of the user to add." required:"true" minLength:"1" maxLength:"1024"`
		Password string `json:"password" doc:"The password of the user to add." required:"true" minLength:"8" example:"mySamplePasswordChangeBeforeUse"`
		Type     string `json:"type" enum:"admin,standard,restricted" doc:"The kind of user to add; this governs the privileges of the user."`
	}
}

// AddUserResponse is the response object for the AddUser endpoint.
type AddUserResponse struct {
	Body TokenResponseBody `json:"body" doc:"Content of the response."`
}

// AddUserResponseBody is the response object for the AddUser endpoint.
type TokenResponseBody struct {
	Token  string    `json:"token" doc:"The Auth token of the requested user."`
	Expiry time.Time `json:"expiry" doc:"The time the token will expire."`
}

// RemoveUserRequest is the request object for the RemoveUser endpoint.
type RemoveUserRequest struct {
	Email string `path:"email" format:"email" doc:"The email of the user to remove." required:"true" minLength:"1" maxLength:"1024" example:"user@example.com"`
}

// RemoveUserResponse is the response object for the RemoveUser endpoint.
type RemoveUserResponse struct {
	Body struct{} `json:"body" doc:"Content of the response."`
}

// LoginUserRequest is the request object for the LoginUser endpoint.
type LoginUserRequest struct {
	Body struct {
		Email    string `json:"email" format:"email" doc:"The email of the user to login." required:"true" minLength:"1" maxLength:"1024"`
		Password string `json:"password" doc:"The password of the user to login." required:"true" minLength:"8" example:"mySamplePasswordChangeBeforeUse"`
	}
}

// LoginUserResponse is the response object for the LoginUser endpoint.
type LoginUserResponse struct {
	Authorization string            `header:"Authorization" doc:"The Auth token of the requested user. Obtain using the '/login' endpoint." example:"Bearer token"`
	Body          TokenResponseBody `json:"body" doc:"Content of the response."`
}

// LogoutUserRequest is the request object for the LogoutUser endpoint.
type LogoutUserRequest struct {
	Authorization string `header:"Authorization" doc:"The Auth token of the requested user. Obtain using the '/login' endpoint." example:"Bearer token"`
}

// LogoutUserResponse is the response object for the LogoutUser endpoint.
type LogoutUserResponse struct {
	Body struct{} `json:"body" doc:"Content of the response."`
}

// GetUserPermissionRequest is the request object for the GetUserPermission endpoint.
type GetUserPermissionRequest LogoutUserRequest

// GetUserPermissionResponse is the response object for the GetUserPermission endpoint.
type GetUserPermissionResponse struct {
	Body users.Privileges `json:"privileges" doc:"The privileges of the user."`
}

// GetKeyRequest is the request object for the GetKey endpoint.
type GetKeyRequest struct {
	Authorization string `header:"Authorization" doc:"The Auth token of the requested user. Obtain using the '/login' endpoint." example:"Bearer token"`
	Key           string `path:"key" maxLength:"1024" example:"myObjectKey" doc:"The object key of the desired delivery. Obtain this from the sender."`
}

// GetKeyResponse is the response object for the GetKey endpoint.
type GetKeyResponse struct {
	// Body keyValue.KeyValueDelivery `json:"body" doc:"Content of the response."`
	Body []byte `doc:"The byte content of the stored object."`
}

// PutKeyRequest is the request object for the PutKey endpoint.
type PutKeyRequest struct {
	Authorization string `header:"Authorization" doc:"The Auth token of the requested user. Obtain using the '/login' endpoint." example:"Bearer token"`
	Key           string `path:"key" maxLength:"1024" example:"myObjectKey" doc:"The object key of the desired delivery. Obtain this from the sender."`
	RawBody       []byte
}

// GetKeyResponse is the response object for the GetKey endpoint.
type PutKeyResponse struct {
	Body int `json:"body" doc:"Size of the bytes inserted."`
}

// PostKeyRequest is the request object for the PostKey endpoint.
type PostKeyRequest PutKeyRequest

// PostKeyResponse is the response object for the PostKey endpoint.
type PostKeyResponse PutKeyResponse

// DeleteKeyRequest is the request object for the DeleteKey endpoint.
type DeleteKeyRequest GetKeyRequest

// DeleteKeyResponse is the response object for the DeleteKey endpoint.
type DeleteKeyResponse GetKeyResponse
