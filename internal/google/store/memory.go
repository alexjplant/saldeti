package store

import "context"

// MemoryStore implements Store with in-memory storage.
type MemoryStore struct{}

// NewMemoryStore creates a new in-memory store for Google Workspace data.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Ping verifies the store is accessible.
func (s *MemoryStore) Ping(ctx context.Context) error {
	return nil
}