package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/saldeti/saldeti/internal/entra/model"
)

// BenchmarkApplyOData_FilterParsing benchmarks OData filter parsing at various complexity levels.
// Creates a slice of 100 model.User structs, then runs sub-benchmarks for each filter expression.
func BenchmarkApplyOData_FilterParsing(b *testing.B) {
	users := make([]model.User, 100)
	for i := range users {
		enabled := i%2 == 0
		userType := "Member"
		if i%3 == 0 {
			userType = "Guest"
		}
		users[i] = model.User{
			ODataType:         "#microsoft.graph.user",
			ID:                fmt.Sprintf("user-%03d", i),
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &enabled,
			UserType:          userType,
		}
	}

	filters := []struct {
		name   string
		filter string
	}{
		{"SimpleEq", "displayName eq 'User 50'"},
		{"Ne", "userType ne 'Member'"},
		{"And", "userType eq 'Member' and accountEnabled eq true"},
		{"Or", "userType eq 'Guest' or displayName eq 'User 5'"},
		{"Startswith", "startswith(displayName,'User')"},
		{"Contains", "contains(mail,'example')"},
		{"Complex", "accountEnabled eq true and (userType eq 'Member' or startswith(displayName,'User 1'))"},
	}

	for _, f := range filters {
		b.Run(f.name, func(b *testing.B) {
			opts := model.ListOptions{Filter: f.filter}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = ApplyOData(users, opts)
			}
		})
	}
}

// BenchmarkStore_ListUsers benchmarks ListUsers through the Store interface.
// Seeds 1000 users then runs sub-benchmarks with different ListOptions.
func BenchmarkStore_ListUsers(b *testing.B) {
	ctx := context.Background()
	s := NewMemoryStore()

	for i := 0; i < 1000; i++ {
		enabled := i%2 == 0
		if _, err := s.CreateUser(ctx, model.User{
			ODataType:         "#microsoft.graph.user",
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &enabled,
			UserType:          "Member",
		}); err != nil {
			b.Fatal(err)
		}
	}

	scenarios := []struct {
		name string
		opts model.ListOptions
	}{
		{"NoFilter", model.ListOptions{}},
		{"FilterDisplayName", model.ListOptions{Filter: "displayName eq 'User 500'"}},
		{"FilterAndTop10", model.ListOptions{Filter: "accountEnabled eq true", Top: 10}},
		{"FilterOrderByTop", model.ListOptions{Filter: "accountEnabled eq true", OrderBy: "displayName", Top: 20}},
		{"FilterSelect", model.ListOptions{Filter: "accountEnabled eq true", Select: []string{"displayName", "mail"}}},
		{"Search", model.ListOptions{Search: "User 50"}},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = s.ListUsers(ctx, sc.opts)
			}
		})
	}
}

// BenchmarkApplyOData_FullPipeline benchmarks the full OData pipeline on 1000 users.
func BenchmarkApplyOData_FullPipeline(b *testing.B) {
	users := make([]model.User, 1000)
	for i := range users {
		enabled := i%2 == 0
		users[i] = model.User{
			ODataType:         "#microsoft.graph.user",
			ID:                fmt.Sprintf("user-%03d", i),
			DisplayName:       fmt.Sprintf("User %d", i),
			UserPrincipalName: fmt.Sprintf("user%d@example.com", i),
			Mail:              fmt.Sprintf("user%d@example.com", i),
			AccountEnabled:    &enabled,
			UserType:          "Member",
		}
	}

	opts := model.ListOptions{
		Filter:  "accountEnabled eq true",
		OrderBy: "displayName",
		Top:     50,
		Skip:    10,
		Select:  []string{"id", "displayName", "mail"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ApplyOData(users, opts)
	}
}
