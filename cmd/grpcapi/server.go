package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/api/handlers"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	s := grpc.NewServer()

	pb.RegisterExecsServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterTeachersServiceServer(s, &handlers.Server{})

	reflection.Register(s)
	// 监听端口，启动服务器等逻辑
	port := os.Getenv("SERVER_PORT")

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
