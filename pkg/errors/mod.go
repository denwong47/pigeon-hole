/*
Package errorMessages contains all the error messages that are used in the application.
*/
package errorMessages

import (
	"errors"
)

var ErrNoAuthorizationHeader = errors.New("NoAuthorizationHeader")
var ErrUnauthorized = errors.New("Unauthorized")

var ErrRemoteHostForbidden = errors.New("RemoteHostForbidden")

var ErrUserAlreadyExists = errors.New("UserAlreadyExists")
var ErrUserNotFound = errors.New("UserNotFound")
var ErrAuthenticationFailed = errors.New("AuthenticationFailed")
var ErrNotPermitted = errors.New("NotPermitted")

var ErrUserFileNotFound = errors.New("UserFileNotFound")

var ErrUnknownUserType = errors.New("UnknownUserType")

var ErrInvalidRemoteAddr = errors.New("InvalidRemoteAddr")

var ErrKeyExists = errors.New("ErrKeyExists")
var ErrKeyNotFound = errors.New("ErrKeyNotFound")

var ErrTokenGeneration = errors.New("TokenGeneration")
var ErrTokenInvalid = errors.New("TokenInvalid")
var ErrTokenExpired = errors.New("TokenExpired")

var ErrOperationTimeout = errors.New("OperationTimeout")

func Matches(candidate error, target error) bool {
	return errors.Is(candidate, target)
}
