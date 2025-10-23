package handlers

import (
	"net/http"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

// GetFavorites 获取用户收藏的景点列表
func GetFavorites(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	favorites, err := service.GetUserFavorites(c.Request.Context(), username)
	if err != nil {
		service.LogError("Failed to get favorites for user %s: %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	service.LogInfo("User %s retrieved %d favorites", username, len(favorites))
	c.JSON(http.StatusOK, favorites)
}

// AddFavorite 添加景点到收藏夹
func AddFavorite(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var favorite service.Favorite
	if err := c.ShouldBindJSON(&favorite); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := service.AddFavorite(c.Request.Context(), username, favorite); err != nil {
		if err.Error() == "favorite already exists" {
			service.LogWarn("User %s attempted to add duplicate favorite: %s", username, favorite.Name)
			c.JSON(http.StatusConflict, gin.H{"error": "已收藏"})
			return
		}
		service.LogError("Failed to add favorite for user %s: %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	service.LogInfo("User %s added favorite: %s", username, favorite.Name)
	c.JSON(http.StatusOK, gin.H{"message": "收藏成功"})
}

// RemoveFavorite 从收藏夹中删除景点
func RemoveFavorite(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	favoriteID := c.Param("id")
	if favoriteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := service.RemoveFavorite(c.Request.Context(), username, favoriteID); err != nil {
		service.LogError("Failed to remove favorite %s for user %s: %v", favoriteID, username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	service.LogInfo("User %s removed favorite: %s", username, favoriteID)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 未来扩展功能区域 ====================
// TODO: 可以在这里添加更多探索景点相关的功能
// - SearchAttractions: 景点搜索
// - GetAttractionDetail: 获取景点详情
// - GetRecommendations: 获取推荐景点
// - AddAttractionReview: 添加景点评论
// - GetAttractionReviews: 获取景点评论列表
