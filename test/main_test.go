//go:build e2e && entra

package e2e

import (
	"os"
	"testing"

	"github.com/saldeti/saldeti/internal/entra/auth"
)

func TestMain(m *testing.M) {
	auth.SetSigningKey([]byte("test-signing-key-32-bytes-long-!!"))
	os.Exit(m.Run())
}
