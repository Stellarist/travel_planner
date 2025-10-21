package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TripPlanRequest 用户行程规划请求
type TripPlanRequest struct {
	Destination  string   `json:"destination"`  // 目的地
	StartDate    string   `json:"startDate"`    // 开始日期 YYYY-MM-DD
	EndDate      string   `json:"endDate"`      // 结束日期 YYYY-MM-DD
	Budget       float64  `json:"budget"`       // 预算（元）
	Travelers    int      `json:"travelers"`    // 同行人数
	Preferences  []string `json:"preferences"`  // 旅行偏好（美食、动漫、亲子等）
	SpecialNeeds string   `json:"specialNeeds"` // 特殊需求
}

// DayItinerary 单日行程
type DayItinerary struct {
	Day           int        `json:"day"`           // 第几天
	Date          string     `json:"date"`          // 日期
	Activities    []Activity `json:"activities"`    // 活动列表
	Accommodation string     `json:"accommodation"` // 住宿
	DailyCost     float64    `json:"dailyCost"`     // 当日花费
}

// Activity 单个活动
type Activity struct {
	Time        string  `json:"time"`        // 时间
	Type        string  `json:"type"`        // 类型（景点、餐厅、交通等）
	Name        string  `json:"name"`        // 名称
	Location    string  `json:"location"`    // 地点
	Duration    string  `json:"duration"`    // 持续时间
	Cost        float64 `json:"cost"`        // 花费
	Description string  `json:"description"` // 描述
	Tips        string  `json:"tips"`        // 小贴士
}

// TripPlan 完整行程计划
type TripPlan struct {
	ID        string          `json:"id"`        // 行程ID
	UserID    int             `json:"userId"`    // 用户ID
	Username  string          `json:"username"`  // 用户名
	Request   TripPlanRequest `json:"request"`   // 原始请求
	Itinerary []DayItinerary  `json:"itinerary"` // 详细行程
	TotalCost float64         `json:"totalCost"` // 总花费
	Summary   string          `json:"summary"`   // 行程摘要
	CreatedAt time.Time       `json:"createdAt"` // 创建时间
	UpdatedAt time.Time       `json:"updatedAt"` // 更新时间
}

// 键设计
func tripKey(tripID string) string {
	return "trip:" + tripID
}

func userTripsKey(username string) string {
	return "user_trips:" + username
}

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

	// 保存行程
	if err := rdb.Set(ctx, tripKey(plan.ID), data, 0).Err(); err != nil {
		return err
	}

	// 添加到用户的行程列表
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

	// 获取用户的行程ID列表
	tripIDs, err := rdb.SMembers(ctx, userTripsKey(username)).Result()
	if err != nil {
		return nil, err
	}

	trips := make([]*TripPlan, 0, len(tripIDs))
	for _, id := range tripIDs {
		trip, err := GetTripPlan(ctx, id)
		if err != nil {
			continue // 跳过错误的行程
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

	// 从用户列表移除
	if err := rdb.SRem(ctx, userTripsKey(username), tripID).Err(); err != nil {
		return err
	}

	// 删除行程数据
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
