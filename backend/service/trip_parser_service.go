package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"example.com/travel_planner/backend/api"
)

// ParsedTripInfo 解析后的旅行信息
type ParsedTripInfo struct {
	Destination  string   `json:"destination"`  // 目的地
	StartDate    string   `json:"startDate"`    // 开始日期 (YYYY-MM-DD)
	EndDate      string   `json:"endDate"`      // 结束日期 (YYYY-MM-DD)
	Duration     int      `json:"duration"`     // 持续天数
	Budget       float64  `json:"budget"`       // 预算（元）
	Travelers    int      `json:"travelers"`    // 旅行人数
	Preferences  []string `json:"preferences"`  // 偏好（美食、购物、文化等）
	Keywords     []string `json:"keywords"`     // 提取的关键词
	OriginalText string   `json:"originalText"` // 原始文本
	Confidence   string   `json:"confidence"`   // 解析置信度 (high/medium/low)
}

// ParseTripText 解析旅行相关的语音文字
func ParseTripText(text string) *ParsedTripInfo {
	seg := api.GetSegmenter()

	result := &ParsedTripInfo{
		OriginalText: text,
		Travelers:    1, // 默认1人
		Confidence:   "low",
	}

	// 1. 使用gse分词
	words := seg.Cut(text, true) // true表示使用HMM

	// 2. 提取关键词
	result.Keywords = extractKeywordsFromWords(words)

	// 3. 提取目的地
	result.Destination = extractDestination(text, words)

	// 4. 提取时间信息
	startDate, endDate, duration := extractDateInfo(text)
	result.StartDate = startDate
	result.EndDate = endDate
	result.Duration = duration

	// 5. 提取预算
	result.Budget = extractBudget(text)

	// 6. 提取人数
	result.Travelers = extractTravelers(text)

	// 7. 提取偏好
	result.Preferences = extractPreferences(text, words)

	// 8. 评估置信度
	result.Confidence = evaluateConfidence(result)

	return result
}

// extractKeywordsFromWords 从分词结果提取关键词
func extractKeywordsFromWords(words []string) []string {
	keywords := []string{}

	// 旅行相关关键词
	travelKeywords := map[string]bool{
		"旅游": true, "旅行": true, "出游": true, "度假": true, "游玩": true,
		"美食": true, "购物": true, "文化": true, "自然": true, "风景": true,
		"海滩": true, "冒险": true, "放松": true, "摄影": true, "夜生活": true,
		"亲子": true, "博物馆": true, "历史": true, "古迹": true, "寺庙": true,
	}

	seen := make(map[string]bool)
	for _, word := range words {
		if travelKeywords[word] && !seen[word] && len(word) > 1 {
			keywords = append(keywords, word)
			seen[word] = true
		}
	}

	return keywords
}

