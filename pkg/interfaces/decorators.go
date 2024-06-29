package interfaces

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/denwong47/pigeon-hole/pkg/auth"
	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
	keyValue "github.com/denwong47/pigeon-hole/pkg/key_value"
)

// Decorator to transform a `EndpointHandlerWithAuthManager` into a `EndpointHandler`.
func UsesAuthManager[T, R any](
	authManager *auth.AuthManager,
	handler EndpointHandlerWithAuthManager[T, R],
) EndpointHandler[T, R] {
	return func(ctx context.Context, input *T) (*R, error) {
		return handler(ctx, authManager, input)
	}
}

// Decorator to transform a `EndpointHandlerWithKeyValueCache` into a `EndpointHandler`.
func UsesAuthManagerAndKeyValueCache[T, R any](
	authManager *auth.AuthManager,
	keyValueCache *keyValue.KeyValueCache,
	handler EndpointHandlerWithKeyValueCache[T, R],
) EndpointHandler[T, R] {
	return func(ctx context.Context, input *T) (*R, error) {
		return handler(ctx, authManager, keyValueCache, input)
	}
}

// Check if the origin of the request is from the loopback address; if not return an Forbidden error.
//
// Must be used in conjunction with the `PassThroughRemoteHost` middleware.
func MustBeCalledFromLoopBack[T, R any](handler EndpointHandler[T, R]) EndpointHandler[T, R] {
	return func(ctx context.Context, input *T) (*R, error) {
		if isLoopback, ok := ctx.Value(CONTEXT_VALUE_REMOTE_IS_LOOPBACK).(bool); !ok || !isLoopback {
			return nil, huma.Error403Forbidden(
				"This request must originate from the loopback address. "+
					"Please SSH into the host and perform the request from there.",
				errorMessages.ErrRemoteHostForbidden,
			)
		}
		return handler(ctx, input)
	}
}
