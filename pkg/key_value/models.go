package keyValue

import (
	"sync"
	"time"

	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
	"github.com/denwong47/pigeon-hole/pkg/users"
)

// KeyValueEntry is a single entry in the key-value store.
//
// This is an internal struct that is not exposed to the API; it contains mutex locks
// to ensure that the data is accessed safely.
type KeyValueEntry struct {
	Delivery KeyValueDelivery
	lock     *sync.RWMutex
}

// Lock the cache for writing, and perform the specified operation.
func (kvc *KeyValueCache) LockAndDo(key string, operation func(*KeyValueDelivery) error) error {
	// Lock the whole cache for reading in case the key got deleted between the check
	// and the lock.
	kvc.lock.RLock()
	defer kvc.lock.RUnlock()

	if entry, ok := kvc.Contents[key]; !ok {
		return errorMessages.ErrKeyNotFound
	} else {
		// Since these locks are passed by reference, locking this entry will lock
		// the entry in the cache.
		entry.lock.Lock()
		defer entry.lock.Unlock()

		// Pass the entry by reference, and perform the operation
		if err := operation(&(&entry).Delivery); err != nil {
			return err
		}

		// Since we may have reassigned some fields in `entry`, which would NOT be
		// reflected in the cache, we need to reassign the entry back to the cache.
		kvc.Contents[key] = entry

		return nil
	}
}

// KeyValueTimestamps is a struct that contains the timestamps associated with a key-value object.
//
// Not all of them are required; the `CreatedAt` field is always required, but the `DeliveredAt` field
// is only required when the object is retrieved.
type KeyValueTimestamps struct {
	CreatedAt   time.Time `json:"createdAt,omitempty" doc:"The time this object was created."`
	DeliveredAt time.Time `json:"deliveredAt,omitempty" doc:"The time this object was retrieved."`
}

// KeyValueOwnership is a struct that contains the ownership information of a key-value object.
type KeyValueOwnership struct {
	Email *string `json:"email" doc:"The owner of this object."`
	Name  *string `json:"name,omitempty" doc:"The name of the owner of this object."`
}

// KeyValueDelivery is the response object for the delivery endpoint.
type KeyValueDelivery struct {
	Value      []byte             `json:"value" doc:"The byte content of the stored object in base64 encoding."`
	Timestamps KeyValueTimestamps `json:"timestamps" doc:"The timestamps associated with this object."`
	Ownership  KeyValueOwnership  `json:"ownedBy"`
}

// KeyValueCache is a simple key-value store that can be used to store and retrieve data.
//
// The `sync.RWMutex` in this struct is used to ensure key creation and deletion is thread-safe;
// for getting and setting existing values, use the `lock` field in `KeyValueEntry` instead.
type KeyValueCache struct {
	Contents map[string]KeyValueEntry
	lock     *sync.RWMutex
}

// New creates a new key-value cache with empty contents.
func NewCache() KeyValueCache {
	return KeyValueCache{
		Contents: make(map[string]KeyValueEntry),
		lock:     &sync.RWMutex{},
	}
}

// Fetch an object from the cache.
func (kvc *KeyValueCache) Get(key string) (KeyValueDelivery, error) {
	if found, ok := kvc.Contents[key]; !ok {
		return KeyValueDelivery{}, errorMessages.ErrKeyNotFound
	} else {
		return found.Delivery, nil
	}
}

// Fetch a value from the cache.
func (kvc *KeyValueCache) GetValue(key string, user *users.User) (KeyValueDelivery, error) {
	if delivery, err := kvc.Get(key); err == nil {
		if user.CanSelect(delivery.Ownership.Email == &user.Email) {
			return delivery, nil
		} else {
			return KeyValueDelivery{}, errorMessages.ErrNotPermitted
		}
	} else {
		return KeyValueDelivery{}, err
	}
}

// Get the length of the cache.
func (kvc *KeyValueCache) Length() int {
	return len(kvc.Contents)
}

