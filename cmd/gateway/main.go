package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/kijexs/github-repository-indexer/docs"
	"github.com/kijexs/github-repository-indexer/internal/config"
	"github.com/kijexs/github-repository-indexer/internal/gateway/adapter"
	"github.com/kijexs/github-repository-indexer/internal/gateway/handler"
	"github.com/kijexs/github-repository-indexer/internal/gateway/usecase"
	"github.com/kijexs/github-repository-indexer/proto/repopb"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Repository API
// @version 1.0
// @description API Gateway for GitHub repository information
// @host localhost:8080
// @BasePath /
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	requestTimeout, err := config.Duration("REQUEST_TIMEOUT", 70*time.Second)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	collectorAddr := os.Getenv("COLLECTOR_ADDR")
	if collectorAddr == "" {
		collectorAddr = "localhost:50051"
	}
	conn, err := grpc.NewClient(collectorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create grpc client: %w", err)
	}
	defer conn.Close()

	client := repopb.NewRepositoryServiceClient(conn)
	collectorAdapter := adapter.NewCollectorAdapter(client)

	useCase := usecase.NewGatewayUseCase(collectorAdapter)

	repoHandler := handler.NewRepositoryHandler(useCase)

	mux := http.NewServeMux()

	mux.Handle("GET /repositories/{owner}/{repo}", repoHandler)
	mux.HandleFunc("POST /repositories", repoHandler.LoadRepositories)

	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	server := &http.Server{
		Addr: config.String("HTTP_ADDR", ":8080"),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCtx, cancel := context.WithTimeout(r.Context(), requestTimeout)
			defer cancel()
			mux.ServeHTTP(w, r.WithContext(requestCtx))
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       requestTimeout,
		WriteTimeout:      requestTimeout + 5*time.Second,
		IdleTimeout:       time.Minute,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			server.Close()
		case <-done:
		}
	}()
	defer close(done)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}
