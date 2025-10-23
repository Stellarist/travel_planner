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
		"午餐": "食物", "晚餐": "食物", "宵夜": "食物", "夜宵": "食物",
		"饮料": "食物", "咖啡": "食物", "零食": "食物", "小吃": "食物",
		"美食": "食物", "大餐": "食物", "快餐": "食物", "外卖": "食物",
		"烧烤": "食物", "火锅": "食物", "自助": "食物", "甜品": "食物",
		"蛋糕": "食物", "面包": "食物", "水果": "食物", "菜": "食物",
		"肉": "食物", "海鲜": "食物", "饮食": "食物", "用餐": "食物",
		"酒": "食物", "茶": "食物", "奶茶": "食物", "果汁": "食物",

		// 交通类
		"交通": "交通", "打车": "交通", "出租车": "交通", "的士": "交通",
		"地铁": "交通", "公交": "交通", "火车": "交通", "高铁": "交通",
		"飞机": "交通", "机票": "交通", "车票": "交通", "票": "交通",
		"油费": "交通", "停车": "交通", "停车费": "交通", "过路费": "交通",
		"滴滴": "交通", "uber": "交通", "租车": "交通", "包车": "交通",
		"大巴": "交通", "巴士": "交通", "轮船": "交通", "船票": "交通",
		"动车": "交通", "出行": "交通", "车费": "交通", "路费": "交通",
		"打的": "交通", "坐车": "交通", "开车": "交通", "加油": "交通",
		"高速": "交通", "ETC": "交通",

		// 住宿类
		"住宿": "住宿", "酒店": "住宿", "宾馆": "住宿", "民宿": "住宿",
		"旅馆": "住宿", "住": "住宿", "房费": "住宿", "airbnb": "住宿",
		"旅社": "住宿", "客栈": "住宿", "青旅": "住宿", "招待所": "住宿",
		"房间": "住宿", "订房": "住宿", "住房": "住宿", "住处": "住宿",
		"住店": "住宿", "入住": "住宿", "民居": "住宿",

		// 购物类
		"购物": "购物", "买": "购物", "商场": "购物", "超市": "购物",
		"纪念品": "购物", "礼物": "购物", "衣服": "购物", "鞋": "购物",
		"化妆品": "购物", "shopping": "购物", "商店": "购物", "便利店": "购物",
		"服装": "购物", "鞋子": "购物", "包": "购物", "手表": "购物",
		"首饰": "购物", "饰品": "购物", "电子产品": "购物", "数码": "购物",
		"手机": "购物", "相机": "购物", "特产": "购物", "土特产": "购物",
		"药": "购物", "药品": "购物", "日用品": "购物", "生活用品": "购物",
		"买东西": "购物", "采购": "购物", "扫货": "购物", "剁手": "购物",
		"淘": "购物", "逛街": "购物", "逛": "购物",

		// 活动类
		"活动": "活动", "娱乐": "活动", "门票": "活动", "景点": "活动",
		"游乐园": "活动", "电影": "活动", "演出": "活动", "表演": "活动",
		"玩": "活动", "游玩": "活动", "参观": "活动", "观光": "活动",
		"旅游": "活动", "游览": "活动", "展览": "活动", "博物馆": "活动",
		"动物园": "活动", "海洋馆": "活动", "主题公园": "活动", "乐园": "活动",
		"演唱会": "活动", "音乐会": "活动", "话剧": "活动", "戏剧": "活动",
		"KTV": "活动", "唱歌": "活动", "酒吧": "活动", "夜店": "活动",
		"温泉": "活动", "SPA": "活动", "按摩": "活动", "足浴": "活动",
		"健身": "活动", "游泳": "活动", "运动": "活动", "球": "活动",
		"攀岩": "活动", "滑雪": "活动", "潜水": "活动", "冲浪": "活动",
		"漂流": "活动", "蹦极": "活动", "跳伞": "活动", "游戏": "活动",
		"玩乐": "活动", "娱乐活动": "活动", "休闲": "活动",
	}

	// 优先匹配完整词
	for _, word := range words {
		if category, ok := categoryMap[word]; ok {
			return category
		}
	}

	// 再检查包含关系（从长到短匹配，避免误匹配）
	keywords := make([]string, 0, len(categoryMap))
	for keyword := range categoryMap {
		keywords = append(keywords, keyword)
	}

	// 按长度排序，优先匹配长关键词
	for i := 0; i < len(keywords); i++ {
		for j := i + 1; j < len(keywords); j++ {
			if len(keywords[i]) < len(keywords[j]) {
				keywords[i], keywords[j] = keywords[j], keywords[i]
			}
		}
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return categoryMap[keyword]
		}
	}

	return ""
}

