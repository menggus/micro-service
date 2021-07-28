package service

import "sync"

// RateStore is a interface to store laptop rating
type RateStore interface {
	// Add a new laptop score to the store and returns its rating
	Add(laptopID string, score float64) (*Rating, error)
}

// Rating contains the rating information of laptop
type Rating struct {
	Count uint32
	Sum   float64
}

// InMemoryRatingStore is store the laptop rating
type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	// used a lock
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// Detect if laptopID exists and update rating
	rat := store.rating[laptopID]
	if rat == nil {
		rat = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rat.Count++
		rat.Sum += score
	}

	store.rating[laptopID] = rat

	return rat, nil
}
