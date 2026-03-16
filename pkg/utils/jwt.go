package utils

import (
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SignToken 用于生成 JWT Token
// userId  : 用户ID
// username: 用户名
// role    : 用户角色
func SignToken(userId string, username, role string) (string, error) {

	// 从环境变量中读取 JWT 密钥
	jwtSecret := os.Getenv("JWT_SECRET")

	// 从环境变量中读取 JWT 过期时间，例如： "15m"、"1h"
	jwtExpiresIn := os.Getenv("JWT_EXPIRES_IN")

	// 构建 JWT Claims（Token中存储的数据）
	claims := jwt.MapClaims{
		"uid":  userId,   // 用户ID
		"user": username, // 用户名
		"role": role,     // 用户角色
	}

	// 如果配置了 JWT_EXPIRES_IN
	if jwtExpiresIn != "" {

		// 解析时间字符串为 duration，例如 "15m" -> 15分钟
		duration, err := time.ParseDuration(jwtExpiresIn)
		if err != nil {
			return "", ErrorHandler(err, "Internal error")
		}

		// 设置过期时间 exp（JWT 标准字段）
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))

	} else {

		// 如果没有配置默认 15 分钟过期
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	}

	// 创建 JWT Token，使用 HS256 签名算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用 secret 对 token 进行签名，生成最终字符串
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", ErrorHandler(err, "Internal error")
	}

	// 返回 JWT Token
	return signedToken, nil
}

////////////////////////////////////////////////////////

// 全局 JWTStore 实例
// 用于存储已登出的 Token（黑名单机制）

var JwtStore = JWTStore{
	Tokens: make(map[string]time.Time),
}

// JWTStore 用于保存 Token 黑名单
// JWT 本身是 无状态的：
// 问题：
// 用户登出后 token 仍然有效
// Logout -> 把 token 放进 blacklist
type JWTStore struct {

	// 互斥锁，保证并发安全
	mu sync.Mutex

	// Token -> 过期时间
	Tokens map[string]time.Time
}

// AddToken 将 Token 加入黑名单（通常用于用户登出）
func (store *JWTStore) AddToken(token string, expiryTime time.Time) {

	// 加锁，保证 map 并发安全
	store.mu.Lock()
	defer store.mu.Unlock()

	// 保存 token 和过期时间
	store.Tokens[token] = expiryTime
}

// CleanUpExpiredTokens 定期清理过期 Token
func (store *JWTStore) CleanUpExpiredTokens() {

	for {

		// 每2分钟执行一次清理
		time.Sleep(2 * time.Minute)

		store.mu.Lock()

		// 遍历所有 token
		for token, timeStamp := range store.Tokens {

			// 如果当前时间已经超过 token 过期时间
			if time.Now().After(timeStamp) {

				// 从黑名单删除
				delete(store.Tokens, token)
			}
		}

		store.mu.Unlock()
	}
}

// IsLoggedOut 判断 Token 是否已经被加入黑名单
// 每次 API 请求：
// 验证 JWT 签名
// 检查 exp 是否过期
// 调用 IsLoggedOut(token)
func (store *JWTStore) IsLoggedOut(token string) bool {

	// 加锁保证并发安全
	store.mu.Lock()
	defer store.mu.Unlock()

	// 检查 token 是否存在
	_, ok := store.Tokens[token]

	// 存在表示已经登出
	return ok
}
