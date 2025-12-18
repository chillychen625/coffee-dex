package storage

import (
	"errors"
	"go-coffee-log/models"
	"sync"
)

// MemoryStorage implements CoffeeStorage using an in-memory map
type MemoryStorage struct {
	coffees map[string]models.Coffee
	mu sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		coffees: make(map[string]models.Coffee),
	}
}

// Save stores a new coffee entry
func (m *MemoryStorage) Save(coffee models.Coffee) error {
	if (m == nil) {
		return errors.New("memory storage is not initialized")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coffees[coffee.ID] = coffee
	
	return nil
}

// GetByID retrieves a coffee by ID
func (m *MemoryStorage) GetByID(id string) (models.Coffee, error) {
	if m == nil {
		return models.Coffee{}, errors.New("memory storage is not initialized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	coffee, ok := m.coffees[id]
	if !ok {
		return models.Coffee{}, errors.New("coffee not found")
	}
	return coffee, nil
}

// GetAll retrieves all coffees
func (m *MemoryStorage) GetAll() ([]models.Coffee, error) {
	if m == nil {
		return nil, errors.New("memory storage is not initialized")
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var coffees []models.Coffee
	for _, coffee := range m.coffees {
		coffees = append(coffees, coffee)
	}
	
	return coffees, nil
}

// GetRecent retrieves the most recent coffees (sorted by creation date)
func (m *MemoryStorage) GetRecent(limit int) ([]models.Coffee, error) {
	if m == nil {
		return nil, errors.New("memory storage is not initialized")
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var coffees []models.Coffee
	for _, coffee := range m.coffees {
		coffees = append(coffees, coffee)
	}
	
	// Sort by creation date descending
	for i := 0; i < len(coffees)-1; i++ {
		for j := i + 1; j < len(coffees); j++ {
			if coffees[j].CreatedAt.After(coffees[i].CreatedAt) {
				coffees[i], coffees[j] = coffees[j], coffees[i]
			}
		}
	}
	
	// Limit the results
	if limit > 0 && limit < len(coffees) {
		coffees = coffees[:limit]
	}
	
	return coffees, nil
}

// Update modifies an existing coffee entry
func (m *MemoryStorage) Update(id string, coffee models.Coffee) error {
	if m == nil {
		return errors.New("memory storage is not initialized")
	}

	if _, ok := m.coffees[id]; !ok {
		return errors.New("coffee not found")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.coffees[id] = coffee
	return nil
}

// Delete removes a coffee entry
func (m *MemoryStorage) Delete(id string) error {
	if m == nil {
		return errors.New("memory storage is not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.coffees, id)
	return nil
}