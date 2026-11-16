package main

import (
	_ "github.com/kijexs/github-repository-indexer/docs"
	"github.com/kijexs/github-repository-indexer/internal/gateway/adapter"
	"github.com/kijexs/github-repository-indexer/internal/gateway/handler"
	"github.com/kijexs/github-repository-indexer/internal/gateway/usecase"
	"github.com/kijexs/github-repository-indexer/proto/repopb"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"net/http"
	"os"
)

// @title Repository API
// @version 1.0
// @description API Gateway for GitHub repository information
// @host localhost:8080
// @BasePath /
func main() {
	collectorAddr := os.Getenv("COLLECTOR_ADDR")
	if collectorAddr == "" {
		collectorAddr = "localhost:50051"
	}
	conn, err := grpc.NewClient(collectorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		grpclog.Fatalf("failed to create grpc client connection: %v", err)
		return
	}
	defer conn.Close()

	client := repopb.NewRepositoryServiceClient(conn)
	collectorAdapter := adapter.NewCollectorAdapter(client)

	useCase := usecase.NewGatewayUseCase(collectorAdapter)

	repoHandler := handler.NewRepositoryHandler(useCase)

	mux := http.NewServeMux()

	mux.Handle("GET /repositories/{owner}/{repo}", repoHandler)

	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	http.ListenAndServe(":8080", mux)
}