// extractDestination 提取目的地
func extractDestination(text string, words []string) string {
	// 常见目的地列表
	destinations := []string{
		// 国内城市
		"北京", "上海", "广州", "深圳", "杭州", "成都", "重庆", "西安", "南京", "武汉",
		"天津", "苏州", "郑州", "长沙", "沈阳", "青岛", "无锡", "宁波", "昆明", "大连",
		"厦门", "合肥", "福州", "哈尔滨", "济南", "石家庄", "长春", "温州", "南昌", "贵阳",
		"三亚", "丽江", "大理", "桂林", "张家界", "西双版纳", "拉萨", "西藏", "九寨沟", "黄山",
		// 国外城市
		"东京", "大阪", "京都", "北海道", "冲绳", "首尔", "济州岛", "釜山",
		"曼谷", "清迈", "普吉岛", "芭提雅", "新加坡", "吉隆坡", "巴厘岛", "马尔代夫",
		"巴黎", "伦敦", "罗马", "威尼斯", "巴塞罗那", "阿姆斯特丹", "布拉格", "维也纳",
		"纽约", "洛杉矶", "旧金山", "拉斯维加斯", "迈阿密", "夏威夷",
		"悉尼", "墨尔本", "奥克兰", "迪拜", "伊斯坦布尔",
		// 港澳台
		"香港", "澳门", "台北", "高雄", "台中", "台南",
	}

	// 1. 直接匹配目的地列表
	for _, dest := range destinations {
		if strings.Contains(text, dest) {
			return dest
		}
	}

	// 2. 查找"去/到/飞"等动词后面的地点
	for i, word := range words {
		if word == "去" || word == "到" || word == "飞" || word == "游" || word == "玩" {
			if i+1 < len(words) {
				candidate := words[i+1]
				// 验证候选词（2-10个字符，排除常见动词）
				excludes := []string{"吃", "看", "玩", "住", "买", "逛", "玩儿", "看看", "一下", "了", "的"}
				isValid := len(candidate) >= 2 && len(candidate) <= 30
				for _, ex := range excludes {
					if candidate == ex {
						isValid = false
						break
					}
				}
				if isValid {
					return candidate
				}
			}
		}
	}

	// 3. 使用正则表达式提取
	patterns := []string{
		`(?:去|到|飞|游|玩)([^\s，。！？、]{2,10})(?:玩|游|旅游|旅行)?`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractDateInfo 提取日期信息
func extractDateInfo(text string) (startDate, endDate string, duration int) {
	now := time.Now()

	// 1. 提取持续天数
	duration = extractDuration(text)

	// 2. 检测相对时间
	if strings.Contains(text, "今天") {
		startDate = now.Format("2006-01-02")
	} else if strings.Contains(text, "明天") {
		startDate = now.AddDate(0, 0, 1).Format("2006-01-02")
	} else if strings.Contains(text, "后天") {
		startDate = now.AddDate(0, 0, 2).Format("2006-01-02")
	} else if strings.Contains(text, "大后天") {
		startDate = now.AddDate(0, 0, 3).Format("2006-01-02")
	} else if strings.Contains(text, "这周末") || strings.Contains(text, "本周末") {
		// 找到本周六
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7
		}
		startDate = now.AddDate(0, 0, daysUntilSaturday).Format("2006-01-02")
	} else if strings.Contains(text, "下周") || strings.Contains(text, "下星期") {
		startDate = now.AddDate(0, 0, 7).Format("2006-01-02")
	} else if strings.Contains(text, "下个月") || strings.Contains(text, "下月") {
		startDate = now.AddDate(0, 1, 0).Format("2006-01-02")
	}

	// 3. 提取具体日期
	// 格式1: YYYY-MM-DD, YYYY/MM/DD, YYYY.MM.DD
	re1 := regexp.MustCompile(`(\d{4})[-/.](\d{1,2})[-/.](\d{1,2})`)
	if matches := re1.FindStringSubmatch(text); len(matches) == 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		startDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	}

	// 格式2: MM月DD日
	re2 := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日?`)
	if startDate == "" {
		if matches := re2.FindStringSubmatch(text); len(matches) == 3 {
			month, _ := strconv.Atoi(matches[1])
			day, _ := strconv.Atoi(matches[2])
			year := now.Year()
			// 如果月份小于当前月份，可能是明年
			if month < int(now.Month()) || (month == int(now.Month()) && day < now.Day()) {
				year++
			}
			startDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		}
	}

	// 格式3: X号
	re3 := regexp.MustCompile(`(\d{1,2})号`)
	if startDate == "" {
		if matches := re3.FindStringSubmatch(text); len(matches) == 2 {
			day, _ := strconv.Atoi(matches[1])
			year := now.Year()
			month := now.Month()
			// 如果日期小于今天，可能是下个月
			if day < now.Day() {
				month++
				if month > 12 {
					month = 1
					year++
				}
			}
			startDate = time.Date(year, month, day, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		}
	}

	// 4. 如果有开始日期和持续天数，计算结束日期（含起止，持续N天则结束日=开始日+N-1）
	if startDate != "" && duration > 0 {
		start, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			endDate = start.AddDate(0, 0, duration-1).Format("2006-01-02")
		}
	}

	return
}

// extractDuration 提取持续天数
func extractDuration(text string) int {
	// 1. 匹配 "X天", "X日"
	re := regexp.MustCompile(`(\d+)\s*[天日]`)
	if matches := re.FindStringSubmatch(text); len(matches) > 1 {
		days, _ := strconv.Atoi(matches[1])
		if days > 0 && days < 365 {
			return days
		}
	}

	// 2. 匹配 "一周", "两周"
	weekPatterns := map[string]int{
		"一周": 7, "1周": 7, "两周": 14, "2周": 14,
		"三周": 21, "3周": 21, "四周": 28, "4周": 28,
	}
	for pattern, days := range weekPatterns {
		if strings.Contains(text, pattern) {
			return days
		}
	}

	// 3. 匹配中文数字
	chineseNums := map[string]int{
		"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
		"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
	}

	for cn, num := range chineseNums {
		if strings.Contains(text, cn+"天") || strings.Contains(text, cn+"日") {
			return num
		}
	}

	return 0
}

// extractBudget 提取预算
func extractBudget(text string) float64 {
	// 匹配各种预算表达
	patterns := []struct {
		regex      string
		multiplier float64
	}{
		{`预算\s*[:：]?\s*(\d+\.?\d*)万`, 10000},
		{`预算\s*[:：]?\s*(\d+\.?\d*)千`, 1000},
		{`预算\s*[:：]?\s*(\d+\.?\d*)k`, 1000},
		{`预算\s*[:：]?\s*(\d+\.?\d*)`, 1},
		{`(\d+\.?\d*)万[元块]?`, 10000},
		{`(\d+\.?\d*)千[元块]?`, 1000},
		{`(\d+\.?\d*)k`, 1000},
		{`(\d+\.?\d*)\s*[元块]`, 1},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			amount, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && amount > 0 {
				return amount * p.multiplier
			}
		}
	}

	return 0
}

// extractTravelers 提取旅行人数
func extractTravelers(text string) int {
	// 1. 匹配 "X人", "X个人"
	re := regexp.MustCompile(`(\d+)\s*[个]?人`)
	if matches := re.FindStringSubmatch(text); len(matches) > 1 {
		count, _ := strconv.Atoi(matches[1])
		if count > 0 && count < 100 {
			return count
		}
	}

	// 2. 检测关键词
	keywords := map[string]int{
		"一个人": 1, "独自": 1, "solo": 1, "自己": 1,
		"两个人": 2, "俩人": 2, "情侣": 2, "两人": 2,
		"三个人": 3, "仨人": 3, "三人": 3,
		"四个人": 4, "全家": 4, "家庭": 4, "四人": 4,
	}

	for keyword, count := range keywords {
		if strings.Contains(text, keyword) {
			return count
		}
	}

	return 1
}

// extractPreferences 提取偏好
func extractPreferences(text string, words []string) []string {
	preferences := []string{}
	seen := make(map[string]bool)

	// 偏好关键词映射 - 按优先级和精确度排序
	// 使用结构体来区分是否需要完整词匹配
	type PrefKeyword struct {
		keyword    string
		preference string
		exactMatch bool // true表示需要完整词匹配，false表示可以包含匹配
	}

	prefKeywords := []PrefKeyword{
		// 美食类 - 完整词匹配
		{"美食", "美食", true}, {"小吃", "美食", true}, {"餐厅", "美食", true},
		{"食物", "美食", true}, {"美味", "美食", true}, {"品尝", "美食", true},
		{"特色菜", "美食", true}, {"吃货", "美食", true},

		// 购物类 - 完整词匹配
		{"购物", "购物", true}, {"商场", "购物", true}, {"逛街", "购物", true},
		{"买买买", "购物", true}, {"shopping", "购物", true}, {"奥特莱斯", "购物", true},

		// 文化类 - 完整词匹配
		{"文化", "文化", true}, {"博物馆", "文化", true}, {"历史", "文化", true},
		{"古迹", "文化", true}, {"寺庙", "文化", true}, {"遗址", "文化", true},
		{"古城", "文化", true}, {"古镇", "文化", true},

		// 自然类 - 更精确的匹配
		{"自然", "自然", true}, {"风景", "自然", true}, {"登山", "自然", true},
		{"爬山", "自然", true}, {"海滩", "自然", true}, {"沙滩", "自然", true},
		{"海边", "自然", true}, {"湖边", "自然", true}, {"观景", "自然", true},
		{"看海", "自然", false}, {"看山", "自然", false}, {"赏景", "自然", true},

		// 冒险类 - 完整词匹配
		{"冒险", "冒险", true}, {"刺激", "冒险", true}, {"极限", "冒险", true},
		{"探险", "冒险", true}, {"户外", "冒险", true}, {"徒步", "冒险", true},

		// 放松类 - 完整词匹配
		{"放松", "放松", true}, {"休闲", "放松", true}, {"度假", "放松", true},
		{"spa", "放松", true}, {"按摩", "放松", true}, {"疗养", "放松", true},
		{"温泉", "放松", true},

		// 摄影类 - 完整词匹配
		{"摄影", "摄影", true}, {"拍照", "摄影", true}, {"打卡", "摄影", true},
		{"照相", "摄影", true},

		// 夜生活类 - 完整词匹配
		{"夜生活", "夜生活", true}, {"酒吧", "夜生活", true}, {"夜景", "夜生活", true},
		{"夜市", "夜生活", true}, {"夜店", "夜生活", true},

		// 亲子类 - 完整词匹配
		{"亲子", "亲子", true}, {"儿童", "亲子", true}, {"小孩", "亲子", true},
		{"游乐园", "亲子", true}, {"孩子", "亲子", true}, {"带娃", "亲子", true},
		{"迪士尼", "亲子", true}, {"乐高", "亲子", true}, {"欢乐谷", "亲子", true},
		{"一家", "亲子", false}, // "一家人"、"一家三口"等
	}

	// 优先检查完整文本中的长关键词（避免单字误匹配）
	for _, pk := range prefKeywords {
		if pk.exactMatch {
			// 完整词匹配：检查分词结果
			for _, word := range words {
				if word == pk.keyword && !seen[pk.preference] {
					preferences = append(preferences, pk.preference)
					seen[pk.preference] = true
					break
				}
			}
		} else {
			// 包含匹配：检查文本
			if strings.Contains(text, pk.keyword) && !seen[pk.preference] {
				preferences = append(preferences, pk.preference)
				seen[pk.preference] = true
			}
		}
	}

	return preferences
}

// evaluateConfidence 评估解析置信度
func evaluateConfidence(info *ParsedTripInfo) string {
	score := 0

	// 目的地最重要
	if info.Destination != "" {
		score += 3
	}
	// 日期很重要
	if info.StartDate != "" {
		score += 2
	}
	// 天数重要
	if info.Duration > 0 {
		score += 2
	}
	// 预算次要
	if info.Budget > 0 {
		score += 1
	}
	// 偏好加分
	if len(info.Preferences) > 0 {
		score += 1
	}
	// 关键词加分
	if len(info.Keywords) > 0 {
		score += 1
	}

	if score >= 7 {
		return "high"
	} else if score >= 4 {
		return "medium"
	}
	return "low"
}
