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
	Date      string  `json:"date"` // YYYY-MM-DD user-provided date for the expense
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
	// ensure date is set (YYYY-MM-DD)
	if rec.Date == "" {
		rec.Date = time.Now().Format("2006-01-02")
	}
	rec.CreatedAt = time.Now().Format(time.RFC3339)

	if err := service.SaveExpense(ctx, username, &rec); err != nil {
		service.LogError("Failed to save expense for user %s: %v", username, err)
		api.RespondError(c, http.StatusInternalServerError, "保存失败")
		return
	}

	service.LogInfo("User %s created expense %s (category: %s, amount: %.2f %s)", username, rec.ID, rec.Category, rec.Amount, rec.Currency)
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

	// optional filters
	category := c.Query("category")
	from := c.Query("from") // YYYY-MM-DD
	to := c.Query("to")     // YYYY-MM-DD

	list, err := service.GetExpenses(ctx, username)
	if err != nil {
		service.LogError("Failed to get expenses for user %s: %v", username, err)
		api.RespondError(c, http.StatusInternalServerError, "获取失败")
		return
	}

	// apply simple filtering by category and date range
	filtered := make([]*service.ExpenseRecord, 0, len(list))
	var fromT, toT time.Time
	var errf error
	if from != "" {
		fromT, errf = time.Parse("2006-01-02", from)
		if errf != nil {
			api.RespondError(c, http.StatusBadRequest, "invalid 'from' date")
			return
		}
	}
	if to != "" {
		toT, errf = time.Parse("2006-01-02", to)
		if errf != nil {
			api.RespondError(c, http.StatusBadRequest, "invalid 'to' date")
			return
		}
	}

	for _, e := range list {
		if category != "" && e.Category != category {
			continue
		}
		if from != "" || to != "" {
			d, err := time.Parse("2006-01-02", e.Date)
			if err != nil {
				// skip malformed dates
				continue
			}
			if !fromT.IsZero() && d.Before(fromT) {
				continue
			}
			if !toT.IsZero() && d.After(toT) {
				continue
			}
		}
		filtered = append(filtered, e)
	}

	service.LogInfo("User %s retrieved %d expenses (filtered from %d total)", username, len(filtered), len(list))
	api.RespondSuccess(c, filtered)
}

// AnalyzeExpensesHandler 使用大模型分析用户开销并返回建议
func AnalyzeExpensesHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 增加超时时间到 90 秒，因为 AI 分析需要较长时间
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	// 从 POST body 中读取参数
	var req struct {
		Category string `json:"category"`
		From     string `json:"from"`
		To       string `json:"to"`
		Query    string `json:"q"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果 body 为空或解析失败，使用默认值
		req = struct {
			Category string `json:"category"`
			From     string `json:"from"`
			To       string `json:"to"`
			Query    string `json:"q"`
		}{}
	}

	category := req.Category
	from := req.From
	to := req.To
	userQuery := req.Query

	list, err := service.GetExpenses(ctx, username)
	if err != nil {
		service.LogError("Failed to get expenses for analysis for user %s: %v", username, err)
		api.RespondError(c, http.StatusInternalServerError, "获取失败")
		return
	}

	// apply same filtering logic as ListExpensesHandler
	filtered := make([]*service.ExpenseRecord, 0, len(list))
	var fromT, toT time.Time
	var errf error
	if from != "" {
		fromT, errf = time.Parse("2006-01-02", from)
		if errf != nil {
			api.RespondError(c, http.StatusBadRequest, "invalid 'from' date")
			return
		}
	}
	if to != "" {
		toT, errf = time.Parse("2006-01-02", to)
		if errf != nil {
			api.RespondError(c, http.StatusBadRequest, "invalid 'to' date")
			return
		}
	}
	for _, e := range list {
		if category != "" && e.Category != category {
			continue
		}
		if from != "" || to != "" {
			d, err := time.Parse("2006-01-02", e.Date)
			if err != nil {
				continue
			}
			if !fromT.IsZero() && d.Before(fromT) {
				continue
			}
			if !toT.IsZero() && d.After(toT) {
				continue
			}
		}
		filtered = append(filtered, e)
	}

	analysis, err := service.AnalyzeExpenses(ctx, username, filtered, userQuery)
	if err != nil {
		service.LogError("Failed to analyze expenses for user %s: %v", username, err)
		api.RespondError(c, http.StatusInternalServerError, "分析失败: "+err.Error())
		return
	}

	service.LogInfo("User %s analyzed %d expenses", username, len(filtered))
	api.RespondSuccess(c, gin.H{"analysis": analysis})
}
