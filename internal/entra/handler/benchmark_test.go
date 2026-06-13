package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saldeti/saldeti/internal/entra/auth"
	"github.com/saldeti/saldeti/internal/entra/model"
	"github.com/saldeti/saldeti/internal/entra/store"
)

// BenchmarkBatchHandler benchmarks the batch handler with varying request counts.
func BenchmarkBatchHandler(b *testing.B) {
	s := store.NewMemoryStore()
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		enabled := true
		if _, err := s.CreateUser(ctx, model.User{
			DisplayName:       fmt.Sprintf("BenchUser %d", i),
			UserPrincipalName: fmt.Sprintf("benchuser%d@example.com", i),
			Mail:              fmt.Sprintf("benchuser%d@example.com", i),
			AccountEnabled:    &enabled,
		}); err != nil {
			b.Fatal(err)
		}
	}

	router := NewRouter(s)
	server := httptest.NewServer(router)
	defer server.Close()

	token, err := auth.MintToken("test-tenant", "test-client", "admin@example.com", []string{"User.Read.All"}, []string{"User"}, time.Hour, "", "")
	if err != nil {
		b.Fatal(err)
	}

	for _, count := range []int{1, 5, 10} {
		b.Run(fmt.Sprintf("%dRequests", count), func(b *testing.B) {
			requests := make([]map[string]interface{}, count)
			for i := 0; i < count; i++ {
				requests[i] = map[string]interface{}{
					"id":     fmt.Sprintf("%d", i+1),
					"method": "GET",
					"url":    "/v1.0/users",
				}
			}
			batchBody, _ := json.Marshal(map[string]interface{}{
				"requests": requests,
			})

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req, err := http.NewRequest("POST", server.URL+"/v1.0/$batch", bytes.NewReader(batchBody))
			if err != nil {
				b.Fatal(err)
			}
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				resp.Body.Close()
			}
		})
	}
}
