package service

import "sync"

// UserStore is an interface to store user
type UserStore interface {
	// Save saves a user in the store
	Save(user *User) error
	// Find a user from the store by username
	Find(username string) (*User, error)
}

// InMemoryUserStore store users in memory
type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

// NewInMemoryUserStore return a new in-memory store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// Save save new user to the store
func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.users[user.UserName] != nil {
		return ErrAlreadyExists
	}

	store.users[user.UserName] = user.Clone()

	return nil
}

// Find find a user from the user store
func (store *InMemoryUserStore) Find(username string) (*User, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	user := store.users[username]
	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}
