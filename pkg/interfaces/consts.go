package interfaces

// CONTEXT_VALUE_REMOTE_ADDR is the key for the remote address stored in the context by the Middleware.
const CONTEXT_VALUE_REMOTE_ADDR = "remote_addr"

// CONTEXT_VALUE_REMOTE_IS_LOOPBACK is the key for whether the remote address is a loopback address.
// Some security-sensitive operations may only be permitted from loopback addresses.
const CONTEXT_VALUE_REMOTE_IS_LOOPBACK = "remote_is_loopback"

// CONTEXT_VALUE_AUTH_TOKEN is the key for the authorization token stored in the context by the Middleware.
const CONTEXT_VALUE_AUTH_TOKEN = "auth_token"

// CONTEXT_VALUE_AUTH_USER is the key for the user object stored in the context by the Middleware.
const CONTEXT_VALUE_AUTH_USER = "auth_user"
