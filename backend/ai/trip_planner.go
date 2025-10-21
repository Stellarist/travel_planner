package ai

import (
	"context"
	"fmt"
	"time"

	"example.com/travel_planner/backend/store"
)

// GenerateTripPlan 生成行程计划（当前为模拟实现）
// TODO: 接入真实的 OpenAI/Azure OpenAI API
func GenerateTripPlan(ctx context.Context, req *store.TripPlanRequest) (*store.TripPlan, error) {
	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %v", err)
	}

	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days <= 0 {
		return nil, fmt.Errorf("invalid date range")
	}

	// 生成行程ID
	tripID, err := store.GenerateTripID(ctx)
	if err != nil {
		return nil, err
	}

	// 构建模拟行程
	itinerary := make([]store.DayItinerary, days)
	totalCost := 0.0

	for i := 0; i < days; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dailyCost := 0.0

		activities := generateDayActivities(req, i+1, &dailyCost)

		itinerary[i] = store.DayItinerary{
			Day:           i + 1,
			Date:          currentDate.Format("2006-01-02"),
			Activities:    activities,
			Accommodation: generateAccommodation(req),
			DailyCost:     dailyCost,
		}

		totalCost += dailyCost
	}

	plan := &store.TripPlan{
		ID:        tripID,
		Request:   *req,
		Itinerary: itinerary,
		TotalCost: totalCost,
		Summary:   generateSummary(req, days, totalCost),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return plan, nil
}

// 生成单日活动（模拟）
func generateDayActivities(req *store.TripPlanRequest, day int, dailyCost *float64) []store.Activity {
	activities := []store.Activity{}

	// 早餐
	breakfast := store.Activity{
		Time:        "08:00",
		Type:        "餐厅",
		Name:        fmt.Sprintf("%s特色早餐店", req.Destination),
		Location:    fmt.Sprintf("%s市中心", req.Destination),
		Duration:    "1小时",
		Cost:        50 * float64(req.Travelers),
		Description: "品尝当地特色早餐",
		Tips:        "建议提前预约，避免排队",
	}
	activities = append(activities, breakfast)
	*dailyCost += breakfast.Cost

	// 上午景点
	morning := store.Activity{
		Time:        "09:30",
		Type:        "景点",
		Name:        getAttraction(req, day, "morning"),
		Location:    fmt.Sprintf("%s著名景区", req.Destination),
		Duration:    "3小时",
		Cost:        200 * float64(req.Travelers),
		Description: "游览热门景点，拍照打卡",
		Tips:        "避开高峰时段，建议网上购票",
	}
	activities = append(activities, morning)
	*dailyCost += morning.Cost

	// 午餐
	lunch := store.Activity{
		Time:        "12:30",
		Type:        "餐厅",
		Name:        fmt.Sprintf("%s地道美食餐厅", req.Destination),
		Location:    "景区附近",
		Duration:    "1.5小时",
		Cost:        150 * float64(req.Travelers),
		Description: "品尝地道美食",
		Tips:        "推荐特色菜品",
	}
	activities = append(activities, lunch)
	*dailyCost += lunch.Cost

	// 下午活动
	afternoon := store.Activity{
		Time:        "14:30",
		Type:        getAfternoonType(req),
		Name:        getAttraction(req, day, "afternoon"),
		Location:    fmt.Sprintf("%s文化区", req.Destination),
		Duration:    "2.5小时",
		Cost:        180 * float64(req.Travelers),
		Description: getAfternoonDescription(req),
		Tips:        "注意营业时间",
	}
	activities = append(activities, afternoon)
	*dailyCost += afternoon.Cost

	// 晚餐
	dinner := store.Activity{
		Time:        "18:00",
		Type:        "餐厅",
		Name:        fmt.Sprintf("%s夜市/美食街", req.Destination),
		Location:    "市中心商业区",
		Duration:    "2小时",
		Cost:        200 * float64(req.Travelers),
		Description: "体验当地夜生活，品尝小吃",
		Tips:        "注意人身财物安全",
	}
	activities = append(activities, dinner)
	*dailyCost += dinner.Cost

	return activities
}

// 根据偏好生成景点名称
func getAttraction(req *store.TripPlanRequest, day int, timeSlot string) string {
	baseAttractions := map[string][]string{
		"morning": {
			fmt.Sprintf("%s古城/历史街区", req.Destination),
			fmt.Sprintf("%s博物馆", req.Destination),
			fmt.Sprintf("%s公园", req.Destination),
		},
		"afternoon": {
			fmt.Sprintf("%s购物中心", req.Destination),
			fmt.Sprintf("%s艺术区", req.Destination),
			fmt.Sprintf("%s主题乐园", req.Destination),
		},
	}

	// 根据偏好调整
	for _, pref := range req.Preferences {
		switch pref {
		case "美食":
			if timeSlot == "afternoon" {
				return fmt.Sprintf("%s美食街探店", req.Destination)
			}
		case "动漫":
			return fmt.Sprintf("%s动漫主题馆", req.Destination)
		case "亲子", "带孩子":
			return fmt.Sprintf("%s儿童乐园", req.Destination)
		case "历史":
			return fmt.Sprintf("%s历史遗迹", req.Destination)
		case "自然":
			return fmt.Sprintf("%s自然风景区", req.Destination)
		}
	}

	attractions := baseAttractions[timeSlot]
	return attractions[(day-1)%len(attractions)]
}

// 根据偏好确定下午活动类型
func getAfternoonType(req *store.TripPlanRequest) string {
	for _, pref := range req.Preferences {
		switch pref {
		case "美食":
			return "美食体验"
		case "动漫":
			return "主题馆"
		case "亲子", "带孩子":
			return "亲子活动"
		case "购物":
			return "购物"
		}
	}
	return "景点"
}

// 生成下午活动描述
func getAfternoonDescription(req *store.TripPlanRequest) string {
	for _, pref := range req.Preferences {
		switch pref {
		case "美食":
			return "深度探索当地美食文化"
		case "动漫":
			return "体验动漫主题互动"
		case "亲子", "带孩子":
			return "适合全家的互动体验"
		case "购物":
			return "逛当地特色商店"
		}
	}
	return "深度文化体验"
}

// 生成住宿推荐
func generateAccommodation(req *store.TripPlanRequest) string {
	if req.Travelers > 4 {
		return fmt.Sprintf("%s市中心家庭公寓（可容纳%d人）", req.Destination, req.Travelers)
	} else if req.Travelers > 2 {
		return fmt.Sprintf("%s商务酒店标准间×2", req.Destination)
	}
	return fmt.Sprintf("%s精品酒店标准间", req.Destination)
}

// 生成行程摘要
func generateSummary(req *store.TripPlanRequest, days int, totalCost float64) string {
	preferences := "经典路线"
	if len(req.Preferences) > 0 {
		preferences = ""
		for i, p := range req.Preferences {
			if i > 0 {
				preferences += "、"
			}
			preferences += p
		}
		preferences += "主题"
	}

	return fmt.Sprintf(
		"为您规划了%s %d天%d晚的%s旅行。行程包含经典景点、地道美食和特色体验，预计总花费约%.0f元（人均%.0f元）。适合%d人出行，已根据您的偏好优化。",
		req.Destination,
		days,
		days-1,
		preferences,
		totalCost,
		totalCost/float64(req.Travelers),
		req.Travelers,
	)
}
