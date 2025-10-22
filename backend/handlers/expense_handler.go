package handlers

import (
	"context"
	"net/http"
	"time"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

// ExpenseRecord represents a single expense
type ExpenseRecord struct {
	ID        string  `json:"id"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Note      string  `json:"note"`
	CreatedAt string  `json:"createdAt"`
}

// CreateExpenseHandler 保存一条花费记录
func CreateExpenseHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	var rec service.ExpenseRecord
	if err := c.ShouldBindJSON(&rec); err != nil {
		api.RespondError(c, http.StatusBadRequest, "参数错误")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if rec.ID == "" {
		id, _ := service.GenerateExpenseID(ctx)
		rec.ID = id
	}
	rec.CreatedAt = time.Now().Format(time.RFC3339)

	if err := service.SaveExpense(ctx, username, &rec); err != nil {
		api.RespondError(c, http.StatusInternalServerError, "保存失败")
		return
	}

	api.RespondSuccess(c, rec)
}

// ListExpensesHandler 列出用户花费
func ListExpensesHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := service.GetExpenses(ctx, username)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "获取失败")
		return
	}

	api.RespondSuccess(c, list)
}

// AnalyzeExpensesHandler 使用大模型分析用户开销并返回建议
func AnalyzeExpensesHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	list, err := service.GetExpenses(ctx, username)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "获取失败")
		return
	}

	analysis, err := service.AnalyzeExpenses(ctx, username, list)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "分析失败: "+err.Error())
		return
	}

	api.RespondSuccess(c, gin.H{"analysis": analysis})
}
