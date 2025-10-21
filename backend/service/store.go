package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var rdb *redis.Client

// InitRedis 初始化 Redis 客户端
func InitRedis(addr, password string, db int) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}

// ==================== 用户相关 ====================

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
	exists, err := rdb.Exists(ctx, userKey(username)).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, errors.New("user exists")
	}
	id, err := rdb.Incr(ctx, "user:next_id").Result()
	if err != nil {
		return nil, err
	}
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

// ==================== 行程相关 ====================

// TripPlanRequest 用户行程规划请求
type TripPlanRequest struct {
	Destination  string   `json:"destination"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	Budget       float64  `json:"budget"`
	Travelers    int      `json:"travelers"`
	Preferences  []string `json:"preferences"`
	SpecialNeeds string   `json:"specialNeeds"`
}

// DayItinerary 单日行程
type DayItinerary struct {
	Day           int        `json:"day"`
	Date          string     `json:"date"`
	Activities    []Activity `json:"activities"`
	Accommodation string     `json:"accommodation"`
	DailyCost     float64    `json:"dailyCost"`
}

// Activity 单个活动
type Activity struct {
	Time        string  `json:"time"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Duration    string  `json:"duration"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
	Tips        string  `json:"tips"`
}

// TripPlan 完整行程计划
type TripPlan struct {
	ID        string          `json:"id"`
	UserID    int             `json:"userId"`
	Username  string          `json:"username"`
	Request   TripPlanRequest `json:"request"`
	Itinerary []DayItinerary  `json:"itinerary"`
	TotalCost float64         `json:"totalCost"`
	Summary   string          `json:"summary"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func tripKey(tripID string) string        { return "trip:" + tripID }
func userTripsKey(username string) string { return "user_trips:" + username }

// SaveTripPlan 保存行程计划
func SaveTripPlan(ctx context.Context, plan *TripPlan) error {
	if rdb == nil {
		return errors.New("redis not initialized")
	}
	plan.UpdatedAt = time.Now()
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = plan.UpdatedAt
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	if err := rdb.Set(ctx, tripKey(plan.ID), data, 0).Err(); err != nil {
		return err
	}
	if err := rdb.SAdd(ctx, userTripsKey(plan.Username), plan.ID).Err(); err != nil {
		return err
	}
	return nil
}

// GetTripPlan 获取行程计划
func GetTripPlan(ctx context.Context, tripID string) (*TripPlan, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}
	data, err := rdb.Get(ctx, tripKey(tripID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var plan TripPlan
	if err := json.Unmarshal([]byte(data), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetUserTrips 获取用户的所有行程
func GetUserTrips(ctx context.Context, username string) ([]*TripPlan, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}
	tripIDs, err := rdb.SMembers(ctx, userTripsKey(username)).Result()
	if err != nil {
		return nil, err
	}
	trips := make([]*TripPlan, 0, len(tripIDs))
	for _, id := range tripIDs {
		trip, err := GetTripPlan(ctx, id)
		if err != nil {
			continue
		}
		if trip != nil {
			trips = append(trips, trip)
		}
	}
	return trips, nil
}

// DeleteTripPlan 删除行程计划
func DeleteTripPlan(ctx context.Context, tripID, username string) error {
	if rdb == nil {
		return errors.New("redis not initialized")
	}
	if err := rdb.SRem(ctx, userTripsKey(username), tripID).Err(); err != nil {
		return err
	}
	if err := rdb.Del(ctx, tripKey(tripID)).Err(); err != nil {
		return err
	}
	return nil
}

// GenerateTripID 生成唯一行程ID
func GenerateTripID(ctx context.Context) (string, error) {
	if rdb == nil {
		return "", errors.New("redis not initialized")
	}
	id, err := rdb.Incr(ctx, "trip:next_id").Result()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("trip_%d_%d", id, time.Now().Unix()), nil
}
