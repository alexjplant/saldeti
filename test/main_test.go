//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/saldeti/saldeti/internal/auth"
)

func TestMain(m *testing.M) {
	auth.SetSigningKey([]byte("test-signing-key-32-bytes-long-!!"))
	os.Exit(m.Run())
}
