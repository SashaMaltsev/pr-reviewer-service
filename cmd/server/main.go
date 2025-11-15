package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SashaMalcev/pr-reviewer-service/internal/config"
	"github.com/SashaMalcev/pr-reviewer-service/internal/http/router"
	"github.com/SashaMalcev/pr-reviewer-service/internal/repository/postgres"
	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*

Main application entry point with graceful shutdown.
Initializes logger, config, database, services and HTTP server.
Handles OS signals for clean shutdown.

*/

func main() {

	// Init logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)

	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	// Initialize database connection
	ctx := context.Background()

	dbConfig := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	pool, err := pgxpool.New(ctx, dbConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	defer pool.Close()

	// Ping database
	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	log.Info().Msg("Successfully connected to database")

	// Init repositories
	teamRepo := postgres.NewTeamRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	prRepo := postgres.NewPRRepository(pool)

	// Init services
	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)
	statsService := service.NewStatsService(prRepo, userRepo)

	// Init HTTP router
	r := router.New(teamService, userService, prService, statsService)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Starting HTTP server
	go func() {
		log.Info().Str("port", cfg.ServerPort).Msg("Starting HTTP server")

		err := server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
