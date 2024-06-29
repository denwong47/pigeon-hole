package interfaces

import (
	"context"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/denwong47/pigeon-hole/pkg/auth"
	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

// The Regex pattern for parsing the remote address.
var HOSTNAME_PATTERN = regexp.MustCompile(`^(?P<host>(?:\[?(?P<ipv6>[\d:]+\d)\]?)|(?P<ipv4>\d{1,3}(?:\.\d{1,3}){3})|(?P<hostname>\S+)):(?P<port>\d+)$`)

type RemoteHost struct {
	IP   net.IP `doc:"The IP address of the client making the request."`
	Port int    `doc:"The port of the client making the request."`
}

// ParseRemoteAddr extracts the IP address and port from the remote address.
func parseRemoteAddr(ctx huma.Context) (RemoteHost, error) {
	remoteAddr := ctx.RemoteAddr()

	parsed := HOSTNAME_PATTERN.FindAllStringSubmatch(remoteAddr, -1)
	if parsed == nil {
		return RemoteHost{}, errorMessages.ErrInvalidRemoteAddr
	}

	captures := make(map[string]string)
	// Extract the named captures from the regex.
	for i, name := range HOSTNAME_PATTERN.SubexpNames() {
		if i != 0 && name != "" {
			captures[name] = parsed[0][i]
		}
	}

	if port, err := strconv.Atoi(captures["port"]); err != nil {
		return RemoteHost{}, errorMessages.ErrInvalidRemoteAddr
	} else {
		// IPv6 could be wrapped in square brackets, so we need to remove them.
		if ip := net.ParseIP(strings.Trim(captures["host"], "[]")); ip == nil {
			return RemoteHost{}, errorMessages.ErrInvalidRemoteAddr
		} else {
			return RemoteHost{ip, port}, nil
		}
	}
}

// Repackage the remote address into the `huma.Context`.
func PassThroughRemoteHost(ctx huma.Context, next func(huma.Context)) {

	remoteHost, err := parseRemoteAddr(ctx)
	if err == nil {
		ctx = huma.WithValue(ctx, CONTEXT_VALUE_REMOTE_ADDR, remoteHost)
		ctx = huma.WithValue(ctx, CONTEXT_VALUE_REMOTE_IS_LOOPBACK, remoteHost.IP.IsLoopback())
	} else {
		log.Printf("Error parsing remote address %s: %s\n", ctx.RemoteAddr(), err)
	}

	// Call the next middleware in the chain. This eventually calls the
	// operation handler as well.
	next(ctx)
}

// Repackage the authorization header into the `huma.Context`.
func PassThroughAuthorizationToken(
	authManager *auth.AuthManager,
) func(huma.Context, func(huma.Context)) {

	log.Printf("Will use AuthManager `%s`.\n", authManager.Name)

	return func(ctx huma.Context, next func(huma.Context)) {
		token, err := ParseBearerAuthorization(ctx.Header("Authorization"))

		if err == nil {
			ctx = huma.WithValue(ctx, CONTEXT_VALUE_AUTH_TOKEN, token)

			// If the token is valid, add the user to the context.
			if tokenData, err := authManager.Tokens.GetToken(token); err == nil {
				ctx = huma.WithValue(ctx, CONTEXT_VALUE_AUTH_USER, tokenData.User)
			} else {
				log.Printf("Error getting user by token %s: %s\n", token, err)
			}
		} else if token := ctx.Header("Authorization"); token != "" {
			log.Printf("Error parsing Authorisation Token %s: %s\n", ctx.Header("Authorization"), err)
		}

		// Call the next middleware in the chain. This eventually calls the
		// operation handler as well.
		next(ctx)
	}
}

// Extracts the remote address from the context as inserted by the `PassThroughAuthorizationToken` middleware.
func GetTokenFromContext(ctx context.Context) (string, bool) {
	if token, ok := ctx.Value(CONTEXT_VALUE_AUTH_TOKEN).(string); ok {
		return token, true
	}
	return "", false
}

// Extracts the `users.User` object from the context as inserted by the `PassThroughAuthorizationToken` middleware.
func GetUserFromContext(ctx context.Context) (*users.User, bool) {
	if user, ok := ctx.Value(CONTEXT_VALUE_AUTH_USER).(*users.User); ok {
		return user, true
	}
	return nil, false
}
