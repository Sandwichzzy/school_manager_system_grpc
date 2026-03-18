package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Sandwichzzy/school_manager_system_grpc/internals/api/handlers"
	"github.com/Sandwichzzy/school_manager_system_grpc/internals/api/interceptors"
	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
	pb "github.com/Sandwichzzy/school_manager_system_grpc/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

// 本质原因：godotenv.Load() 只能读“文件路径”，不能读“内存数据”
// embed.FS (内存)
//
//	↓
//
// 写入 temp file（磁盘）
//
//	↓
//
// godotenv.Load(文件路径)
func loadEnvFromEmbeddedFile() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file:%v", err)
	}

	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating .env file:%v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing to temp file:%v", err)
	}

	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error closing write temp file:%v", err)
	}

	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file:%v", err)
	}
}

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
	loadEnvFromEmbeddedFile()

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	creds, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		log.Fatalf("Failed to load TLS certificates")
	}

	// rateLimiter := interceptors.NewRateLimiter(50, time.Minute)
	// s := grpc.NewServer(grpc.ChainUnaryInterceptor(rateLimiter.RateLimitInterceptor, interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor), grpc.Creds(creds))
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor), grpc.Creds(creds))

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
