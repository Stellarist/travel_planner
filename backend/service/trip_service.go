package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GenerateTripPlan builds prompt and delegates to AI client
func GenerateTripPlan(ctx context.Context, req *TripPlanRequest) (*TripPlan, error) {
	prompt := buildPrompt(req)

	// 调用大模型
	content, err := CallModel(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("call model: %w", err)
	}

	// 尝试解析返回的 JSON
	var plan TripPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		// 记录原始内容以便调试
		fmt.Printf("Failed to parse model response. Raw content:\n%s\n", content)
		return nil, fmt.Errorf("parse plan: %w (raw: %s)", err, content[:min(len(content), 200)])
	}

	// 补充必要字段
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

	return &plan, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildPrompt(req *TripPlanRequest) string {
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)
	days := int(endDate.Sub(startDate).Hours()/24) + 1

	prefsStr := "无特定偏好"
	if len(req.Preferences) > 0 {
		prefsStr = strings.Join(req.Preferences, "、")
	}

	specialStr := "无"
	if req.SpecialNeeds != "" {
		specialStr = req.SpecialNeeds
	}

	return fmt.Sprintf(`你是专业的旅行规划助手。根据以下需求生成详细行程计划。

需求：
目的地：%s
日期：%s 至 %s（共%d天）
预算：%.0f元（%d人）
偏好：%s
特殊需求：%s

要求：
1. 严格按JSON格式输出，不要其他文字
2. 每天安排3-5个活动
3. 活动时间合理，预留交通时间
4. 费用分配合理，不超预算
5. 提供实用建议

JSON格式：
{
  "itinerary": [
    {
      "day": 1,
      "date": "%s",
      "activities": [
        {
          "time": "09:00",
          "type": "景点",
          "name": "活动名称",
          "location": "详细地址",
          "duration": "2小时",
          "cost": 100.0,
          "description": "简短描述",
          "tips": "实用提示"
        }
      ],
      "accommodation": "酒店名称和地址",
      "dailyCost": 1000.0
    }
  ],
  "totalCost": %.0f,
  "summary": "行程亮点总结，1-2句话"
}

请立即生成JSON格式的行程计划：`,
		req.Destination,
		req.StartDate, req.EndDate, days,
		req.Budget, req.Travelers,
		prefsStr, specialStr,
		req.StartDate,
		req.Budget)
}
