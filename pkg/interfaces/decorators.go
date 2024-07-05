package interfaces

import (
	"context"
	"log"
	"reflect"
	"runtime"
	"time"

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

// Waiter function.
func timeout(duration time.Duration, channel chan bool) {
	time.Sleep(duration)
	channel <- true
}

// Minimum time return.
//
// The decorated function will not return until the minimum time has passed. This may
// include sleeping if the time has not yet passed.
//
// This is to prevent timing attacks on passwords etc.
func MinimumTimeReturn[T, R any](
	duration time.Duration,
	handler EndpointHandler[T, R],
) EndpointHandler[T, R] {
	return func(ctx context.Context, input *T) (*R, error) {
		channel := make(chan bool)
		go timeout(duration, channel)

		result, err := handler(ctx, input)

		<-channel
		return result, err
	}
}

// Get the name of a function.
func getFunctionName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// Abort the request if the results had not been returned within the specified duration.
func MaximumTimeReturn[T, R any](
	duration time.Duration,
	handler EndpointHandler[T, R],
) EndpointHandler[T, R] {
	return func(ctx context.Context, input *T) (*R, error) {
		ctx, cancelFunc := context.WithTimeout(ctx, duration)
		defer cancelFunc()

		chR := make(chan *R)
		chErr := make(chan error)

		go func() {
			result, err := handler(ctx, input)
			if err != nil {
				chErr <- err
			} else {
				chR <- result
			}
		}()

		select {
		case result := <-chR:
			return result, nil
		case err := <-chErr:
			return nil, err
		case <-ctx.Done():
			log.Printf(
				"Operation `%s` took too long to process, cancelling after %v.\n",
				getFunctionName(handler),
				duration,
			)
			return nil, huma.Error504GatewayTimeout(
				"The request took too long to process. Please check if the request is valid.",
				errorMessages.ErrOperationTimeout,
			)
		}
	}
}
