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
)

// GenerateTripPlan 生成行程计划，使用 OpenAI 兼容接口
func GenerateTripPlan(ctx context.Context, req *TripPlanRequest) (*TripPlan, error) {
	cfg := GlobalAppConfig.Model

	if cfg.ApiKey == "" || cfg.BaseURL == "" || cfg.Model == "" {
		return nil, errors.New("模型配置不完整，请检查 config.json")
	}

	plan, err := callOpenAI(ctx, req, cfg)
	if err != nil {
		return nil, fmt.Errorf("调用大模型失败: %v", err)
	}

	if plan.ID == "" {
		plan.ID, _ = GenerateTripID(ctx)
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = time.Now()
	}
	if plan.UpdatedAt.IsZero() {
		plan.UpdatedAt = time.Now()
	}
	plan.Request = *req

	return plan, nil
}

func callOpenAI(ctx context.Context, req *TripPlanRequest, cfg ModelConfig) (*TripPlan, error) {
	prompt := buildPrompt(req)

	payload := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	reqURL := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(requestCtx, "POST", reqURL, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+cfg.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		if requestCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("请求超时，大模型可能负载过高，请稍后重试")
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Error.Message != "" {
		return nil, fmt.Errorf("API错误: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("API未返回任何结果")
	}

	content := extractJSON(result.Choices[0].Message.Content)

	var plan TripPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("解析行程JSON失败: %v", err)
	}

	return &plan, nil
}

func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}

func buildPrompt(req *TripPlanRequest) string {
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)
	days := int(endDate.Sub(startDate).Hours()/24) + 1

	return fmt.Sprintf(`你是一个专业的旅行规划助手。请根据以下用户需求，生成详细的旅行行程计划。

**用户需求：**
- 目的地：%s
- 出发日期：%s
- 返程日期：%s（共 %d 天）
- 预算：%.2f 元
- 同行人数：%d 人
- 旅行偏好：%s
- 特殊需求：%s

**输出要求：**
请严格按照以下 JSON 格式输出行程计划，不要包含任何其他文字：

{
  "itinerary": [
    {
      "day": 1,
      "date": "2006-01-02",
      "activities": [
        {
          "time": "08:00",
          "type": "餐厅/景点/购物等",
          "name": "活动名称",
          "location": "详细地址",
          "duration": "持续时间",
          "cost": 100.00,
          "description": "活动详细描述",
          "tips": "实用小贴士"
        }
      ],
      "accommodation": "住宿推荐",
      "dailyCost": 1000.00
    }
  ],
  "totalCost": 3000.00,
  "summary": "行程总结"
}`,
		req.Destination, req.StartDate, req.EndDate, days,
		req.Budget, req.Travelers,
		strings.Join(req.Preferences, "、"), req.SpecialNeeds)
}
