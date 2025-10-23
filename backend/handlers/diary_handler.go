package handlers

import (
	"net/http"

	"example.com/travel_planner/backend/api"
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
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取 user_id 并进行类型转换
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		service.LogWarn("User %s missing user_id in context", username)
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	var userID int64
	switch v := userIDInterface.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		service.LogError("Invalid user_id type for user %s: %T", username, userIDInterface)
		api.RespondError(c, http.StatusInternalServerError, "用户信息错误")
		return
	}

	var req CreateDiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogWarn("Invalid diary creation request from user %s: %v", username, err)
		api.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	diary := service.DiaryEntry{
		UserID:   userID,
		Date:     req.Date,
		Title:    req.Title,
		Content:  req.Content,
		Images:   req.Images,
		Location: req.Location,
		Mood:     req.Mood,
	}

	id, err := service.CreateDiary(&diary)
	if err != nil {
		service.LogError("Failed to create diary for user %s: %v", username, err)
		api.RespondError(c, http.StatusInternalServerError, "创建日记失败")
		return
	}

	diary.ID = id
	service.LogInfo("User %s created diary %d (mood: %s, location: %s)", username, id, req.Mood, req.Location)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    diary,
	})
}

// GetDiariesHandler 获取用户的所有日记
func GetDiariesHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取 user_id 并进行类型转换
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		service.LogWarn("User %s missing user_id in context", username)
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	var userID int64
	switch v := userIDInterface.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		service.LogError("Invalid user_id type for user %s: %T", username, userIDInterface)
		api.RespondError(c, http.StatusInternalServerError, "用户信息错误")
		return
	}

	diaries, err := service.GetUserDiaries(userID)
	if err != nil {
		service.LogError("Failed to get diaries for user %s: %v", username, err)
		api.RespondError(c, http.StatusInternalServerError, "获取日记失败")
		return
	}

	service.LogInfo("User %s retrieved %d diaries", username, len(diaries))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    diaries,
	})
}

// GetDiaryHandler 获取单条日记
func GetDiaryHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取 user_id 并进行类型转换
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		service.LogWarn("User %s missing user_id in context", username)
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	var userID int64
	switch v := userIDInterface.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		service.LogError("Invalid user_id type for user %s: %T", username, userIDInterface)
		api.RespondError(c, http.StatusInternalServerError, "用户信息错误")
		return
	}

	id := c.Param("id")
	diary, err := service.GetDiary(id, userID)
	if err != nil {
		service.LogWarn("Diary %s not found for user %s", id, username)
		api.RespondError(c, http.StatusNotFound, "日记不存在")
		return
	}

	service.LogInfo("User %s retrieved diary %s", username, id)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    diary,
	})
}

// UpdateDiaryHandler 更新日记
func UpdateDiaryHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取 user_id 并进行类型转换
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		service.LogWarn("User %s missing user_id in context", username)
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	var userID int64
	switch v := userIDInterface.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		service.LogError("Invalid user_id type for user %s: %T", username, userIDInterface)
		api.RespondError(c, http.StatusInternalServerError, "用户信息错误")
		return
	}

	id := c.Param("id")
	var req UpdateDiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogWarn("Invalid diary update request from user %s: %v", username, err)
		api.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	diary := service.DiaryEntry{
		Date:     req.Date,
		Title:    req.Title,
		Content:  req.Content,
		Images:   req.Images,
		Location: req.Location,
		Mood:     req.Mood,
	}

	err := service.UpdateDiary(id, userID, &diary)
	if err != nil {
		service.LogError("Failed to update diary %s for user %s: %v", id, username, err)
		api.RespondError(c, http.StatusInternalServerError, "更新日记失败")
		return
	}

	service.LogInfo("User %s updated diary %s", username, id)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Diary updated successfully",
	})
}

// DeleteDiaryHandler 删除日记
func DeleteDiaryHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取 user_id 并进行类型转换
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		service.LogWarn("User %s missing user_id in context", username)
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	var userID int64
	switch v := userIDInterface.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		service.LogError("Invalid user_id type for user %s: %T", username, userIDInterface)
		api.RespondError(c, http.StatusInternalServerError, "用户信息错误")
		return
	}

	id := c.Param("id")
	err := service.DeleteDiary(id, userID)
	if err != nil {
		service.LogError("Failed to delete diary %s for user %s: %v", id, username, err)
		api.RespondError(c, http.StatusInternalServerError, "删除日记失败")
		return
	}

	service.LogInfo("User %s deleted diary %s", username, id)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Diary deleted successfully",
	})
}
