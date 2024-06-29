package interfaces

import (
	"regexp"

	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
)

// ParseBearerAuthorization extracts the token from the Authorization header.
func ParseBearerAuthorization(contents string) (string, error) {
	bearerPattern := regexp.MustCompile(`^Bearer (?P<token>.+)\w*$`)

	matches := bearerPattern.FindStringSubmatch(contents)

	if len(matches) < 2 {
		if contents == "" {
			return "", errorMessages.ErrNoAuthorizationHeader
		} else {
			return "", errorMessages.ErrUnauthorized
		}
	}

	return matches[1], nil
}
