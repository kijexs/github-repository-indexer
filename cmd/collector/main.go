package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/kijexs/github-repository-indexer/internal/collector/adapter/github"
	"github.com/kijexs/github-repository-indexer/internal/collector/adapter/postgres"
	grpchandler "github.com/kijexs/github-repository-indexer/internal/collector/handler/grpc"
	"github.com/kijexs/github-repository-indexer/internal/collector/migrations"
	"github.com/kijexs/github-repository-indexer/internal/collector/usecase"
	"github.com/kijexs/github-repository-indexer/internal/config"
	"github.com/kijexs/github-repository-indexer/proto/repopb"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	migrate := flag.String("migrate", "", "apply migrations and exit: up or down (one migration)")
	flag.Parse()
	if *migrate != "" && *migrate != "up" && *migrate != "down" {
		return fmt.Errorf("migrate must be up or down")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}
	workerCount, err := config.PositiveInt("WORKER_COUNT", 4)
	if err != nil {
		return err
	}
	githubTimeout, err := config.Duration("GITHUB_TIMEOUT", 10*time.Second)
	if err != nil {
		return err
	}
	batchTimeout, err := config.Duration("BATCH_TIMEOUT", time.Minute)
	if err != nil {
		return err
	}
	startupTimeout, err := config.Duration("STARTUP_TIMEOUT", 30*time.Second)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	startupCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	pool, err := pgxpool.New(startupCtx, databaseURL)
	if err != nil {
		return fmt.Errorf("create pgxpool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(startupCtx); err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	if *migrate == "down" {
		if err := goose.DownContext(startupCtx, db, "."); err != nil {
			return fmt.Errorf("rollback migration: %w", err)
		}
		return nil
	}
	if err := goose.UpContext(startupCtx, db, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if *migrate == "up" {
		return nil
	}
	cancel()
	httpClient := &http.Client{Timeout: githubTimeout}
	provider := github.NewGitHubAdapter(httpClient)
	repository := postgres.NewRepositoryRepository(pool)
	useCase := usecase.NewRepositoryUseCase(provider, repository, workerCount, batchTimeout)
	repoHandler := grpchandler.NewRepositoryHandler(useCase)
	listener, err := net.Listen("tcp", config.String("GRPC_ADDR", ":50051"))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()
	grpcServer := grpc.NewServer()
	repopb.RegisterRepositoryServiceServer(grpcServer, repoHandler)
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			grpcServer.Stop()
		case <-done:
		}
	}()
	defer close(done)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("serve gRPC: %w", err)
	}
	return nil
}