// extractDateRangeForExpense 提取日期范围（用于开销查询）
func extractDateRangeForExpense(text string) (string, string) {
	now := time.Now()
	today := now.Format("2006-01-02")

	// 检查"N周"、"N天"、"N个月"等模式
	// 匹配：最近两周、两周内、过去三周、最近15天、近一个月等
	timeRangePatterns := []struct {
		pattern string
		handler func([]string) (string, string)
	}{
		{
			// 匹配：最近N周、N周内、过去N周、近N周、这N周
			pattern: `(?:最近|过去|近|这)?([一二三四五六七八九十两1-9][十]?|1[0-9]|2[0-9]|30)(?:个)?周(?:内|以内|之内)?`,
			handler: func(matches []string) (string, string) {
				weeks := parseChineseNumber(matches[1])
				days := weeks * 7
				return now.AddDate(0, 0, -days).Format("2006-01-02"), today
			},
		},
		{
			// 匹配：最近N天、N天内、过去N天、近N天、这N天
			pattern: `(?:最近|过去|近|这)?([一二三四五六七八九十两1-9][十]?|1[0-9]|2[0-9]|30|[1-9])(?:个)?天(?:内|以内|之内)?`,
			handler: func(matches []string) (string, string) {
				days := parseChineseNumber(matches[1])
				return now.AddDate(0, 0, -days).Format("2006-01-02"), today
			},
		},
		{
			// 匹配：最近N个月、N个月内、过去N个月、近N个月、这N个月
			pattern: `(?:最近|过去|近|这)?([一二三四五六七八九十两1-9]|1[0-2])(?:个)?月(?:内|以内|之内)?`,
			handler: func(matches []string) (string, string) {
				months := parseChineseNumber(matches[1])
				days := months * 30 // 简化处理，按30天计算
				return now.AddDate(0, 0, -days).Format("2006-01-02"), today
			},
		},
		{
			// 匹配：N月N日到N月N日、N-N到N-N
			pattern: `(\d{1,2})[-月](\d{1,2})日?(?:到|至|~|-)+(\d{1,2})[-月](\d{1,2})日?`,
			handler: func(matches []string) (string, string) {
				startMonth, _ := strconv.Atoi(matches[1])
				startDay, _ := strconv.Atoi(matches[2])
				endMonth, _ := strconv.Atoi(matches[3])
				endDay, _ := strconv.Atoi(matches[4])

				startDate := time.Date(now.Year(), time.Month(startMonth), startDay, 0, 0, 0, 0, now.Location())
				endDate := time.Date(now.Year(), time.Month(endMonth), endDay, 0, 0, 0, 0, now.Location())

				// 如果开始日期晚于结束日期，可能跨年了
				if startDate.After(endDate) {
					startDate = startDate.AddDate(-1, 0, 0)
				}

				return startDate.Format("2006-01-02"), endDate.Format("2006-01-02")
			},
		},
	}

	for _, tp := range timeRangePatterns {
		re := regexp.MustCompile(tp.pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			return tp.handler(matches)
		}
	}

	// 相对日期映射（固定短语）
	relativeDates := map[string]func() (string, string){
		"今天": func() (string, string) { return today, today },
		"今日": func() (string, string) { return today, today },
		"当天": func() (string, string) { return today, today },
		"今儿": func() (string, string) { return today, today },

		"昨天": func() (string, string) {
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
			return yesterday, yesterday
		},
		"昨日": func() (string, string) {
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
			return yesterday, yesterday
		},
		"昨儿": func() (string, string) {
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
			return yesterday, yesterday
		},

		"前天": func() (string, string) {
			dayBefore := now.AddDate(0, 0, -2).Format("2006-01-02")
			return dayBefore, dayBefore
		},
		"前日": func() (string, string) {
			dayBefore := now.AddDate(0, 0, -2).Format("2006-01-02")
			return dayBefore, dayBefore
		},

		"大前天": func() (string, string) {
			threeDaysAgo := now.AddDate(0, 0, -3).Format("2006-01-02")
			return threeDaysAgo, threeDaysAgo
		},

		"本周": func() (string, string) {
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := now.AddDate(0, 0, -weekday+1).Format("2006-01-02")
			return startOfWeek, today
		},
		"这周": func() (string, string) {
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := now.AddDate(0, 0, -weekday+1).Format("2006-01-02")
			return startOfWeek, today
		},
		"这个星期": func() (string, string) {
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := now.AddDate(0, 0, -weekday+1).Format("2006-01-02")
			return startOfWeek, today
		},

		"上周": func() (string, string) {
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfLastWeek := now.AddDate(0, 0, -weekday-6).Format("2006-01-02")
			endOfLastWeek := now.AddDate(0, 0, -weekday).Format("2006-01-02")
			return startOfLastWeek, endOfLastWeek
		},
		"上个星期": func() (string, string) {
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfLastWeek := now.AddDate(0, 0, -weekday-6).Format("2006-01-02")
			endOfLastWeek := now.AddDate(0, 0, -weekday).Format("2006-01-02")
			return startOfLastWeek, endOfLastWeek
		},

		"本月": func() (string, string) {
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfMonth, today
		},
		"这个月": func() (string, string) {
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfMonth, today
		},
		"当月": func() (string, string) {
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfMonth, today
		},

		"上月": func() (string, string) {
			lastMonth := now.AddDate(0, -1, 0)
			startOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location())
			endOfLastMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
			return startOfLastMonth.Format("2006-01-02"), endOfLastMonth.Format("2006-01-02")
		},
		"上个月": func() (string, string) {
			lastMonth := now.AddDate(0, -1, 0)
			startOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location())
			endOfLastMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
			return startOfLastMonth.Format("2006-01-02"), endOfLastMonth.Format("2006-01-02")
		},

		"本年": func() (string, string) {
			startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfYear, today
		},
		"今年": func() (string, string) {
			startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfYear, today
		},
		"这一年": func() (string, string) {
			startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
			return startOfYear, today
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

// parseChineseNumber 解析中文数字和阿拉伯数字
func parseChineseNumber(s string) int {
	// 中文数字映射
	chineseMap := map[string]int{
		"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
		"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
		"两": 2,
	}

	// 先尝试直接转换阿拉伯数字
	if num, err := strconv.Atoi(s); err == nil {
		return num
	}

	// 处理中文数字
	if len(s) == 1 {
		if num, ok := chineseMap[s]; ok {
			return num
		}
	}

	// 处理"十X"或"X十"的情况
	if strings.HasPrefix(s, "十") && len(s) == 2 {
		// 十一、十二...十九
		unit := string([]rune(s)[1])
		if num, ok := chineseMap[unit]; ok {
			return 10 + num
		}
	} else if strings.HasSuffix(s, "十") && len(s) == 2 {
		// 二十、三十...九十
		tens := string([]rune(s)[0])
		if num, ok := chineseMap[tens]; ok {
			return num * 10
		}
	} else if len(s) == 3 && strings.Contains(s, "十") {
		// 二十一、三十五等
		parts := strings.Split(s, "十")
		if len(parts) == 2 {
			tens := chineseMap[parts[0]]
			unit := 0
			if parts[1] != "" {
				unit = chineseMap[parts[1]]
			}
			return tens*10 + unit
		}
	}

	// 默认返回1
	return 1
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
