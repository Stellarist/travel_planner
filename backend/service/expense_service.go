package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"example.com/travel_planner/backend/config"
)

// ExpenseRecord mirrors handler struct
type ExpenseRecord struct {
	ID        string  `json:"id"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Note      string  `json:"note"`
	CreatedAt string  `json:"createdAt"`
}

func expenseListKey(username string) string { return "user_expenses:" + username }

// SaveExpense 将单条记录追加到用户的 expense 列表
func SaveExpense(ctx context.Context, username string, rec *ExpenseRecord) error {
	if rdb == nil {
		return errors.New("redis not initialized")
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if err := rdb.RPush(ctx, expenseListKey(username), b).Err(); err != nil {
		return err
	}
	return nil
}

// GetExpenses 返回用户的所有 expense 记录
func GetExpenses(ctx context.Context, username string) ([]*ExpenseRecord, error) {
	if rdb == nil {
		return nil, errors.New("redis not initialized")
	}
	vals, err := rdb.LRange(ctx, expenseListKey(username), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	res := make([]*ExpenseRecord, 0, len(vals))
	for _, v := range vals {
		var er ExpenseRecord
		if err := json.Unmarshal([]byte(v), &er); err != nil {
			continue
		}
		res = append(res, &er)
	}
	return res, nil
}

// GenerateExpenseID 生成唯一 expense ID
func GenerateExpenseID(ctx context.Context) (string, error) {
	if rdb == nil {
		return "", errors.New("redis not initialized")
	}
	id, err := rdb.Incr(ctx, "expense:next_id").Result()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("exp_%d_%d", id, time.Now().Unix()), nil
}

// AnalyzeExpenses 使用配置好的模型对花费进行简单分析
func AnalyzeExpenses(ctx context.Context, username string, list []*ExpenseRecord) (string, error) {
	cfg := config.Global.Model
	if cfg.ApiKey == "" || cfg.BaseURL == "" || cfg.Model == "" {
		return "", errors.New("model not configured")
	}

	// build a simple prompt summarizing expenses
	summary := "User expenses summary:\n"
	total := 0.0
	for _, e := range list {
		summary += fmt.Sprintf("- %s: %.2f %s (%s)\n", e.Category, e.Amount, e.Currency, e.Note)
		total += e.Amount
	}
	summary += fmt.Sprintf("Total: %.2f\n", total)

	payload := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Please analyze the following travel expenses and provide budget suggestions and categorization.\n\n" + summary},
		},
	}

	b, _ := json.Marshal(payload)
	reqURL := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	// small inline HTTP call similar to ai_service.go
	client := &http.Client{Timeout: 20 * time.Second}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+cfg.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("model API error: %s", string(body))
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", errors.New("no result from model")
	}
	return result.Choices[0].Message.Content, nil
}
