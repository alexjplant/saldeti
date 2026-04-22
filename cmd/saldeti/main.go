package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/auth"
	"github.com/saldeti/saldeti/internal/handler"
	"github.com/saldeti/saldeti/internal/seed"
	"github.com/saldeti/saldeti/internal/store"
	ui "github.com/saldeti/saldeti/internal/ui"
)

func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Saldeti Simulator"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func main() {
	port := flag.Int("port", 9443, "Port to listen on")
	uiEnabled := flag.Bool("ui", true, "Enable admin UI")
	seedPath := flag.String("seed", "", "Path to JSON seed file (optional)")
	dumpPath := flag.String("dump", "", "Path to write seed JSON on shutdown (optional)")
	signingKey := flag.String("signing-key", "", "JWT signing key (default: random per startup)")
	tlsCert := flag.String("tls-cert", "", "TLS certificate file")
	tlsKey := flag.String("tls-key", "", "TLS key file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	domain := flag.String("domain", "saldeti.local", "Default directory domain (used for admin user UPN and seeded users without an @)")
	stop := flag.Bool("stop", false, "Stop a running daemon")

	// Admin credential flags
	adminClientID := flag.String("admin-client-id", "", "Admin app client ID (default: random UUID; if set, -admin-client-secret and -admin-tenant-id must also be set)")
	adminClientSecret := flag.String("admin-client-secret", "", "Admin app client secret (default: random UUID; if set, -admin-client-id and -admin-tenant-id must also be set)")
	adminTenantID := flag.String("admin-tenant-id", "", "Admin app tenant ID (default: random UUID; if set, -admin-client-id and -admin-client-secret must also be set)")

	// Daemon mode flags
	daemon := flag.Bool("daemon", false, "Run as background daemon")
	pidfile := flag.String("pidfile", "saldeti.pid", "Path to write PID file (daemon mode)")
	logfile := flag.String("logfile", "saldeti.log", "Path to log file (daemon mode)")

	flag.Parse()

	// Stop daemon mode: read PID file and send SIGTERM
	if *stop {
		data, err := os.ReadFile(*pidfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read PID file %s: %v\n", *pidfile, err)
			os.Exit(1)
		}

		pidStr := strings.TrimSpace(string(data))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid PID in %s: %s\n", *pidfile, pidStr)
			os.Exit(1)
		}

		// Find the process
		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find process %d: %v\n", pid, err)
			os.Exit(1)
		}

		// Send SIGTERM
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send SIGTERM to process %d: %v\n", pid, err)
			os.Exit(1)
		}

		// Wait for process to exit (up to 10 seconds)
		fmt.Fprintf(os.Stderr, "Stopping server (PID %d)...\n", pid)
		for i := 0; i < 100; i++ {
			if err := proc.Signal(syscall.Signal(0)); err != nil {
				// Process has exited
				os.Remove(*pidfile)
				fmt.Fprintf(os.Stderr, "Server stopped\n")
				os.Exit(0)
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Force kill if still running
		fmt.Fprintf(os.Stderr, "Server did not stop gracefully, killing...\n")
		proc.Kill()
		os.Remove(*pidfile)
		os.Exit(0)
	}

	// Configure default HTTP transport to trust self-signed cert.
	// This is a local dev simulator that only talks to itself on localhost.
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// Validate admin credential flags: all-or-nothing
	adminFlagsSet := 0
	if *adminClientID != "" {
		adminFlagsSet++
	}
	if *adminClientSecret != "" {
		adminFlagsSet++
	}
	if *adminTenantID != "" {
		adminFlagsSet++
	}
	if adminFlagsSet != 0 && adminFlagsSet != 3 {
		log.Fatal().Msg("If any admin credential flag is set, all three must be set: -admin-client-id, -admin-client-secret, -admin-tenant-id")
	}

	// Daemon mode: fork child and exit
	if *daemon && os.Getenv("SALDETI_CHILD") != "1" {
		// Check if a daemon is already running by examining the PID file
		if pidData, err := os.ReadFile(*pidfile); err == nil {
			existingPID := strings.TrimSpace(string(pidData))
			if pid, err := strconv.Atoi(existingPID); err == nil {
				if proc, err := os.FindProcess(pid); err == nil {
					if err := proc.Signal(syscall.Signal(0)); err == nil {
						fmt.Fprintf(os.Stderr, "Daemon already running with PID %d (pidfile: %s)\n  Stop it first: %s -stop -pidfile %s\n",
							pid, *pidfile, os.Args[0], *pidfile)
						os.Exit(1)
					}
				}
			}
			// PID file exists but process is not running — stale pidfile, remove it
			os.Remove(*pidfile)
		}
		lf, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}

		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), "SALDETI_CHILD=1")
		cmd.Stdout = lf
		cmd.Stderr = lf
		cmd.Stdin = nil

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(*pidfile, []byte(fmt.Sprintf("%d\n", cmd.Process.Pid)), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write PID file: %v\n", err)
			cmd.Process.Kill()
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Daemon started with PID %d\n  Log: %s\n  PID: %s\n  Stop: %s -stop -pidfile %s\n",
		cmd.Process.Pid, *logfile, *pidfile, os.Args[0], *pidfile)
		lf.Close()
		os.Exit(0)
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Str("domain", *domain).Msg("Directory domain")

	// Set JWT signing key
	if *signingKey == "" {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			log.Fatal().Err(err).Msg("Failed to generate random signing key")
		}
		auth.SetSigningKey(key)
		// Log the generated key so developers can reuse it
		log.Info().Str("key", hex.EncodeToString(key)).Msg("Generated random JWT signing key")
	} else {
		auth.SetSigningKey([]byte(*signingKey))
	}

	// Create store
	store := store.NewMemoryStore()

	// Seed data from JSON file if provided
	if *seedPath != "" {
		cfg, err := seed.LoadFromFile(*seedPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load seed file")
		}
		if err := seed.SeedFromConfig(store, cfg); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed data")
		}
		if len(cfg.Clients) > 0 {
			log.Info().Str("client_id", cfg.Clients[0].ClientID).Msg("Seeded client")
			log.Info().Str("tenant_id", cfg.Clients[0].TenantID).Msg("Seeded tenant")
		}
		log.Info().Int("count", len(cfg.Users)).Msg("Seeded users")
	}

	// Always create an admin app for UI and API access
	ctx := context.Background()
	var finalAdminClientID, finalAdminClientSecret, finalAdminTenantID string
	if adminFlagsSet == 3 {
		finalAdminClientID = *adminClientID
		finalAdminClientSecret = *adminClientSecret
		finalAdminTenantID = *adminTenantID
	} else {
		finalAdminClientID = uuid.New().String()
		finalAdminClientSecret = uuid.New().String()
		finalAdminTenantID = uuid.New().String()
	}
	if err := store.RegisterClient(ctx, finalAdminClientID, finalAdminClientSecret, finalAdminTenantID); err != nil {
		log.Fatal().Err(err).Msg("Failed to register admin client")
	}
	log.Info().Str("client_id", finalAdminClientID).Str("client_secret", finalAdminClientSecret).Str("tenant_id", finalAdminTenantID).Msg("Admin app credentials")

	// Create router
	router := handler.NewRouter(store)

	// Register UI routes if enabled
	if *uiEnabled {
		baseURL := fmt.Sprintf("https://localhost:%d", *port)
		ui.RegisterUIRoutes(router, baseURL, finalAdminClientID, finalAdminClientSecret, finalAdminTenantID)
	}

	// Generate self-signed TLS cert if not provided
	addr := fmt.Sprintf(":%d", *port)
	var tlsCertObj *tls.Certificate
	if *tlsCert != "" && *tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load TLS certificate")
		}
		tlsCertObj = &cert
		log.Info().Str("addr", addr).Msg("Starting HTTPS server")
	} else {
		cert, err := generateSelfSignedCert()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate self-signed certificate")
		}
		tlsCertObj = &cert
		log.Info().Str("addr", addr).Msg("Starting HTTPS server (self-signed certificate)")
	}

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*tlsCertObj},
		},
	}

	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Start refresh token cleanup
	auth.StartRefreshTokenCleanup(context.Background(), 10*time.Minute)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Dump store if -dump is set
	if *dumpPath != "" {
		cfg, err := seed.DumpStore(store)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to dump store")
		} else {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				log.Warn().Err(err).Msg("Failed to marshal dump")
			} else if err := os.WriteFile(*dumpPath, data, 0600); err != nil {
				log.Warn().Err(err).Msg("Failed to write dump file")
			} else {
				log.Info().Str("path", *dumpPath).Msg("Store dumped")
			}
		}
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}
	log.Info().Msg("Server stopped")
}
