package interceptors

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

func AuthenticationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("AuthenticationInterceptor started")

	// skip specific rpcs
	log.Println(info.FullMethod)
	skipMethod := map[string]bool{
		"/main.ExecsService/Login":          true,
		"/main.ExecsService/ForgotPassword": true,
		"/main.ExecsService/ResetPassword":  true,
	}
	if skipMethod[info.FullMethod] {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata unavailable")
	}
	authHeader, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authorization token unavailable")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	jwtSecret := os.Getenv("JWT_SECRET")
	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Unauthorized Access")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	if !parsedToken.Valid {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	userId := claims["uid"].(string)
	username := claims["user"].(string)
	expiresAtF64 := claims["exp"].(float64)
	expiresAtI64 := int64(expiresAtF64)
	expiresAt := fmt.Sprintf("%v", expiresAtI64)

	newCtx := context.WithValue(ctx, ContextKey("role"), role)
	newCtx = context.WithValue(newCtx, ContextKey("userId"), userId)
	newCtx = context.WithValue(newCtx, ContextKey("username"), username)
	newCtx = context.WithValue(newCtx, ContextKey("expiresAt"), expiresAt)

	log.Println("Auth interceptor ending")
	return handler(newCtx, req)
}
