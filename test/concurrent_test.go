//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// TestE2E_ConcurrentRequests tests concurrent operations to ensure no data corruption
func TestE2E_ConcurrentRequests(t *testing.T) {
	tss := setupTestServer(t)
	defer tss.Server.Close()

	// Number of concurrent goroutines
	numGoroutines := 10

	// Channel to collect errors
	errors := make(chan error, numGoroutines)
	var wg sync.WaitGroup

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine creates, reads, updates, and deletes a user via SDK
			userID, err := performUserLifecycleSDK(t, tss, id)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: %v", id, err)
				return
			}

			// Verify the user was actually deleted via SDK
			ctx := context.Background()
			_, err = tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
			if err == nil {
				// User still exists, which means it wasn't deleted properly
				errors <- fmt.Errorf("goroutine %d: user %s was not deleted", id, userID)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Test failed with %d errors", errorCount)
	}
}

// performUserLifecycleSDK creates, reads, updates, and deletes a user via SDK
// Returns the user ID and any error encountered
func performUserLifecycleSDK(t *testing.T, tss *TestServer, id int) (string, error) {
	ctx := context.Background()

	// 1. Create a user via SDK
	user := models.NewUser()
	dn := fmt.Sprintf("Concurrent User %d-%d", id, time.Now().UnixNano())
	upn := fmt.Sprintf("concurrent%d-%d@test.local", id, time.Now().UnixNano())
	user.SetDisplayName(&dn)
	user.SetUserPrincipalName(&upn)
	user.SetMail(&upn)
	enabled := true
	user.SetAccountEnabled(&enabled)

	passwordProfile := models.NewPasswordProfile()
	password := "Test1234!"
	passwordProfile.SetPassword(&password)
	user.SetPasswordProfile(passwordProfile)

	created, err := tss.SDKClient.Users().Post(ctx, user, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}

	userID := *created.GetId()

	// 2. Read the user via SDK
	fetched, err := tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
	if err != nil {
		return userID, fmt.Errorf("failed to read user: %v", err)
	}

	if fetched.GetDisplayName() == nil || *fetched.GetDisplayName() != dn {
		return userID, fmt.Errorf("displayName mismatch: expected '%s', got '%v'", dn, fetched.GetDisplayName())
	}

	// 3. Update the user via SDK
	patch := models.NewUser()
	dept := "Updated"
	patch.SetDepartment(&dept)
	_, err = tss.SDKClient.Users().ByUserId(userID).Patch(ctx, patch, nil)
	if err != nil {
		return userID, fmt.Errorf("failed to update user: %v", err)
	}

	// 4. Read again to verify update via SDK
	updated, err := tss.SDKClient.Users().ByUserId(userID).Get(ctx, nil)
	if err != nil {
		return userID, fmt.Errorf("failed to read user after update: %v", err)
	}

	if updated.GetDepartment() == nil || *updated.GetDepartment() != dept {
		return userID, fmt.Errorf("department not updated: expected '%s', got '%v'", dept, updated.GetDepartment())
	}

	// 5. Delete the user via SDK
	err = tss.SDKClient.Users().ByUserId(userID).Delete(ctx, nil)
	if err != nil {
		return userID, fmt.Errorf("failed to delete user: %v", err)
	}

	return userID, nil
}
