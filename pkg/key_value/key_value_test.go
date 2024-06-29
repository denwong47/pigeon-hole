package keyValue

import (
	"crypto/rand"
	"slices"
	"testing"

	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"

	users "github.com/denwong47/pigeon-hole/pkg/users"
)

func TestKeyValueCacheWithoutPermissions(t *testing.T) {
	kvc := NewCache()

	secret1 := make([]byte, 32)
	rand.Read(secret1)

	secret2 := make([]byte, 32)
	rand.Read(secret2)

	// Check that the cache is empty
	if kvc.Length() != 0 {
		t.Errorf("Expected 0 length, got %d", kvc.Length())
	}

	// Put a secret into the cache
	if err := kvc.Put("myKey1", KeyValueDelivery{Value: secret1}); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	}
	if kvc.Length() != 1 {
		t.Errorf("Expected 1 length, got %d", kvc.Length())
	}

	// Try putting the same key again
	if err := kvc.Put("myKey1", KeyValueDelivery{Value: secret2}); !errorMessages.Matches(err, errorMessages.ErrKeyExists) {
		t.Errorf(`Expected '%s', got '%s'`, errorMessages.ErrKeyExists, err)
	}

	// Put another secret into the cache
	if err := kvc.Put("myKey2", KeyValueDelivery{Value: secret2}); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	}
	if kvc.Length() != 2 {
		t.Errorf("Expected 2 length, got %d", kvc.Length())
	}

	// Get the first secret
	if delivery, err := kvc.Get("myKey1"); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, secret1) {
			t.Errorf("Expected secret1, got %v", delivery.Value)
		}
	}

	// Get the second secret
	if delivery, err := kvc.Get("myKey2"); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, secret2) {
			t.Errorf("Expected secret2, got %v", delivery.Value)
		}
	}

	// Try getting a non-existent key
	if _, err := kvc.Get("myKey3"); !errorMessages.Matches(err, errorMessages.ErrKeyNotFound) {
		t.Errorf(`Expected key to be not found, got '%s'`, err)
	}

	// Update the first secret
	if err := kvc.Update("myKey1", KeyValueDelivery{Value: secret2}); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	}

	// Get the first secret again
	if delivery, err := kvc.Get("myKey1"); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, secret2) {
			if slices.Equal(delivery.Value, secret1) {
				t.Errorf("Expected secret2, got secret1 instead")
			} else {
				t.Errorf("Expected secret2 (%v), got unknown value %v", secret2, delivery.Value)
			}
		}
	}

	// Put or update the first secret back to secret1
	if err := kvc.PutOrUpdate("myKey1", KeyValueDelivery{Value: secret1}); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	} else {
		if delivery, err := kvc.Get("myKey1"); err != nil {
			t.Errorf(`Expected no error, got '%s'`, err)
		} else {
			if !slices.Equal(delivery.Value, secret1) {
				t.Errorf("Expected secret1, got %v", delivery.Value)
			}
		}
	}

	// Put or update a new key, and check that it exists and is correct
	if err := kvc.PutOrUpdate("myKey3", KeyValueDelivery{Value: secret2}); err != nil {
		t.Errorf(`Expected no error, got '%s'`, err)
	} else {
		if delivery, err := kvc.Get("myKey3"); err != nil {
			t.Errorf(`Expected no error, got '%s'`, err)
		} else {
			if !slices.Equal(delivery.Value, secret2) {
				t.Errorf("Expected secret2, got %v", delivery.Value)
			}
		}
	}
}

