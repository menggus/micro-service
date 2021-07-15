package service

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"library/v1/pb"
	"sync"
)

// ErrAlreadyExists is returned when a record with the same ID already exists in the store
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore is an interface to store laptop
type LaptopStore interface {
	// Save method saves the laptop to the store
	Save(laptop *pb.Laptop) error

	// Find laptop from store
	Find(id string) (*pb.Laptop, error)
}

type InMemoryLaptopStore struct {
	mutex sync.RWMutex          // concurrency security
	data  map[string]*pb.Laptop // save the *laptop
}

// NewInMemoryLaptopStore create a InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Save implement the LaptopStore interface
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	// concurrency add lock
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// Check laptop.ID
	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	// Save the laptop
	// Security!!!, deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("cannot copy laptop data:%w", err)
	}
	store.data[other.Id] = other
	return nil
}

// Find according to id find laptop
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	// Security deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}
