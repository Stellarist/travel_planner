# 地图搜索功能实现 - 类似地图软件的搜索体验

## 实现时间

2025-10-23

## 功能说明

模仿主流地图软件（高德、百度地图等）的搜索体验：

1. **搜索联想**：输入时展示候选区域列表（最多8个）
2. **选择聚焦**：点击候选项后，地图聚焦到该区域并显示标记
3. **附近推荐**：自动搜索该区域附近5公里内的游玩场所
4. **导航功能**：每个景点信息窗口包含"导航"按钮，跳转到高德地图

## 技术实现

### 1. 新增插件加载

```typescript
plugins: ['AMap.PlaceSearch', 'AMap.Geocoder', 'AMap.InfoWindow', 'AMap.AutoComplete']
```

新增 `AMap.AutoComplete` 插件支持搜索联想。

### 2. 新增状态管理

```typescript
const areaMarkerRef = useRef<any>(null)  // 区域标记（与景点标记分开）
const [suggestions, setSuggestions] = useState<Array<...>>([])  // 候选列表
const [showSuggestions, setShowSuggestions] = useState(false)    // 是否显示下拉
const [activeSuggestIndex, setActiveSuggestIndex] = useState(0)  // 当前高亮索引
```

### 3. 核心函数

#### `fetchSuggestions(keyword: string)`

- 使用 `AMap.AutoComplete` 获取候选区域
- 过滤并限制最多8个结果
- 更新 `suggestions` 状态并显示下拉框

#### `selectSuggestion(item)`

- 清除旧的区域标记，创建新的区域标记
- 地图聚焦到该区域（zoom=12）
- 弹出信息窗口，包含"导航"按钮
- 调用 `searchNearbyAttractions` 搜索附近游玩场所

#### `searchNearbyAttractions(center: [lng, lat])`

- 使用 `AMap.PlaceSearch.searchNearBy` 搜索中心点5公里内的景点
- 结合用户筛选的类型
- 清除旧的景点标记，添加新标记
- 调整视野包含所有标记（区域标记 + 景点标记）

#### `getNavUrl(lng, lat, name)`

- 构建高德地图导航URL
- 格式：`https://uri.amap.com/navigation?to={lng},{lat},{name}&mode=car&policy=1`
- 在新窗口打开，自动使用用户当前位置作为起点

### 4. UI改进

#### 搜索输入框

- 支持键盘导航：
  - `Enter`：执行搜索
  - `ArrowDown/ArrowUp`：切换候选项
  - `Escape`：关闭下拉框
- Placeholder更新为："搜索区域或景点 (如: 故宫 / 西湖)"

#### 候选下拉框 (`.suggestions-dropdown`)

- 绝对定位在输入框下方
- 最多显示8个候选项
- 高亮当前选中项（`.active`）
- 每项显示名称 + 地址
- 支持鼠标悬停和点击

#### 信息窗口增强

- 区域标记和景点标记的信息窗口都包含"导航"按钮
- 导航按钮样式：蓝紫色渐变背景，圆角，白色文字

### 5. CSS新增样式

```css
.suggestions-dropdown { ... }
.suggestion-item { ... }
.suggestion-item:hover, .suggestion-item.active { ... }
.suggestion-name { ... }
.suggestion-address { ... }
```

## 用户交互流程

1. 用户输入"西湖" → 点击搜索或按Enter
2. 显示候选列表（如：西湖风景区、西湖公园、西湖文化广场...）
3. 用户选择"西湖风景区" → 地图聚焦到西湖，显示标记和信息窗
4. 自动搜索西湖附近5公里内的景点（雷峰塔、断桥、岳王庙...）
5. 侧边栏展示景点列表，地图上显示标记
6. 用户点击任意标记 → 显示详情 + "导航"按钮
7. 点击导航 → 跳转到高德地图导航页面

## 文件变更清单

### 修改文件

1. `frontend/src/components/ExploreMap.tsx`
   - 新增联想搜索、区域选择、附近推荐、导航功能
   - 重构搜索逻辑，去除旧的多步判断（城市前缀匹配等）
   - 约+125行，-72行

2. `frontend/src/components/ExploreMap.css`
   - 新增候选下拉框样式
   - search-area 添加 `position: relative`
   - 约+47行

### 删除的代码

- 移除 `HOT_CITIES` 常量（不再需要前缀匹配）
- 移除复杂的城市/关键词解析逻辑（4步判断）

## 优化点

1. **性能优化**
   - 限制候选列表最多8个
   - 保持区域标记和景点标记分离，避免重复创建

2. **用户体验**
   - 键盘导航支持（↑↓方向键、Enter、Esc）
   - 鼠标悬停高亮
   - 默认选中第一个候选项（可直接按Enter快速选择）

3. **视野管理**
   - 自动调整视野包含所有标记（区域 + 景点）
   - 使用 `setFitView` 优先，兜底使用 `setCenter`

## 后续可优化

1. 搜索历史记录（本地存储）
2. 热门区域推荐（无输入时展示）
3. 搜索结果分页（当前仅显示前20个景点）
4. 导航方式选择（驾车、步行、公交、骑行）
5. 语音输入集成（科大讯飞API，已在计划中）

## 配置需求

确保 `config.json` 中配置了有效的高德地图API Key：

```json
{
  "frontend": {
    "amapKey": "your_actual_amap_key"
  }
}
```
