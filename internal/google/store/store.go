package store

import "context"

// Store defines the interface for Google Workspace data storage.
type Store interface {
	// Ping verifies the store is accessible.
	Ping(ctx context.Context) error
}