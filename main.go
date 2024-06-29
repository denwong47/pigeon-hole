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
				BearerFormat: "Hex",
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

		// Add the User endpoints

		// `AddUser``
		huma.Put(api, "/user", interfaces.MustBeCalledFromLoopBack(interfaces.UsesAuthManager(authManager, interfaces.AddUser)))
		// `RemoveUser``
		huma.Delete(api, "/user", interfaces.MustBeCalledFromLoopBack(interfaces.UsesAuthManager(authManager, interfaces.RemoveUser)))
		// `LoginUser``
		huma.Post(api, "/login", interfaces.UsesAuthManager(authManager, interfaces.LoginUser))
		// `LogoutUser``
		huma.Get(api, "/logout", interfaces.UsesAuthManager(authManager, interfaces.LogoutUser))

		kvc := keyValue.NewCache()
		// Add the Key Value endpoints

		// `GetKey``
		huma.Get(api, "/key/{key}", interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.GetKey))
		// `PatchKey``
		huma.Patch(api, "/key/{key}", interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.PatchKey))
		// `PutKey``
		huma.Put(api, "/key/{key}", interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.PutKey))
		// `PostKey``
		huma.Post(api, "/key/{key}", interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.PostKey))
		// `DeleteKey``
		huma.Delete(api, "/key/{key}", interfaces.UsesAuthManagerAndKeyValueCache(authManager, &kvc, interfaces.DeleteKey))

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
