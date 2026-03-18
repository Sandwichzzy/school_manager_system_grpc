package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/api/handlers"
	"github.com/Sandwichzzy/school_manager_system_grpc/internals/api/interceptors"
	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	creds, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		log.Fatalf("Failed to load TLS certificates")
	}

	rateLimiter := interceptors.NewRateLimiter(50, time.Minute)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(rateLimiter.RateLimitInterceptor, interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor), grpc.Creds(creds))

	pb.RegisterExecsServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterTeachersServiceServer(s, &handlers.Server{})

	reflection.Register(s)
	// 监听端口，启动服务器等逻辑
	port := os.Getenv("SERVER_PORT")

	go utils.JwtStore.CleanUpExpiredTokens()

	fmt.Println("grpc server running on port:", port)
	// 监听端口
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 启动 gRPC 服务器
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