func TestKeyValueCacheWithPermissions(t *testing.T) {
	kvc := NewCache()

	adminUser := users.NewUser(
		"Dave",
		"dave@test.com",
		users.AdminUser(),
	)
	standardUser := users.NewUser(
		"Steve",
		"steve@test.com",
		users.StandardUser(),
	)
	restrictedUser := users.NewUser(
		"Bob",
		"bob@test.com",
		users.RestrictedUser(),
	)
	readOnlyUser := users.NewUser(
		"John",
		"john@test.com",
		users.ReadOnlyUser(),
	)

	davesSecret := make([]byte, 32)
	rand.Read(davesSecret)
	stevesSecret := make([]byte, 32)
	rand.Read(stevesSecret)
	bobsSecret := make([]byte, 32)
	rand.Read(bobsSecret)
	johnsSecret := make([]byte, 32)
	rand.Read(johnsSecret)

	empty := make([]byte, 32)

	// Check insert own secret
	if err := kvc.PutValue("davesSecret", empty, &adminUser); err != nil {
		t.Errorf(`Expected no error inserting admin's secret, got '%s'`, err)
	}
	if err := kvc.PutValue("stevesSecret", empty, &standardUser); err != nil {
		t.Errorf(`Expected no error inserting standard's secret, got '%s'`, err)
	}
	if err := kvc.PutValue("bobsSecret", empty, &restrictedUser); err != nil {
		t.Errorf(`Expected no error inserting restricted's secret, got '%s'`, err)
	}
	if err := kvc.PutValue("johnsSecret", empty, &readOnlyUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error inserting read only's secret, got '%s'`, err)
	}

	// Use admin to force John's secret in
	if err := kvc.PutValueWithOwner("johnsSecret", empty, &readOnlyUser, &standardUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error inserting read only's secret by standard User, got '%s'`, err)
	}
	if err := kvc.PutValueWithOwner("johnsSecret", empty, &readOnlyUser, &adminUser); err != nil {
		t.Errorf(`Expected no error inserting read only's secret, got '%s'`, err)
	}

	// Check update own secret
	if err := kvc.UpdateValue("davesSecret", davesSecret, &adminUser); err != nil {
		t.Errorf(`Expected no error updating admin's secret, got '%s'`, err)
	}
	if err := kvc.UpdateValue("stevesSecret", stevesSecret, &standardUser); err != nil {
		t.Errorf(`Expected no error updating standard's secret, got '%s'`, err)
	}
	if err := kvc.UpdateValue("bobsSecret", bobsSecret, &restrictedUser); err != nil {
		t.Errorf(`Expected no error updating restricted's secret, got '%s'`, err)
	}
	if err := kvc.UpdateValue("johnsSecret", johnsSecret, &readOnlyUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error updating read only's secret, got '%s'`, err)
	}

	// Check update others secret
	if err := kvc.UpdateValue("johnsSecret", johnsSecret, &adminUser); err != nil {
		t.Errorf(`Expected no error updating read only's secret with admin, got '%s'`, err)
	}
	if err := kvc.UpdateValue("johnsSecret", stevesSecret, &standardUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error updating read only's secret with standard, got '%s'`, err)
	}
	if err := kvc.UpdateValue("johnsSecret", bobsSecret, &restrictedUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error updating read only's secret with restricted, got '%s'`, err)
	}

	// Check get own secret
	if delivery, err := kvc.GetValue("davesSecret", &adminUser); err != nil {
		t.Errorf(`Expected no error getting admin's secret, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, davesSecret) {
			t.Errorf("Expected davesSecret, got %v", delivery)
		}
	}
	if delivery, err := kvc.GetValue("stevesSecret", &standardUser); err != nil {
		t.Errorf(`Expected no error getting standard's secret, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, stevesSecret) {
			t.Errorf("Expected stevesSecret, got %v", delivery)
		}
	}
	if delivery, err := kvc.GetValue("bobsSecret", &restrictedUser); err != nil {
		t.Errorf(`Expected no error getting restricted's secret, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, bobsSecret) {
			t.Errorf("Expected bobsSecret, got %v", delivery)
		}
	}
	if delivery, err := kvc.GetValue("johnsSecret", &readOnlyUser); err != nil {
		t.Errorf(`Expected no error getting read only's secret, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, johnsSecret) {
			t.Errorf("Expected johnsSecret, got %v", delivery)
		}
	}

	// Check get others secret
	if delivery, err := kvc.GetValue("johnsSecret", &adminUser); err != nil {
		t.Errorf(`Expected no error getting read only's secret with admin, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, johnsSecret) {
			t.Errorf("Expected johnsSecret, got %v", delivery)
		}
	}
	if delivery, err := kvc.GetValue("johnsSecret", &standardUser); err != nil {
		t.Errorf(`Expected no error getting read only's secret with standard, got '%s'`, err)
	} else {
		if !slices.Equal(delivery.Value, johnsSecret) {
			t.Errorf("Expected johnsSecret, got %v", delivery)
		}
	}
	if _, err := kvc.GetValue("johnsSecret", &restrictedUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error getting read only's secret with restricted, got '%s'`, err)
	}
	if _, err := kvc.GetValue("davesSecret", &readOnlyUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error getting admin's secret with read only, got '%s'`, err)
	}

	// Delete someone else's secret
	if _, err := kvc.DeleteValue("johnsSecret", &standardUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error deleting read only's secret with standard, got '%s'`, err)
	}
	if _, err := kvc.DeleteValue("johnsSecret", &restrictedUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error deleting read only's secret with restricted, got '%s'`, err)
	}
	if _, err := kvc.DeleteValue("johnsSecret", &readOnlyUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error deleting read only's secret with read only, got '%s'`, err)
	}

	// Delete own secret
	if _, err := kvc.DeleteValue("davesSecret", &adminUser); err != nil {
		t.Errorf(`Expected no error deleting admin's secret, got '%s'`, err)
	}
	if _, err := kvc.DeleteValue("stevesSecret", &standardUser); err != nil {
		t.Errorf(`Expected no error deleting standard's secret, got '%s'`, err)
	}
	if _, err := kvc.DeleteValue("bobsSecret", &restrictedUser); err != nil {
		t.Errorf(`Expected no error deleting restricted's secret, got '%s'`, err)
	}
	if _, err := kvc.DeleteValue("johnsSecret", &readOnlyUser); !errorMessages.Matches(err, errorMessages.ErrNotPermitted) {
		t.Errorf(`Expected "ErrNotPermitted" error deleting read only's secret, got '%s'`, err)
	}

	// Admin delete someone else's secret
	if _, err := kvc.DeleteValue("johnsSecret", &adminUser); err != nil {
		t.Errorf(`Expected no error deleting read only's secret with admin, got '%s'`, err)
	}

	// Try Put and Update
	curLength := kvc.Length()
	if err := kvc.PutOrUpdateValue("davesSecret", empty, &adminUser); err != nil {
		t.Errorf(`Expected no error upserting admin's secret, got '%s'`, err)
	}
	if err := kvc.PutOrUpdateValue("stevesSecret", empty, &standardUser); err != nil {
		t.Errorf(`Expected no error upserting standard's secret, got '%s'`, err)
	}
	if err := kvc.PutOrUpdateValue("bobsSecret", empty, &restrictedUser); err != nil {
		t.Errorf(`Expected no error upserting restricted's secret, got '%s'`, err)
	}

	if kvc.Length() != curLength+3 {
		t.Errorf("Expected %d length, got %d", curLength+3, kvc.Length())
	}

	// Try Put and Update again on own secret
	if err := kvc.PutOrUpdateValue("davesSecret", davesSecret, &adminUser); err != nil {
		t.Errorf(`Expected no error upserting admin's secret, got '%s'`, err)
	}
	if err := kvc.PutOrUpdateValue("stevesSecret", stevesSecret, &standardUser); err != nil {
		t.Errorf(`Expected no error upserting standard's secret, got '%s'`, err)
	}
	if err := kvc.PutOrUpdateValue("bobsSecret", bobsSecret, &restrictedUser); err != nil {
		t.Errorf(`Expected no error upserting restricted's secret, got '%s'`, err)
	}

	if kvc.Length() != curLength+3 {
		t.Errorf("Expected %d length, got %d", curLength+3, kvc.Length())
	}
}
