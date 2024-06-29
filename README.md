# PigeonHole service

![Go CI Badge](https://github.com/denwong47/pigeon-hole/actions/workflows/audit.yml/badge.svg?branch=main)

A simple in-memory key-value store with user authentication meant for local
networks.

This is like a pigeon hole at the conceige where you can leave messages for
others to pick up. It provides a simple REST API for storing and retrieving
bytes of data with a key. Very rudimentary user authentication is provided,
with support for different permissions for owned and unowned keys.

Thread safety is designed, so multiple clients can read and write to the store
while maintaining regularity of the register. This is achieved by using a
read-write mutex to protect each key-value pair.

The authentication system requires the host system to be secure, as user
management is automatically permitted from loopback addresses. As such, this
service is not suitable for systems that can have multiple user logins, or
run in docker containers with "host" network access.

## Usage

To run the service, simply run the main file:

```sh
$ go run .
```

Possible CLI options are:

```
  -h, --help               help for pigeon-hole
      --host string        Host to listen on. (default "0.0.0.0")
  -p, --port int           Port to listen on. (default 8888)
      --salt string        Salt for hashing passwords.
                           This should not be stored anywhere, as they can make cracking the 
                           stored hashes easier. Provide this at runtime to minimise the 
                           chance of attack.
      --user-list string   Path to the user list file. (default "./users.json")
```

When the service is running, you can check the API documentation at the `/docs`
endpoint. For example, if the service is running on `localhost:8888`, you can
access the documentation at `http://localhost:8888/docs`.
