package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// VerifyPassword 用于验证用户输入的密码是否和存储的密码一致
// password: 用户输入的明文密码
// encodedHash: 数据库存储的密码，格式为 "saltBase64.hashBase64"
func VerifyPassword(password, encodedHash string) error {

	// 按 "." 分割存储的 hash，得到 salt 和 hash 两部分
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		// 如果格式不正确，说明数据库存储的密码格式异常
		return ErrorHandler(errors.New("invalide encoded hash format"), "internal server error")
	}

	// Base64 编码的 salt
	saltBase64 := parts[0]
	// Base64 编码的 hash
	hashedPasswordBase64 := parts[1]

	// 将 Base64 的 salt 解码为原始字节
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(err, "internal server error")
	}

	// 将 Base64 的 hash 解码为原始字节
	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return ErrorHandler(err, "internal error")
	}

	// 使用 Argon2id 算法重新计算用户输入密码的 hash
	// 参数说明：
	// password  : 用户输入的密码
	// salt      : 原始 salt
	// 1         : time cost（迭代次数）
	// 64*1024   : memory cost（64MB 内存）
	// 4         : 并行度（线程数）
	// 32        : 输出 hash 长度（bytes）
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// 如果长度不一致，说明 hash 不匹配
	if len(hash) != len(hashedPassword) {
		return ErrorHandler(errors.New("hash length mismatch"), "incorrect password")
	}

	// 使用 ConstantTimeCompare 进行常量时间比较
	// 可以防止 timing attack（时间侧信道攻击）
	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		// 返回 nil 说明密码正确
		return nil
	}

	// 密码错误
	return ErrorHandler(errors.New("incorrect password"), "incorrect password")
}

// HashPassword 用于生成密码 hash（注册或修改密码时使用）
func HashPassword(password string) (string, error) {

	// 不允许空密码
	if password == "" {
		return "", ErrorHandler(errors.New("password is blank"), "please enter password")
	}

	// 生成 16 字节随机 salt
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("failed to generate salt"), "internal error")
	}

	// 使用 Argon2id 算法生成密码 hash
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// 将 salt 编码为 Base64 方便存储
	saltBase64 := base64.StdEncoding.EncodeToString(salt)

	// 将 hash 编码为 Base64
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	// 最终存储格式：
	// saltBase64.hashBase64
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)

	return encodedHash, nil
}