// Put an object into the cache.
//
// If the key already exists, this will return an error.
//
// This is a low level function that does not check any user permissions;
// the whole `KeyValueDelivery` object is stored as-is.
func (kvc *KeyValueCache) Put(key string, value KeyValueDelivery) error {
	kvc.lock.Lock()
	defer kvc.lock.Unlock()

	if _, ok := kvc.Contents[key]; ok {
		return errorMessages.ErrKeyExists
	}

	kvc.Contents[key] = KeyValueEntry{
		Delivery: value,
		lock:     &sync.RWMutex{},
	}

	return nil
}

// Put a value into the cache, with the user as the owner.
//
// This is a high level function that will set the `Ownership` field of the object.
func (kvc *KeyValueCache) PutValue(key string, value []byte, user *users.User) error {
	return kvc.PutValueWithOwner(key, value, user, user)
}

// Put a value into the cache, using another user as the owner.
//
// This requires the user to have the `All.Insert` privilege.
func (kvc *KeyValueCache) PutValueWithOwner(key string, value []byte, owner *users.User, user *users.User) error {
	if user.Email == "" || !user.CanInsert(owner.Email == user.Email) {
		return errorMessages.ErrNotPermitted
	}

	return kvc.Put(key, KeyValueDelivery{
		Value: value,
		Timestamps: KeyValueTimestamps{
			CreatedAt: time.Now().UTC(),
		},
		Ownership: KeyValueOwnership{
			Email: &owner.Email,
			Name:  &owner.Name,
		},
	})
}

// Update an object in the cache.
//
// If the key does not exist, this will return an error.
func (kvc *KeyValueCache) Update(
	key string,
	value KeyValueDelivery,
) error {
	return kvc.LockAndDo(
		key,
		func(delivery *KeyValueDelivery) error {
			delivery.Value = value.Value
			delivery.Timestamps.CreatedAt = time.Now().UTC()
			return nil
		},
	)
}

// Update a value in the cache, with the user as the owner.
//
// This is a high level function that will check if the user has permission to update the object.
func (kvc *KeyValueCache) UpdateValue(
	key string,
	value []byte,
	user *users.User,
) error {
	return kvc.LockAndDo(
		key,
		func(delivery *KeyValueDelivery) error {
			owner := delivery.Ownership
			if owner.Email == nil || user.CanUpdate(owner.Email == &user.Email) {
				delivery.Value = value
				delivery.Timestamps.CreatedAt = time.Now().UTC()

				return nil
			} else {
				return errorMessages.ErrNotPermitted
			}
		},
	)
}

// Put or update an object in the cache.
//
// If the key already exists, this will update the object; otherwise, it will create a new one.
func (kvc *KeyValueCache) PutOrUpdate(key string, value KeyValueDelivery) error {
	// Attempt to create the key; if it already exists, update it instead
	if err := kvc.Put(key, value); err != nil {
		return kvc.Update(key, value)
	}

	return nil
}

// Put or update a value in the cache.
func (kvc *KeyValueCache) PutOrUpdateValue(
	key string,
	value []byte,
	user *users.User,
) error {
	// Attempt to create the key; if it already exists, update it instead
	if err := kvc.PutValue(key, value, user); err != nil {
		return kvc.UpdateValue(key, value, user)
	}

	return nil
}

// Delete an object from the cache.
//
// If the key does not exist, this will return an error.
// This is a low level function that does not check any user permissions.
func (kvc *KeyValueCache) Delete(key string) error {
	kvc.lock.Lock()
	defer kvc.lock.Unlock()

	if _, ok := kvc.Contents[key]; !ok {
		return errorMessages.ErrKeyNotFound
	}

	delete(kvc.Contents, key)

	return nil
}

// Delete a value from the cache.
//
// This is a high level function that will check if the user has permission to delete the object.
func (kvc *KeyValueCache) DeleteValue(
	key string,
	user *users.User,
) (*KeyValueDelivery, error) {
	if delivery, err := kvc.Get(key); err != nil {
		return &KeyValueDelivery{}, err
	} else {
		owner := delivery.Ownership
		if owner.Email == nil || user.CanDelete(owner.Email == &user.Email) {
			if err := kvc.Delete(key); err != nil {
				return &KeyValueDelivery{}, err
			} else {
				return &delivery, nil
			}
		} else {
			return &KeyValueDelivery{}, errorMessages.ErrNotPermitted
		}
	}
}
