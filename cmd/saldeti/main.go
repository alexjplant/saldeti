package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/saldeti/saldeti/internal/auth"
	"github.com/saldeti/saldeti/internal/handler"
	"github.com/saldeti/saldeti/internal/model"
	"github.com/saldeti/saldeti/internal/seed"
	"github.com/saldeti/saldeti/internal/store"
	ui "github.com/saldeti/saldeti/internal/ui"
)

func main() {
	port := flag.Int("port", 9443, "Port to listen on")
	uiEnabled := flag.Bool("ui", true, "Enable admin UI")
	seedPath := flag.String("seed", "", "Path to JSON seed file (optional)")
	dumpPath := flag.String("dump", "", "Path to write seed JSON on shutdown (optional)")
	signingKey := flag.String("signing-key", "", "JWT signing key (default: random per startup)")
	tlsCert := flag.String("tls-cert", "", "TLS certificate file")
	tlsKey := flag.String("tls-key", "", "TLS key file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

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
		if len(cfg.Users) > 0 {
			log.Info().Str("email", cfg.Users[0].Email).Msg("Seeded admin user")
		}
	}

	// Check if store is empty and show warning
	ctx := context.Background()
	users, _, _ := store.ListUsers(ctx, model.ListOptions{Top: 1})
	clients, _ := store.ListClients(ctx)
	if len(users) == 0 && len(clients) == 0 {
		log.Warn().Msg("Store is empty. All API calls requiring auth will fail. Pass -seed to load data.")
	}

	// Create router
	router := handler.NewRouter(store)

	// Register UI routes if enabled
	if *uiEnabled {
		ui.RegisterUIRoutes(router, store, *port)
	}

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if *tlsCert != "" && *tlsKey != "" {
			log.Info().Str("addr", addr).Bool("tls", true).Msg("Starting HTTPS server")
			if err := srv.ListenAndServeTLS(*tlsCert, *tlsKey); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Server failed")
			}
		} else {
			log.Info().Str("addr", addr).Bool("tls", false).Msg("Starting HTTP server")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Server failed")
			}
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
