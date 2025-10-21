package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var rdb *redis.Client

// Init 初始化 Redis 客户端
func Init(addr, password string, db int) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}

// UserRecord 存储在 Redis 的用户结构
type UserRecord struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
}

func userKey(username string) string { return "user:" + username }

// GetUser 获取用户
func GetUser(ctx context.Context, username string) (*UserRecord, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}
	val, err := rdb.Get(ctx, userKey(username)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var u UserRecord
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser 创建新用户
func CreateUser(ctx context.Context, username, password string) (*UserRecord, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}
	// 已存在检查
	exists, err := rdb.Exists(ctx, userKey(username)).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, errors.New("user exists")
	}
	// 生成递增ID
	id, err := rdb.Incr(ctx, "user:next_id").Result()
	if err != nil {
		return nil, err
	}
	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &UserRecord{
		ID:           int(id),
		Username:     username,
		PasswordHash: string(hash),
	}
	b, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	if err := rdb.Set(ctx, userKey(username), b, 0).Err(); err != nil {
		return nil, err
	}
	return u, nil
}

// VerifyPassword 校验密码
func VerifyPassword(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// CreateSession 生成会话 token 并保存，默认 24 小时有效
func CreateSession(ctx context.Context, username string, ttl time.Duration) (string, error) {
	if rdb == nil {
		return "", errors.New("redis not initialized")
	}
	// 生成随机 token
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if err := rdb.Set(ctx, "session:"+token, username, ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// GetUsernameByToken 通过 token 获取用户名
func GetUsernameByToken(ctx context.Context, token string) (string, error) {
	if rdb == nil {
		return "", errors.New("redis not initialized")
	}
	val, err := rdb.Get(ctx, "session:"+token).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
