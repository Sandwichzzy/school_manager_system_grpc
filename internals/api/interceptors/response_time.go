package interceptors

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ResponseTimeInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	log.Println("ResponseTimeInterceptor Ran")

	start := time.Now() //record the start time
	//Call the handler to proceed with the client request
	resp, err := handler(ctx, req)

	//Calculate the duration
	duration := time.Since(start)

	//log the request details with duration
	st, _ := status.FromError(err)
	fmt.Printf("Method: %s ,Status: %d, Dutation: %v\n", info.FullMethod, st.Code(), duration)

	md := metadata.Pairs("X-Response-Time", duration.String())
	grpc.SetHeader(ctx, md)

	log.Println("Sending response from ResponseTimeInterceptor")
	return resp, err
}
