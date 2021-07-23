package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"library/v1/pb"
	"log"
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

	// Search laptop from store
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
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
	//other := &pb.Laptop{}
	//err := copier.Copy(other, laptop)
	other, err := deepCopy(laptop)
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
	//other := &pb.Laptop{}
	//err := copier.Copy(other, laptop)
	//other, err := deepcopy(laptop)
	//if err != nil {
	//	return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	//}

	return deepCopy(laptop)
}

// Search according to filter find laptop
func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, laptop := range store.data {
		//time.Sleep(time.Second)
		//log.Println("checking laptop id: ", laptop.GetId())

		// Deadline exceeded control
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			log.Println("context is cancel")
			return errors.New("context is cancel")
		}

		// Filter laptop
		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Filter
func isQualified(f *pb.Filter, l *pb.Laptop) bool {
	if l.GetPriceUsd() > f.GetMaxPriceUsd() {
		return false
	}
	if l.GetCpu().GetNumCores() < f.MinCpuCores {
		return false
	}
	if l.GetCpu().GetMinGhz() < f.MinCpuGhz {
		return false
	}
	if toBit(l.GetRam()) < toBit(f.MinRam) {
		return false
	}
	return true
}

// To transfer bit unit,
func toBit(m *pb.Memory) uint64 {
	value := m.GetValue()
	switch m.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8bit = 1kb
	case pb.Memory_KILOBYTE:
		return value << 13
	case pb.Memory_MEGABYTE:
		return value << 23 // 1024 = 2^10
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(l *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, l)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}
	return other, nil
}
