package handlers

import (
	"net/http"

	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

type CreateDiaryRequest struct {
	Date     string   `json:"date" binding:"required"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	Location string   `json:"location"`
	Mood     string   `json:"mood"`
}

type UpdateDiaryRequest struct {
	Date     string   `json:"date"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	Location string   `json:"location"`
	Mood     string   `json:"mood"`
}

// CreateDiaryHandler 创建日记
func CreateDiaryHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateDiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将 int 转换为 int64
	var uid int64
	switch v := userID.(type) {
	case int:
		uid = int64(v)
	case int64:
		uid = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	diary := &service.DiaryEntry{
		UserID:   uid,
		Date:     req.Date,
		Title:    req.Title,
		Content:  req.Content,
		Images:   req.Images,
		Location: req.Location,
		Mood:     req.Mood,
	}

	id, err := service.CreateDiary(diary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create diary"})
		return
	}

	diary.ID = id
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    diary,
	})
}

// GetDiariesHandler 获取用户的所有日记
func GetDiariesHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 将 int 转换为 int64
	var uid int64
	switch v := userID.(type) {
	case int:
		uid = int64(v)
	case int64:
		uid = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	diaries, err := service.GetUserDiaries(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get diaries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    diaries,
	})
}

// GetDiaryHandler 获取单条日记
func GetDiaryHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 将 int 转换为 int64
	var uid int64
	switch v := userID.(type) {
	case int:
		uid = int64(v)
	case int64:
		uid = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	id := c.Param("id")
	diary, err := service.GetDiary(id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Diary not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    diary,
	})
}

// UpdateDiaryHandler 更新日记
func UpdateDiaryHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 将 int 转换为 int64
	var uid int64
	switch v := userID.(type) {
	case int:
		uid = int64(v)
	case int64:
		uid = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	id := c.Param("id")
	var req UpdateDiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	diary := &service.DiaryEntry{
		Date:     req.Date,
		Title:    req.Title,
		Content:  req.Content,
		Images:   req.Images,
		Location: req.Location,
		Mood:     req.Mood,
	}

	err := service.UpdateDiary(id, uid, diary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update diary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Diary updated successfully",
	})
}

// DeleteDiaryHandler 删除日记
func DeleteDiaryHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 将 int 转换为 int64
	var uid int64
	switch v := userID.(type) {
	case int:
		uid = int64(v)
	case int64:
		uid = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	id := c.Param("id")
	err := service.DeleteDiary(id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete diary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Diary deleted successfully",
	})
}
