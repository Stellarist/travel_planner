package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type DiaryEntry struct {
	ID       int64    `json:"id"`
	UserID   int64    `json:"user_id"`
	Date     string   `json:"date"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	Location string   `json:"location,omitempty"`
	Mood     string   `json:"mood,omitempty"`
}

// CreateDiary 创建日记
func CreateDiary(diary *DiaryEntry) (int64, error) {
	ctx := context.Background()
	id := time.Now().UnixNano() / 1e6 // 使用毫秒时间戳作为ID
	diary.ID = id

	key := fmt.Sprintf("diary:%d:%d", diary.UserID, id)
	data, err := json.Marshal(diary)
	if err != nil {
		return 0, err
	}

	// 保存日记
	if err := rdb.Set(ctx, key, data, 0).Err(); err != nil {
		return 0, err
	}

	// 将日记ID添加到用户的日记列表
	userDiariesKey := fmt.Sprintf("user:%d:diaries", diary.UserID)
	if err := rdb.ZAdd(ctx, userDiariesKey, redis.Z{
		Score:  float64(id),
		Member: strconv.FormatInt(id, 10),
	}).Err(); err != nil {
		return 0, err
	}

	return id, nil
}

// GetUserDiaries 获取用户的所有日记
func GetUserDiaries(userID int64) ([]DiaryEntry, error) {
	ctx := context.Background()
	userDiariesKey := fmt.Sprintf("user:%d:diaries", userID)

	// 获取所有日记ID，按时间倒序
	ids, err := rdb.ZRevRange(ctx, userDiariesKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var diaries []DiaryEntry
	for _, idStr := range ids {
		key := fmt.Sprintf("diary:%d:%s", userID, idStr)
		data, err := rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var diary DiaryEntry
		if err := json.Unmarshal([]byte(data), &diary); err != nil {
			continue
		}
		diaries = append(diaries, diary)
	}

	return diaries, nil
}

// GetDiary 获取单条日记
func GetDiary(idStr string, userID int64) (*DiaryEntry, error) {
	ctx := context.Background()
	key := fmt.Sprintf("diary:%d:%s", userID, idStr)

	data, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var diary DiaryEntry
	if err := json.Unmarshal([]byte(data), &diary); err != nil {
		return nil, err
	}

	return &diary, nil
}

// UpdateDiary 更新日记
func UpdateDiary(idStr string, userID int64, updated *DiaryEntry) error {
	ctx := context.Background()
	key := fmt.Sprintf("diary:%d:%s", userID, idStr)

	// 先获取原有日记
	data, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	var diary DiaryEntry
	if err := json.Unmarshal([]byte(data), &diary); err != nil {
		return err
	}

	// 更新字段
	if updated.Date != "" {
		diary.Date = updated.Date
	}
	if updated.Title != "" {
		diary.Title = updated.Title
	}
	if updated.Content != "" {
		diary.Content = updated.Content
	}
	if updated.Images != nil {
		diary.Images = updated.Images
	}
	diary.Location = updated.Location
	diary.Mood = updated.Mood

	// 保存更新后的日记
	newData, err := json.Marshal(diary)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, newData, 0).Err()
}

// DeleteDiary 删除日记
func DeleteDiary(idStr string, userID int64) error {
	ctx := context.Background()
	key := fmt.Sprintf("diary:%d:%s", userID, idStr)

	// 从Redis删除日记数据
	if err := rdb.Del(ctx, key).Err(); err != nil {
		return err
	}

	// 从用户的日记列表中删除
	userDiariesKey := fmt.Sprintf("user:%d:diaries", userID)
	return rdb.ZRem(ctx, userDiariesKey, idStr).Err()
}
