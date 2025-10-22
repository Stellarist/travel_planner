# 语音文字解析功能 - 使用说明

## 技术栈

- **后端**: Go + **gse** (go-ego/gse) 中文分词库
- **前端**: TypeScript + React

## 功能说明

将语音识别后的文字解析成结构化的旅行信息，包括：

- ✅ 目的地
- ✅ 出发/结束日期
- ✅ 持续天数
- ✅ 预算
- ✅ 旅行人数
- ✅ 偏好（美食、购物、文化等）
- ✅ 关键词提取
- ✅ 置信度评估

## API 接口

### 解析文本

**POST** `/api/parser/parse`

需要认证（Bearer Token）

**请求**:

```json
{
  "text": "我想下周去北京玩5天，预算5000元"
}
```

**响应**:

```json
{
  "success": true,
  "message": "解析成功",
  "data": {
    "destination": "北京",
    "startDate": "2025-10-30",
    "endDate": "2025-11-04",
    "duration": 5,
    "budget": 5000,
    "travelers": 1,
    "preferences": [],
    "keywords": [],
    "originalText": "我想下周去北京玩5天，预算5000元",
    "confidence": "high"
  }
}
```

## 测试

### 后端测试

```powershell
# 运行测试程序
cd c:\Data\repos\travel_planner\backend
go run .\cmd\test_parser\main.go
```

### API 测试

```powershell
# 1. 启动服务
go run main.go

# 2. 登录获取token
$body = @{username="test"; password="test123"} | ConvertTo-Json
$res = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method Post -Body $body -ContentType "application/json"
$token = $res.token

# 3. 测试解析
$headers = @{Authorization="Bearer $token"; "Content-Type"="application/json"}
$parseBody = @{text="明天去上海，两个人，喜欢美食和购物"} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/parser/parse" -Method Post -Headers $headers -Body $parseBody
```

## 支持的文本格式

### 测试用例

```
✅ "我想下周去北京玩5天，预算5000元"
✅ "明天去上海，两个人，喜欢美食和购物"
✅ "3月15日到成都旅游，一周时间，预算1万，喜欢文化和摄影"
✅ "下个月带孩子去三亚度假，全家4个人，10天左右，预算2万"
✅ "周末去杭州放松一下，两天一夜"
✅ "去东京玩，大概5天，3个人，喜欢动漫和美食"
✅ "下周三飞巴厘岛，预算8千块，情侣旅行，想看海滩"
```

## 前端集成

### 类型定义

类型已定义在 `frontend/src/shared/types.ts`:

- `ParsedTripInfo`
- `ParseTextRequest`
- `ParseTextResponse`

### 调用示例

```typescript
import { apiPost } from '@/shared/utils'
import type { ParsedTripInfo, ParseTextResponse } from '@/shared/types'

const parseText = async (text: string): Promise<ParsedTripInfo | null> => {
  const response: ParseTextResponse = await apiPost('/api/parser/parse', { text })
  
  if (response.success && response.data) {
    return response.data
  }
  return null
}

// 使用
const result = await parseText("明天去上海，两个人，喜欢美食和购物")
console.log(result)
```

## 识别能力

### 1. 目的地识别

支持 50+ 热门城市：

- 国内：北京、上海、广州、深圳、杭州、成都、重庆、西安、三亚、丽江...
- 国外：东京、大阪、首尔、曼谷、新加坡、巴黎、伦敦、纽约...

### 2. 时间识别

- 相对时间：今天、明天、后天、下周、下个月、周末
- 具体日期：3月15日、2025-03-15、15号
- 持续时间：5天、一周、10天

### 3. 预算识别

- 5000元、5000块
- 1万、5千
- 5k、10k

### 4. 人数识别

- 数字：2人、3个人
- 文字：一个人、两个人、情侣、全家

### 5. 偏好识别

9大偏好类别：

- 美食、购物、文化、自然
- 冒险、放松、摄影、夜生活、亲子

## 下一步

可以将解析结果直接用于：

1. 自动填充旅行规划表单
2. 调用AI生成详细行程
3. 推荐相关景点和活动
