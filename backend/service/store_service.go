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

// MarshalJSON 自定义 JSON 序列化，将 Request 中的字段提升到顶层
func (t *TripPlan) MarshalJSON() ([]byte, error) {
	type Alias TripPlan
	return json.Marshal(&struct {
		Destination string `json:"destination"`
		StartDate   string `json:"startDate"`
		EndDate     string `json:"endDate"`
		*Alias
	}{
		Destination: t.Request.Destination,
		StartDate:   t.Request.StartDate,
		EndDate:     t.Request.EndDate,
		Alias:       (*Alias)(t),
	})
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

// Favorite 收藏的景点
type Favorite struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Lng     float64 `json:"lng"`
	Lat     float64 `json:"lat"`
	Address string  `json:"address"`
}

func userFavoritesKey(username string) string { return "user_favorites:" + username }

// GetUserFavorites 获取用户的收藏夹
func GetUserFavorites(ctx context.Context, username string) ([]Favorite, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}
	data, err := rdb.Get(ctx, userFavoritesKey(username)).Result()
	if err == redis.Nil {
		return []Favorite{}, nil
	}
	if err != nil {
		return nil, err
	}
	var favorites []Favorite
	if err := json.Unmarshal([]byte(data), &favorites); err != nil {
		return nil, err
	}
	return favorites, nil
}

// SaveUserFavorites 保存用户的收藏夹
func SaveUserFavorites(ctx context.Context, username string, favorites []Favorite) error {
	if rdb == nil {
		return errors.New("redis not initialized")
	}
	data, err := json.Marshal(favorites)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, userFavoritesKey(username), data, 0).Err()
}

// AddFavorite 添加收藏
func AddFavorite(ctx context.Context, username string, favorite Favorite) error {
	favorites, err := GetUserFavorites(ctx, username)
	if err != nil {
		return err
	}
	// 检查是否已存在
	for _, f := range favorites {
		if f.ID == favorite.ID {
			return errors.New("favorite already exists")
		}
	}
	favorites = append(favorites, favorite)
	return SaveUserFavorites(ctx, username, favorites)
}

// RemoveFavorite 删除收藏
func RemoveFavorite(ctx context.Context, username, favoriteID string) error {
	favorites, err := GetUserFavorites(ctx, username)
	if err != nil {
		return err
	}
	newFavorites := make([]Favorite, 0, len(favorites))
	for _, f := range favorites {
		if f.ID != favoriteID {
			newFavorites = append(newFavorites, f)
		}
	}
	return SaveUserFavorites(ctx, username, newFavorites)
}

// ==================== 行程收藏功能 ====================

func userFavoriteTripIDsKey(username string) string { return "user_favorite_trips:" + username }

// GetUserFavoriteTrips 获取用户收藏的行程列表
func GetUserFavoriteTrips(ctx context.Context, username string) ([]*TripPlan, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}

	// 获取收藏的行程 ID 列表
	tripIDs, err := rdb.SMembers(ctx, userFavoriteTripIDsKey(username)).Result()
	if err != nil {
		return nil, err
	}

	// 获取每个行程的详细信息
	trips := make([]*TripPlan, 0, len(tripIDs))
	for _, id := range tripIDs {
		trip, err := GetTripPlan(ctx, id)
		if err != nil {
			continue // 忽略获取失败的行程
		}
		if trip != nil {
			trips = append(trips, trip)
		}
	}

	return trips, nil
}

// AddFavoriteTrip 添加行程到收藏
func AddFavoriteTrip(ctx context.Context, username, tripID string) error {
	if rdb == nil {
		return errors.New("redis not initialized")
	}

	// 检查行程是否存在
	trip, err := GetTripPlan(ctx, tripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return errors.New("trip not found")
	}

	// 添加到收藏集合
	return rdb.SAdd(ctx, userFavoriteTripIDsKey(username), tripID).Err()
}

// RemoveFavoriteTrip 从收藏中移除行程
func RemoveFavoriteTrip(ctx context.Context, username, tripID string) error {
	if rdb == nil {
		return errors.New("redis not initialized")
	}
	return rdb.SRem(ctx, userFavoriteTripIDsKey(username), tripID).Err()
}

// IsTripFavorited 检查行程是否已收藏
func IsTripFavorited(ctx context.Context, username, tripID string) (bool, error) {
	if rdb == nil {
		return false, errors.New("redis not initialized")
	}
	return rdb.SIsMember(ctx, userFavoriteTripIDsKey(username), tripID).Result()
}
