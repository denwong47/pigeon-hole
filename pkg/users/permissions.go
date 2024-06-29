package users

import errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"

// Permissions is a struct that represents the permissions that a user has on an object.
type Permissions struct {
	Select bool `json:"read" doc:"Whether the user can read the object."`
	Insert bool `json:"write" doc:"Whether the user can write a new object."`
	Update bool `json:"update" doc:"Whether the user can update an object."`
	Delete bool `json:"delete" doc:"Whether the user can delete the object."`
}

// Privileges is a struct that represents the permissions that a user has on objects they own and all objects.
type Privileges struct {
	Owned Permissions `json:"owned" doc:"The permissions the user has on objects they own."`
	All   Permissions `json:"all" doc:"The permissions the user has on all objects."`
}

const (
	AdminUserType      string = "admin"
	StandardUserType   string = "standard"
	RestrictedUserType string = "restricted"
)

// Helper function to create a new Permissions struct with read-only permissions.
func ReadOnlyPermissions() Permissions {
	return Permissions{
		Select: true,
		Insert: false,
		Update: false,
		Delete: false,
	}
}

// Helper function to create a new Permissions struct with write-only permissions.
func NoPermissions() Permissions {
	return Permissions{
		Select: false,
		Insert: false,
		Update: false,
		Delete: false,
	}
}

// Helper function to create a new Privileges struct with Full permissions.
func FullPermissions() Permissions {
	return Permissions{
		Select: true,
		Insert: true,
		Update: true,
		Delete: true,
	}
}

// Helper function to create a new Privileges struct with full permissions
// for owned objects and read-only permissions for other objects.
func StandardUser() Privileges {
	return Privileges{
		Owned: FullPermissions(),
		All:   ReadOnlyPermissions(),
	}
}

// Helper function to create a new Privileges struct with full permissions
// for owned objects and no permissions for other objects.
func RestrictedUser() Privileges {
	return Privileges{
		Owned: FullPermissions(),
		All:   NoPermissions(),
	}
}

// Helper function to create a new Privileges struct with read-only permissions
// for owned objects and full permissions for other objects.
//
// This is typically not used, as there is no mechanism for a user to insert data
// other than their own.
func ReadOnlyUser() Privileges {
	return Privileges{
		Owned: ReadOnlyPermissions(),
		All:   NoPermissions(),
	}
}

// Helper function to create a new Privileges struct for Admin users.
func AdminUser() Privileges {
	return Privileges{
		Owned: FullPermissions(),
		All:   FullPermissions(),
	}
}

// Helper function to get the Privileges for a given UserType.
func GetPrivilegesByType(userType string) (Privileges, error) {
	switch userType {
	case AdminUserType:
		return AdminUser(), nil
	case StandardUserType:
		return StandardUser(), nil
	case RestrictedUserType:
		return RestrictedUser(), nil
	default:
		// No privileges for unknown user types.
		return Privileges{}, errorMessages.ErrUnknownUserType
	}
}
