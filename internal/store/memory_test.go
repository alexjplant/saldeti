package store

import (
	"context"
	"testing"
	"time"

	"github.com/saldeti/saldeti/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_UserOperations(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Test CreateUser
	accountEnabled := true
	user := model.User{
		DisplayName:       "Test User",
		UserPrincipalName: "test@example.com",
		Mail:              "test@example.com",
		AccountEnabled:    &accountEnabled,
	}

	createdUser, err := store.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, createdUser.ID)
	assert.NotNil(t, createdUser.CreatedDateTime)
	assert.WithinDuration(t, time.Now(), *createdUser.CreatedDateTime, time.Second)

	// Test GetUser
	retrievedUser, err := store.GetUser(ctx, createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, "test@example.com", retrievedUser.UserPrincipalName)
	assert.Equal(t, "Test User", retrievedUser.DisplayName)

	// Test GetUserByUPN
	retrievedByUPN, err := store.GetUserByUPN(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedByUPN.ID)

	// Test duplicate user
	_, err = store.CreateUser(ctx, user)
	assert.ErrorIs(t, err, ErrDuplicateUPN)

	// Test GetUser with non-existent ID
	_, err = store.GetUser(ctx, "non-existent-id")
	assert.ErrorIs(t, err, ErrUserNotFound)

	// Test GetUserByUPN with non-existent UPN
	_, err = store.GetUserByUPN(ctx, "nonexistent@example.com")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestMemoryStore_ClientOperations(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Test RegisterClient
	err := store.RegisterClient(ctx, "client1", "secret1", "tenant1")
	require.NoError(t, err)

	// Test GetClient
	clientID, secret, tenantID, err := store.GetClient(ctx, "client1")
	require.NoError(t, err)
	assert.Equal(t, "client1", clientID)
	assert.Equal(t, "secret1", secret)
	assert.Equal(t, "tenant1", tenantID)

	// Test duplicate client registration
	err = store.RegisterClient(ctx, "client1", "secret2", "tenant2")
	assert.ErrorIs(t, err, ErrDuplicateClient)

	// Test GetClient with non-existent client
	_, _, _, err = store.GetClient(ctx, "non-existent-client")
	assert.ErrorIs(t, err, ErrClientNotFound)
}