package grpc

import (
	"fmt"
	gRPCservice "jwt2/gRPCService/src"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/middleware"
	repo "jwt2/repo/src"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func StartGRPCServer(repo *repo.TaskRepo, RedisClient repo.RedisClientInterface) {
	grpcPort := os.Getenv("APP_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	address := fmt.Sprintf(":%s", grpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(middleware.AuthInterceptor("ADMIN", "USER"),

		middleware.RoleInterceptor(map[string][]string{
			"/task.v1.TaskService/CreateTask":     {"USER"},
			"/task.v1.TaskService/GetTask":        {"USER"},
			"/task.v1.TaskService/GetTaskById":    {"USER"},
			"/task.v1.TaskService/DeleteTaskByID": {"USER"},
			"/task.v1.TaskService/UpdateTask":     {"USER"},
		}),
	),
		grpc.ChainStreamInterceptor(middleware.AuthStreamInterceptor("ADMIN", "USER")))

	taskHandler := gRPCservice.NewTaskHandler(repo, RedisClient)
	v1.RegisterTaskServiceServer(grpcServer, taskHandler)
	log.Printf("gRPC server is running on port %v...", grpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
