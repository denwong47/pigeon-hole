package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"

	// Import the interfaces for APIs

	"github.com/denwong47/pigeon-hole/pkg/auth"
	"github.com/denwong47/pigeon-hole/pkg/cli"
	"github.com/denwong47/pigeon-hole/pkg/interfaces"
	keyValue "github.com/denwong47/pigeon-hole/pkg/key_value"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

func main() {
	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *cli.Options) {
		// TODO - Implement CLI flags for setting the port and other options
		router := chi.NewMux()
		config := huma.DefaultConfig("PigeonHole", "0.1.0")
		config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"BearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "base64",
			},
		}
		api := humachi.New(router, config)
		api.OpenAPI().Info.Title = "PigeonHole service"
		api.OpenAPI().Info.Description = description + "\n\n" + disclaimer
		api.OpenAPI().Info.Contact = &huma.Contact{
			Name:  "Denny Wong",
			Email: "denwong47@hotmail.com",
			URL:   "https://github.com/denwong47/pigeon-hole",
		}
		log.Printf("Starting PigeonHole service...\n")
		log.Printf("Put operations will be limited to %v.\n", options.Timeout)

		userOptions := users.UserOptions{
			// TODO Move this into a configuration file
			Salt:            options.Salt,
			TokenExpiration: time.Hour * 2,
		}

		authManager, err := auth.ImportFromOrNew(options.UserList, "Global User List", userOptions)
		log.Printf("Loaded %d users from file.\n", authManager.Length())
		if err != nil {
			fmt.Println("Failed to import users from file:", err)
			os.Exit(1)
		}

		api.UseMiddleware(interfaces.PassThroughRemoteHost)
		api.UseMiddleware(interfaces.PassThroughAuthorizationToken(authManager))

		// Add the User endpoints.
		// These endpoints will have a minimum return time of 1 second to
		// prevent timing attacks, and will only be accessible from the loopback
		// address.

		// `AddUser``
		huma.Register(api, huma.Operation{
			Method:  http.MethodPut,
			Path:    "/user",
			Summary: "Add a new user",
			Description: `Add a new user to the system. The user will be created with the provided
			name, email, and password. The user type will default to "standard" if "type" is not provided.` + loopbackOnly,
			Errors: []int{200, 403, 408, 500},
		}, interfaces.MinimumTimeReturn(
			time.Second,
			interfaces.MustBeCalledFromLoopBack(interfaces.UsesAuthManager(authManager, interfaces.AddUser)),
		))
		// `RemoveUser``
		huma.Register(api, huma.Operation{
			Method:  http.MethodDelete,
			Path:    "/user/{email}",
			Summary: "Remove a user",
			Description: `Remove a user from the system. The user will be removed from the system
			based on the provided email address. This action is irreversible.` + loopbackOnly,
			Errors: []int{200, 401, 403, 404, 500},
		}, interfaces.MinimumTimeReturn(
			time.Second,
			interfaces.MustBeCalledFromLoopBack(interfaces.UsesAuthManager(authManager, interfaces.RemoveUser)),
		))
		// `LoginUser``
		huma.Register(api, huma.Operation{
			Method:  http.MethodPost,
			Path:    "/login",
			Summary: "Login",
			Description: fmt.Sprintf(
				`Login to the system. This will return a base64 token
				that can be used to authenticate future requests. The token will
				expire after %v. To use the token, send it in the Authorization header
				as: <pre>Bearer &lt;token&gt;</pre>`,
				userOptions.TokenExpiration,
			),
			Errors: []int{200, 401, 500},
		}, interfaces.MinimumTimeReturn(
			time.Second,
			interfaces.UsesAuthManager(authManager, interfaces.LoginUser),
		))
		// `LogoutUser``
		huma.Register(api, huma.Operation{
			Method:  http.MethodPost,
			Path:    "/logout",
			Summary: "Logout",
			Description: `Logout of the system. This will invalidate the current token and
			prevent it from being used in future requests.` + requiresBearerAuth,
			Errors: []int{200, 401},
		}, interfaces.UsesAuthManager(authManager, interfaces.LogoutUser))

		// `GetUserPermission``
		huma.Register(api, huma.Operation{
			Method:      http.MethodGet,
			Path:        "/user/permission",
			Summary:     "Get User Permission",
			Description: `Get the permission of the user. This will return the permission of the user based on the provided token.` + requiresBearerAuth,
			Errors:      []int{200, 401},
		}, interfaces.UsesAuthManager(authManager, interfaces.GetUserPermission))

		kvc := keyValue.NewCache()
		// Add the Key Value endpoints

		// `GetKey``
		huma.Register(api, huma.Operation{
			Method:      http.MethodGet,
			Path:        "/key/{key}",
			Summary:     "Get Data by Key",
			Description: `Fetch bytes data by the provided key.` + userPermissionsNote + requiresBearerAuth,
			Errors:      []int{200, 401, 403, 404},
		}, interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.GetKey))
		// `PatchKey`
		huma.Register(api, huma.Operation{
			Method:      http.MethodPatch,
			Path:        "/key/{key}",
			Summary:     "Update Data by Key",
			Description: `Update bytes data by the provided key, only if the key already exists.` + userPermissionsNote + requiresBearerAuth,
			Errors:      []int{200, 401, 403, 404, 504},
		}, interfaces.MaximumTimeReturn(
			options.Timeout,
			interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.PatchKey)),
		)
		// `PutKey`
		huma.Register(api, huma.Operation{
			Method:      http.MethodPut,
			Path:        "/key/{key}",
			Summary:     "Add new Data by Key",
			Description: `Add bytes data to a new key. This will only succeed if the key does not already exist.` + userPermissionsNote + requiresBearerAuth,
			Errors:      []int{200, 401, 403, 408, 504},
		}, interfaces.MaximumTimeReturn(
			options.Timeout,
			interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.PutKey)),
		)
		// `PostKey`
		huma.Register(api, huma.Operation{
			Method:      http.MethodPost,
			Path:        "/key/{key}",
			Summary:     "Add or update Data by Key",
			Description: `Upsert bytes data by the provided key.` + userPermissionsNote + requiresBearerAuth,
			Errors:      []int{200, 401, 403, 504},
		}, interfaces.MaximumTimeReturn(
			options.Timeout,
			interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.PostKey)),
		)
		// `DeleteKey``
		huma.Register(api, huma.Operation{
			Method:      http.MethodDelete,
			Path:        "/key/{key}",
			Summary:     "Delete Data by Key",
			Description: `Delete the provided key from the cache.` + userPermissionsNote + requiresBearerAuth,
			Errors:      []int{200, 401, 403, 404},
		}, interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.DeleteKey))

		server := http.Server{
			Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
			Handler: router,
		}

		hooks.OnStart(func() {
			server.ListenAndServe()
		})

		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
			defer authManager.ExportTo(options.UserList)
		})
	})

	cli.Run()
}
