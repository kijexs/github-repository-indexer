package main

import (
	"github.com/kijexs/github-repository-indexer/collector/adapter/github"
	grpchandler "github.com/kijexs/github-repository-indexer/collector/handler/grpc"
	"github.com/kijexs/github-repository-indexer/collector/usecase"
	"github.com/kijexs/github-repository-indexer/proto/repopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
	"net/http"
	"time"
)

func main() {
	var httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	gitHubAdapter := github.NewGitHubAdapter(httpClient)

	useCase := usecase.NewRepositoryUseCase(gitHubAdapter)

	repoHandler := grpchandler.NewRepositoryHandler(useCase)

	network := "tcp"
	address := ":50051"

	listener, err := net.Listen(network, address)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()

	repopb.RegisterRepositoryServiceServer(grpcServer, repoHandler)
	err = grpcServer.Serve(listener)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
}
