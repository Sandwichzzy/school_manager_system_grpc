package interceptors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	r1 := &rateLimiter{
		visitors:  make(map[string]int),
		limit:     limit,
		resetTime: resetTime,
	}
	//start the reset routine
	go r1.resetVisitorCount()
	return r1
}

func (r1 *rateLimiter) resetVisitorCount() {
	for {
		time.Sleep(r1.resetTime)
		r1.mu.Lock()
		r1.visitors = make(map[string]int)
		r1.mu.Unlock()
	}
}

func (r1 *rateLimiter) RateLimitInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("Rate Limiter Middleware being returned...")
	r1.mu.Lock()
	defer r1.mu.Unlock()

	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to get client IP")
	}

	visitorIP := p.Addr.String()
	r1.visitors[visitorIP]++

	log.Printf("+++++++++++ Visitor count from IP:%s:%d\n", visitorIP, r1.visitors[visitorIP])

	if r1.visitors[visitorIP] > r1.limit {
		return nil, status.Error(codes.ResourceExhausted, "Too many requests")
	}

	fmt.Println("Rate Limiter ends...")
	return handler(ctx, req)
}
