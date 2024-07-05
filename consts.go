package main

const description = `# PigeonHole service

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
`

const disclaimer = `## Disclaimer

This service is provided as-is with no warranty or guarantee of any kind.

You are free to use this service for any purpose, but you do so at
your own risk.  The author(s) of this service are not responsible
for any damages or losses that may occur as a result of using this
service.

By using this service, you agree to these terms.
`

const loopbackOnly = `

> **Note**
> 
> This endpoint is only accessible from the loopback address, i.e. by
> making the request from host machine using '127.0.0.1' or 'localhost' as the
> remote address. This is to minimise unauthorized access to the user management
> system.
`

const requiresBearerAuth = `

> **Note**
>
> This endpoint requires a valid Bearer token to be provided in the
> 'Authorization' header. The token should be in the format:
> <pre>Bearer &lt;token&gt;</pre>
>
> This token can be obtained by logging in to the system via the
> <a href="/paths/login/post">/login</a> endpoint.
`

const userPermissionsNote = `

The provided authorization token must have the correct permissions to perform
the requested action. Use <a href="/paths/user-permission/get">/user/permission</a>
endpoint to check your permissions.

Permissions are handled on a per-key basis. The first user to create a key is
considered the owner of that key; and your permissions to read, insert, update or
delete data are based on whether you are the owner of the key.
`
