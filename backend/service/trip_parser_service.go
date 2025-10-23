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
	// 常见目的地列表（大幅扩充）
	destinations := []string{
		// 国内直辖市
		"北京", "上海", "天津", "重庆",

		// 国内省会及重点城市
		"广州", "深圳", "杭州", "成都", "西安", "南京", "武汉", "苏州", "郑州", "长沙",
		"沈阳", "青岛", "无锡", "宁波", "昆明", "大连", "厦门", "合肥", "福州", "哈尔滨",
		"济南", "石家庄", "长春", "温州", "南昌", "贵阳", "南宁", "太原", "兰州", "乌鲁木齐",
		"呼和浩特", "银川", "西宁", "海口", "珠海", "佛山", "东莞", "中山", "惠州", "汕头",
		"常州", "扬州", "徐州", "南通", "镇江", "盐城", "泰州", "淄博", "烟台", "潍坊",
		"威海", "临沂", "绍兴", "嘉兴", "湖州", "金华", "台州", "洛阳", "开封", "商丘",
		"安阳", "新乡", "南阳", "株洲", "湘潭", "岳阳", "常德", "衡阳", "邵阳", "芜湖",
		"蚌埠", "淮南", "马鞍山", "漳州", "泉州", "莆田", "三明", "南平", "龙岩",

		// 热门旅游城市/景区
		"三亚", "丽江", "大理", "桂林", "张家界", "西双版纳", "拉萨", "西藏", "九寨沟", "黄山",
		"峨眉山", "华山", "泰山", "衡山", "普陀山", "武夷山", "庐山", "五台山", "雁荡山",
		"凤凰", "阳朔", "鼓浪屿", "千岛湖", "乌镇", "周庄", "同里", "西塘", "南浔", "朱家角",
		"婺源", "黄果树", "香格里拉", "稻城", "亚丁", "喀纳斯", "敦煌", "莫高窟", "青海湖",
		"茶卡盐湖", "呼伦贝尔", "长白山", "雪乡", "漠河", "阿尔山", "额济纳", "鸣沙山",
		"月牙泉", "天山", "吐鲁番", "喀什", "伊犁", "那拉提", "巴音布鲁克",

		// 港澳台
		"香港", "澳门", "台北", "高雄", "台中", "台南", "花莲", "垦丁", "日月潭", "阿里山",

		// 日本
		"东京", "大阪", "京都", "北海道", "冲绳", "奈良", "神户", "名古屋", "福冈", "札幌",
		"横滨", "镰仓", "箱根", "富士山", "长崎", "广岛", "仙台", "金泽", "轻井泽",

		// 韩国
		"首尔", "济州岛", "釜山", "仁川", "江原道", "大邱", "光州",

		// 东南亚
		"曼谷", "清迈", "普吉岛", "芭提雅", "苏梅岛", "甲米", "新加坡", "吉隆坡", "槟城",
		"巴厘岛", "雅加达", "日惹", "龙目岛", "马尔代夫", "河内", "胡志明", "芽庄", "岘港",
		"暹粒", "吴哥窟", "金边", "万象", "琅勃拉邦", "仰光", "蒲甘", "文莱",

		// 欧洲
		"巴黎", "伦敦", "罗马", "威尼斯", "佛罗伦萨", "米兰", "那不勒斯", "比萨", "五渔村",
		"巴塞罗那", "马德里", "塞维利亚", "格拉纳达", "阿姆斯特丹", "布鲁塞尔", "布拉格",
		"维也纳", "布达佩斯", "慕尼黑", "柏林", "法兰克福", "科隆", "海德堡", "苏黎世",
		"日内瓦", "因特拉肯", "卢塞恩", "雅典", "圣托里尼", "莫斯科", "圣彼得堡", "里斯本",
		"波尔图", "哥本哈根", "斯德哥尔摩", "奥斯陆", "赫尔辛基", "雷克雅未克", "都柏林",
		"爱丁堡", "曼彻斯特", "牛津", "剑桥", "巴斯", "约克",

		// 美洲
		"纽约", "洛杉矶", "旧金山", "拉斯维加斯", "西雅图", "芝加哥", "波士顿", "迈阿密",
		"奥兰多", "华盛顿", "费城", "圣迭戈", "夏威夷", "阿拉斯加", "黄石公园", "大峡谷",
		"多伦多", "温哥华", "蒙特利尔", "魁北克", "墨西哥城", "坎昆", "布宜诺斯艾利斯",
		"里约热内卢", "圣保罗", "秘鲁", "马丘比丘", "复活节岛",

		// 大洋洲
		"悉尼", "墨尔本", "布里斯班", "黄金海岸", "凯恩斯", "珀斯", "阿德莱德", "奥克兰",
		"皇后镇", "基督城", "惠灵顿", "大堡礁", "乌鲁鲁", "塔斯马尼亚",

		// 中东/非洲
		"迪拜", "阿布扎比", "多哈", "伊斯坦布尔", "开罗", "卢克索", "摩洛哥", "卡萨布兰卡",
		"马拉喀什", "突尼斯", "开普敦", "毛里求斯", "塞舌尔", "肯尼亚", "坦桑尼亚",
	}

	// 1. 按长度排序，优先匹配长地名（避免"西安"被"西"误匹配）
	sortedDests := make([]string, len(destinations))
	copy(sortedDests, destinations)
	for i := 0; i < len(sortedDests); i++ {
		for j := i + 1; j < len(sortedDests); j++ {
			if len(sortedDests[i]) < len(sortedDests[j]) {
				sortedDests[i], sortedDests[j] = sortedDests[j], sortedDests[i]
			}
		}
	}

	// 2. 直接匹配目的地列表
	for _, dest := range sortedDests {
		if strings.Contains(text, dest) {
			return dest
		}
	}

	// 3. 查找"去/到/飞/游/玩/想去/打算去"等动词后面的地点
	actionWords := []string{"去", "到", "飞", "游", "玩", "想去", "打算去", "准备去", "计划去", "前往"}
	for i, word := range words {
		for _, action := range actionWords {
			if word == action || strings.Contains(word, action) {
				if i+1 < len(words) {
					candidate := words[i+1]
					// 验证候选词（2-15个字符，排除常见动词）
					excludes := []string{"吃", "看", "玩", "住", "买", "逛", "玩儿", "看看", "一下", "了", "的", "啊", "呀", "吧", "呢"}
					isValid := len(candidate) >= 2 && len(candidate) <= 45 // 允许更长的地名
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
	}

	// 4. 使用正则表达式提取（更灵活的模式）
	patterns := []string{
		`(?:去|到|飞|游|玩|想去|打算去|准备去|计划去|前往)([^\s，。！？、]{2,15})(?:玩|游|旅游|旅行|度假)?`,
		`在([^\s，。！？、]{2,15})(?:玩|游|旅游|旅行|度假)`,
		`([^\s，。！？、]{2,15})(?:游|之旅|行|自由行|自驾游|跟团游)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			// 再次验证提取的地点不是常见动词
			candidate := matches[1]
			excludes := []string{"旅游", "旅行", "度假", "出游", "游玩", "吃喝", "玩乐"}
			isValid := true
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

	return ""
}

// extractDateInfo 提取日期信息
func extractDateInfo(text string) (startDate, endDate string, duration int) {
	now := time.Now()

	// 1. 提取持续天数（支持更多表达）
	duration = extractDuration(text)

	// 2. 检测相对时间（大幅扩充）
	relativeTimeMap := map[string]func() string{
		"今天":   func() string { return now.Format("2006-01-02") },
		"今日":   func() string { return now.Format("2006-01-02") },
		"今儿":   func() string { return now.Format("2006-01-02") },
		"明天":   func() string { return now.AddDate(0, 0, 1).Format("2006-01-02") },
		"明日":   func() string { return now.AddDate(0, 0, 1).Format("2006-01-02") },
		"明儿":   func() string { return now.AddDate(0, 0, 1).Format("2006-01-02") },
		"后天":   func() string { return now.AddDate(0, 0, 2).Format("2006-01-02") },
		"大后天":  func() string { return now.AddDate(0, 0, 3).Format("2006-01-02") },
		"这周末":  func() string { return getWeekend(now) },
		"本周末":  func() string { return getWeekend(now) },
		"周末":   func() string { return getWeekend(now) },
		"这个周末": func() string { return getWeekend(now) },
		"下周":   func() string { return now.AddDate(0, 0, 7).Format("2006-01-02") },
		"下星期":  func() string { return now.AddDate(0, 0, 7).Format("2006-01-02") },
		"下个星期": func() string { return now.AddDate(0, 0, 7).Format("2006-01-02") },
		"下周末":  func() string { return getWeekend(now.AddDate(0, 0, 7)) },
		"下个月":  func() string { return now.AddDate(0, 1, 0).Format("2006-01-02") },
		"下月":   func() string { return now.AddDate(0, 1, 0).Format("2006-01-02") },
		"月底":   func() string { return getEndOfMonth(now) },
		"月初":   func() string { return getStartOfMonth(now) },
		"年底":   func() string { return time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.Local).Format("2006-01-02") },
		"春节":   func() string { return getSpringFestival(now.Year()) },
		"国庆":   func() string { return time.Date(now.Year(), 10, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02") },
		"五一":   func() string { return time.Date(now.Year(), 5, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02") },
		"十一":   func() string { return time.Date(now.Year(), 10, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02") },
		"元旦":   func() string { return time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02") },
		"中秋":   func() string { return getMidAutumnFestival(now.Year()) },
		"端午":   func() string { return getDragonBoatFestival(now.Year()) },
		"清明":   func() string { return getTombSweepingDay(now.Year()) },
	}

	for keyword, dateFunc := range relativeTimeMap {
		if strings.Contains(text, keyword) {
			startDate = dateFunc()
			break
		}
	}

	// 3. 检测"N天后"、"N周后"、"N个月后"
	afterPatterns := []struct {
		pattern string
		handler func(int) string
	}{
		{
			pattern: `([一二三四五六七八九十两1-9]\d?)(?:个)?天(?:后|之后|以后)`,
			handler: func(n int) string {
				return now.AddDate(0, 0, n).Format("2006-01-02")
			},
		},
		{
			pattern: `([一二三四五六七八九十两1-9])(?:个)?周(?:后|之后|以后)`,
			handler: func(n int) string {
				return now.AddDate(0, 0, n*7).Format("2006-01-02")
			},
		},
		{
			pattern: `([一二三四五六七八九十两1-9]|1[0-2])(?:个)?月(?:后|之后|以后)`,
			handler: func(n int) string {
				return now.AddDate(0, n, 0).Format("2006-01-02")
			},
		},
	}

	if startDate == "" {
		for _, ap := range afterPatterns {
			re := regexp.MustCompile(ap.pattern)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				num := parseChineseNumber(matches[1])
				startDate = ap.handler(num)
				break
			}
		}
	}

	// 4. 提取具体日期
	// 格式1: YYYY-MM-DD, YYYY/MM/DD, YYYY.MM.DD
	re1 := regexp.MustCompile(`(\d{4})[-/.](\d{1,2})[-/.](\d{1,2})`)
	if startDate == "" {
		if matches := re1.FindStringSubmatch(text); len(matches) == 4 {
			year, _ := strconv.Atoi(matches[1])
			month, _ := strconv.Atoi(matches[2])
			day, _ := strconv.Atoi(matches[3])
			startDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		}
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

	// 5. 如果没有明确的开始日期但有持续天数，默认从今天开始
	if startDate == "" && duration > 0 {
		startDate = now.Format("2006-01-02")
	}

	// 6. 如果有开始日期和持续天数，计算结束日期
	// "玩N天" 表示从开始日期起N天后结束
	// 例如：10月1日玩20天 = 10月1日开始，10月21日结束
	if startDate != "" && duration > 0 {
		start, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			endDate = start.AddDate(0, 0, duration).Format("2006-01-02")
		}
	}

	return
}

// 辅助函数：获取周末日期（本周六）
func getWeekend(t time.Time) string {
	daysUntilSaturday := (6 - int(t.Weekday()) + 7) % 7
	if daysUntilSaturday == 0 {
		daysUntilSaturday = 7
	}
	return t.AddDate(0, 0, daysUntilSaturday).Format("2006-01-02")
}

// 辅助函数：获取月底日期
func getEndOfMonth(t time.Time) string {
	nextMonth := t.AddDate(0, 1, 0)
	firstDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDayThisMonth := firstDayNextMonth.AddDate(0, 0, -1)
	return lastDayThisMonth.Format("2006-01-02")
}

// 辅助函数：获取月初日期
func getStartOfMonth(t time.Time) string {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
}

// 辅助函数：获取春节日期（简化处理，返回2月初）
func getSpringFestival(year int) string {
	// 简化处理，实际春节日期每年不同，这里返回2月1日
	return time.Date(year, 2, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
}

// 辅助函数：获取中秋节日期（简化处理，返回9月中旬）
func getMidAutumnFestival(year int) string {
	return time.Date(year, 9, 15, 0, 0, 0, 0, time.Local).Format("2006-01-02")
}

// 辅助函数：获取端午节日期（简化处理，返回6月初）
func getDragonBoatFestival(year int) string {
	return time.Date(year, 6, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
}

// 辅助函数：获取清明节日期（固定4月5日左右）
func getTombSweepingDay(year int) string {
	return time.Date(year, 4, 5, 0, 0, 0, 0, time.Local).Format("2006-01-02")
}

// extractDuration 提取持续天数
func extractDuration(text string) int {
	// 0. 先用简单映射匹配常见表达（优先级最高）
	simpleDurationMap := map[string]int{
		"一天": 1, "两天": 2, "三天": 3, "四天": 4, "五天": 5,
		"六天": 6, "七天": 7, "八天": 8, "九天": 9, "十天": 10,
		"半天": 0, // 半天按0处理
		"一日": 1, "两日": 2, "三日": 3, "四日": 4, "五日": 5,
		"十一天": 11, "十二天": 12, "十五天": 15, "二十天": 20, "三十天": 30,
	}

	// 按长度从长到短排序，优先匹配长的
	keywords := make([]string, 0, len(simpleDurationMap))
	for k := range simpleDurationMap {
		keywords = append(keywords, k)
	}
	for i := 0; i < len(keywords); i++ {
		for j := i + 1; j < len(keywords); j++ {
			if len(keywords[i]) < len(keywords[j]) {
				keywords[i], keywords[j] = keywords[j], keywords[i]
			}
		}
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return simpleDurationMap[keyword]
		}
	}

	// 1. 匹配 "X天", "X日", "X天N夜"
	patterns := []string{
		`(\d+)\s*[天日]`,       // 5天、3日、20天
		`(\d+)\s*天\d*夜`,      // 5天4夜、3天2夜
		`([一二三四五六七八九十两百]+)天`, // 三天、五天、二十天、三十五天
		`([一二三四五六七八九十两百]+)日`, // 三日、二十日
		`([一二三四五六七八九十两百]+)天[一二三四五六七八九十两百]*夜`, // 五天四夜
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			days := 0
			// 尝试解析阿拉伯数字
			if d, err := strconv.Atoi(matches[1]); err == nil {
				days = d
			} else {
				// 解析中文数字
				days = parseChineseNumberForDuration(matches[1])
			}
			if days > 0 && days < 365 {
				return days
			}
		}
	}

	// 2. 匹配周数："一周"、"两周"、"N周"、"N个星期"
	weekPatterns := map[string]int{
		"一周":   7,
		"1周":   7,
		"一个星期": 7,
		"一星期":  7,
		"两周":   14,
		"2周":   14,
		"两个星期": 14,
		"三周":   21,
		"3周":   21,
		"三个星期": 21,
		"四周":   28,
		"4周":   28,
		"四个星期": 28,
		"半个月":  15,
		"半月":   15,
		"一个月":  30,
		"一月":   30,
		"两个月":  60,
		"三个月":  90,
		"一个礼拜": 7,
		"两个礼拜": 14,
		"一礼拜":  7,
		"两礼拜":  14,
	}

	for pattern, days := range weekPatterns {
		if strings.Contains(text, pattern) {
			return days
		}
	}

	// 3. 使用正则匹配更灵活的周数表达
	weekRegex := regexp.MustCompile(`([一二三四五六七八九十两1-9])(?:个)?(?:周|星期|礼拜)`)
	if matches := weekRegex.FindStringSubmatch(text); len(matches) > 1 {
		weeks := parseChineseNumber(matches[1])
		if weeks > 0 && weeks <= 52 {
			return weeks * 7
		}
	}

	// 4. 匹配"N天M夜"中的数字
	dayNightRegex := regexp.MustCompile(`(\d+)天\d+夜`)
	if matches := dayNightRegex.FindStringSubmatch(text); len(matches) > 1 {
		if days, err := strconv.Atoi(matches[1]); err == nil && days > 0 && days < 365 {
			return days
		}
	}

	return 0
}

// parseChineseNumberForDuration 专门用于持续时间的中文数字解析
func parseChineseNumberForDuration(s string) int {
	chineseMap := map[string]int{
		"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
		"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
		"两": 2, "俩": 2,
		"半": 0, // 半天特殊处理
	}

	if len(s) == 1 {
		if num, ok := chineseMap[s]; ok {
			return num
		}
	}

	// 处理"十X"情况（十一到十九）
	if strings.HasPrefix(s, "十") && len(s) > 1 {
		unit := string([]rune(s)[1:])
		if num, ok := chineseMap[unit]; ok {
			return 10 + num
		}
		return 10
	}

	// 处理"X十"情况（二十、三十等）
	if strings.HasSuffix(s, "十") && len(s) > 1 {
		tens := string([]rune(s)[0 : len(s)-1])
		if num, ok := chineseMap[tens]; ok {
			return num * 10
		}
	}

	// 处理"X十Y"情况（二十一、三十五等）
	if strings.Contains(s, "十") {
		parts := strings.Split(s, "十")
		if len(parts) == 2 {
			tens := 0
			unit := 0
			if parts[0] != "" {
				tens = chineseMap[parts[0]]
			} else {
				tens = 1 // "十"开头表示10
			}
			if parts[1] != "" {
				unit = chineseMap[parts[1]]
			}
			return tens*10 + unit
		}
	}

	return 1
}

// extractBudget 提取预算
func extractBudget(text string) float64 {
	// 1. 先匹配中文数字的预算表达（简化版，直接识别常见表达）
	// 包含带单位（元、块、钱）和不带单位的版本
	simpleChinesePatterns := map[string]float64{
		// 千位级别
		"一千元": 1000, "一千块": 1000, "一千": 1000,
		"两千元": 2000, "两千块": 2000, "两千": 2000,
		"三千元": 3000, "三千块": 3000, "三千": 3000,
		"四千元": 4000, "四千块": 4000, "四千": 4000,
		"五千元": 5000, "五千块": 5000, "五千": 5000,
		"六千元": 6000, "六千块": 6000, "六千": 6000,
		"七千元": 7000, "七千块": 7000, "七千": 7000,
		"八千元": 8000, "八千块": 8000, "八千": 8000,
		"九千元": 9000, "九千块": 9000, "九千": 9000,

		// 万位级别
		"一万元": 10000, "一万块": 10000, "一万": 10000,
		"两万元": 20000, "两万块": 20000, "两万": 20000,
		"三万元": 30000, "三万块": 30000, "三万": 30000,
		"四万元": 40000, "四万块": 40000, "四万": 40000,
		"五万元": 50000, "五万块": 50000, "五万": 50000,
		"六万元": 60000, "六万块": 60000, "六万": 60000,
		"七万元": 70000, "七万块": 70000, "七万": 70000,
		"八万元": 80000, "八万块": 80000, "八万": 80000,
		"九万元": 90000, "九万块": 90000, "九万": 90000,
		"十万元": 100000, "十万块": 100000, "十万": 100000,

		// 中间数
		"一万五": 15000, "两万五": 25000, "三万五": 35000,
	}

	// 按长度从长到短排序，优先匹配长的（带单位的优先）
	keywords := make([]string, 0, len(simpleChinesePatterns))
	for k := range simpleChinesePatterns {
		keywords = append(keywords, k)
	}
	for i := 0; i < len(keywords); i++ {
		for j := i + 1; j < len(keywords); j++ {
			if len(keywords[i]) < len(keywords[j]) {
				keywords[i], keywords[j] = keywords[j], keywords[i]
			}
		}
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return simpleChinesePatterns[keyword]
		}
	}

	// 2. 使用正则匹配更复杂的中文数字表达
	chineseBudgetPatterns := []struct {
		pattern    string
		multiplier float64
	}{
		// 匹配：三千元、五千块、八千
		{`([一二三四五六七八九十两百]+)千\s*[元块钱]?`, 1000},
		// 匹配：一万元、两万块、五万
		{`([一二三四五六七八九十两百]+)万\s*[元块钱]?`, 10000},
		// 匹配：预算：三千、预算五万
		{`预算\s*[:：]?\s*([一二三四五六七八九十两百]+)千`, 1000},
		{`预算\s*[:：]?\s*([一二三四五六七八九十两百]+)万`, 10000},
	}

	for _, p := range chineseBudgetPatterns {
		re := regexp.MustCompile(p.pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			// 解析中文数字
			amount := parseChineseBudgetNumber(matches[1])
			if amount > 0 {
				return amount * p.multiplier
			}
		}
	}

	// 2. 匹配阿拉伯数字的预算表达
	arabicBudgetPatterns := []struct {
		regex      string
		multiplier float64
	}{
		{`预算\s*[:：]?\s*(\d+\.?\d*)万`, 10000},
		{`预算\s*[:：]?\s*(\d+\.?\d*)千`, 1000},
		{`预算\s*[:：]?\s*(\d+\.?\d*)k`, 1000},
		{`预算\s*[:：]?\s*(\d+\.?\d*)`, 1},
		{`(\d+\.?\d*)万\s*[元块钱]?`, 10000},
		{`(\d+\.?\d*)千\s*[元块钱]?`, 1000},
		{`(\d+\.?\d*)k`, 1000},
		{`(\d+\.?\d*)\s*[元块]`, 1},
	}

	for _, p := range arabicBudgetPatterns {
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

// parseChineseBudgetNumber 解析预算中的中文数字（支持更复杂的表达）
func parseChineseBudgetNumber(s string) float64 {
	// 基础数字映射
	digitMap := map[rune]int{
		'一': 1, '二': 2, '三': 3, '四': 4, '五': 5,
		'六': 6, '七': 7, '八': 8, '九': 9,
		'零': 0, '两': 2, '俩': 2,
	}

	unitMap := map[rune]int{
		'十': 10,
		'百': 100,
		'千': 1000,
		'万': 10000,
	}

	result := 0
	currentNum := 0
	hasDigit := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if digit, ok := digitMap[r]; ok {
			currentNum = digit
			hasDigit = true
		} else if unit, ok := unitMap[r]; ok {
			if !hasDigit {
				currentNum = 1 // "十"、"百"等前面没有数字时默认为1
			}
			result += currentNum * unit
			currentNum = 0
			hasDigit = false
		}
	}

	// 处理最后剩余的数字
	if hasDigit {
		result += currentNum
	}

	return float64(result)
}

// extractTravelers 提取旅行人数
func extractTravelers(text string) int {
	// 1. 优先匹配关键词（更准确）
	keywords := map[string]int{
		"一个人":  1,
		"独自":   1,
		"solo": 1,
		"自己":   1,
		"单人":   1,
		"两个人":  2,
		"俩人":   2,
		"两人":   2,
		"情侣":   2,
		"夫妻":   2,
		"三个人":  3,
		"仨人":   3,
		"三人":   3,
		"四个人":  4,
		"全家":   4,
		"家庭":   4,
		"四人":   4,
		"五个人":  5,
		"五人":   5,
		"六个人":  6,
		"六人":   6,
		"一家三口": 3,
		"一家四口": 4,
		"一家五口": 5,
	}

	for keyword, count := range keywords {
		if strings.Contains(text, keyword) {
			return count
		}
	}

	// 2. 匹配中文数字 + 人/个人
	chinesePersonPatterns := []string{
		`([一二三四五六七八九十两]+)(?:个)?人`,
	}

	for _, pattern := range chinesePersonPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			count := parseChineseNumber(matches[1])
			if count > 0 && count < 100 {
				return count
			}
		}
	}

	// 3. 匹配阿拉伯数字 + 人/个人
	re := regexp.MustCompile(`(\d+)\s*[个]?人`)
	if matches := re.FindStringSubmatch(text); len(matches) > 1 {
		count, _ := strconv.Atoi(matches[1])
		if count > 0 && count < 100 {
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
