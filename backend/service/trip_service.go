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
	content, err := CallModel(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("call model: %w", err)
	}

	var plan TripPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
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

	return &plan, nil
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
}
`,
		req.Destination, req.StartDate, req.EndDate, days,
		req.Budget, req.Travelers,
		strings.Join(req.Preferences, "、"), req.SpecialNeeds)
}
