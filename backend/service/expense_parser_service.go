package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"example.com/travel_planner/backend/api"
)

// ParsedExpenseQuery 解析后的开销查询信息
type ParsedExpenseQuery struct {
	Category     string `json:"category"`     // 分类：食物、交通、住宿、购物、活动
	StartDate    string `json:"startDate"`    // 开始日期
	EndDate      string `json:"endDate"`      // 结束日期
	Query        string `json:"query"`        // 分析查询文本
	OriginalText string `json:"originalText"` // 原始文本
	Confidence   string `json:"confidence"`   // 置信度：high/medium/low
}

// ParseExpenseQuery 解析开销相关的语音文字
func ParseExpenseQuery(text string) *ParsedExpenseQuery {
	if text == "" {
		return &ParsedExpenseQuery{
			OriginalText: text,
			Confidence:   "low",
		}
	}

	// 使用共享的gse分词器
	seg := api.GetSegmenter()
	words := seg.Cut(text, true)

	result := &ParsedExpenseQuery{
		OriginalText: text,
		Query:        text, // 默认使用原始文本作为查询
	}

	// 提取分类
	result.Category = extractExpenseCategory(text, words)

	// 提取日期范围
	startDate, endDate := extractDateRangeForExpense(text)
	result.StartDate = startDate
	result.EndDate = endDate

	// 评估置信度
	result.Confidence = evaluateExpenseConfidence(result)

	return result
}

// extractExpenseCategory 提取开销分类
func extractExpenseCategory(text string, words []string) string {
	// 分类关键词映射
	categoryMap := map[string]string{
		// 食物类
		"食物": "食物", "吃": "食物", "饭": "食物", "餐": "食物",
		"早饭": "食物", "午饭": "食物", "晚饭": "食物", "早餐": "食物",
		"午餐": "食物", "晚餐": "食物", "宵夜": "食物", "饮料": "食物",
		"咖啡": "食物", "零食": "食物", "小吃": "食物",

		// 交通类
		"交通": "交通", "打车": "交通", "出租车": "交通", "地铁": "交通",
		"公交": "交通", "火车": "交通", "高铁": "交通", "飞机": "交通",
		"机票": "交通", "车票": "交通", "油费": "交通", "停车": "交通",
		"滴滴": "交通", "uber": "交通",

		// 住宿类
		"住宿": "住宿", "酒店": "住宿", "宾馆": "住宿", "民宿": "住宿",
		"旅馆": "住宿", "住": "住宿", "房费": "住宿", "airbnb": "住宿",

		// 购物类
		"购物": "购物", "买": "购物", "商场": "购物", "超市": "购物",
		"纪念品": "购物", "礼物": "购物", "衣服": "购物", "鞋": "购物",
		"化妆品": "购物", "shopping": "购物",

		// 活动类
		"活动": "活动", "娱乐": "活动", "门票": "活动", "景点": "活动",
		"游乐园": "活动", "电影": "活动", "演出": "活动", "表演": "活动",
		"玩": "活动", "游玩": "活动", "参观": "活动",
	}

	// 优先匹配完整词
	for _, word := range words {
		if category, ok := categoryMap[word]; ok {
			return category
		}
	}

	// 再检查包含关系
	for keyword, category := range categoryMap {
		if strings.Contains(text, keyword) {
			return category
		}
	}

	return ""
}

// extractDateRangeForExpense 提取日期范围（用于开销查询）
func extractDateRangeForExpense(text string) (string, string) {
	now := time.Now()
	today := now.Format("2006-01-02")

	// 相对日期映射
	relativeDates := map[string]func() (string, string){
		"今天": func() (string, string) { return today, today },
		"今日": func() (string, string) { return today, today },
		"昨天": func() (string, string) {
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
			return yesterday, yesterday
		},
		"前天": func() (string, string) {
			dayBefore := now.AddDate(0, 0, -2).Format("2006-01-02")
			return dayBefore, dayBefore
		},
		"最近三天": func() (string, string) {
			return now.AddDate(0, 0, -3).Format("2006-01-02"), today
		},
		"三天": func() (string, string) {
			return now.AddDate(0, 0, -3).Format("2006-01-02"), today
		},
		"最近一周": func() (string, string) {
			return now.AddDate(0, 0, -7).Format("2006-01-02"), today
		},
		"一周": func() (string, string) {
			return now.AddDate(0, 0, -7).Format("2006-01-02"), today
		},
		"7天": func() (string, string) {
			return now.AddDate(0, 0, -7).Format("2006-01-02"), today
		},
		"七天": func() (string, string) {
			return now.AddDate(0, 0, -7).Format("2006-01-02"), today
		},
		"最近一个月": func() (string, string) {
			return now.AddDate(0, 0, -30).Format("2006-01-02"), today
		},
		"一个月": func() (string, string) {
			return now.AddDate(0, 0, -30).Format("2006-01-02"), today
		},
		"30天": func() (string, string) {
			return now.AddDate(0, 0, -30).Format("2006-01-02"), today
		},
		"三十天": func() (string, string) {
			return now.AddDate(0, 0, -30).Format("2006-01-02"), today
		},
		"本周": func() (string, string) {
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := now.AddDate(0, 0, -weekday+1).Format("2006-01-02")
			return startOfWeek, today
		},
		"本月": func() (string, string) {
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfMonth, today
		},
	}

	// 检查相对日期
	for keyword, dateFunc := range relativeDates {
		if strings.Contains(text, keyword) {
			return dateFunc()
		}
	}

	// 检查具体日期模式
	patterns := []string{
		`(\d{4})[-年](\d{1,2})[-月](\d{1,2})日?`, // 2025-01-15 或 2025年1月15日
		`(\d{1,2})[-月](\d{1,2})日?`,            // 1-15 或 1月15日
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 0 {
			var dateStr string
			if len(matches) == 4 {
				// 完整日期
				year, _ := strconv.Atoi(matches[1])
				month, _ := strconv.Atoi(matches[2])
				day, _ := strconv.Atoi(matches[3])
				dateStr = time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			} else if len(matches) == 3 {
				// 只有月日，补充当前年份
				month, _ := strconv.Atoi(matches[1])
				day, _ := strconv.Atoi(matches[2])
				dateStr = time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			}
			return dateStr, dateStr
		}
	}

	return "", ""
}

// evaluateExpenseConfidence 评估开销查询的置信度
func evaluateExpenseConfidence(info *ParsedExpenseQuery) string {
	score := 0

	// 有分类 +2
	if info.Category != "" {
		score += 2
	}

	// 有日期范围 +2
	if info.StartDate != "" {
		score += 2
	}

	// 有查询文本 +1
	if info.Query != "" {
		score += 1
	}

	if score >= 4 {
		return "high"
	} else if score >= 2 {
		return "medium"
	}
	return "low"
}
