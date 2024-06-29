package cli

// Options for the CLI.
type Options struct {
	Host     string `doc:"Host to listen on" format:"ipv4" default:"0.0.0.0"`
	Port     int    `doc:"Port to listen on" short:"p" default:"8888"`
	Salt     string `doc:"Salt for hashing passwords. This is not hard coded anywhere, as they can make cracking the stored hashes easier. Provide this at runtime to minimise the chance of attack" default:""`
	UserList string `doc:"Path to the user list file" default:"./users.json"`
}
