export const BUDGET_ITEMS_PER_PAGE = 5

export const AVAILABLE_PREFERENCES: readonly string[] = ['美食', '动漫', '亲子', '历史', '自然', '购物', '冒险'] as const

export const CATEGORY_NAMES: readonly string[] = ['食物', '交通', '住宿', '购物', '活动', '其他'] as const

export const CATEGORY_MAP: Record<string, string> = {
    Meals: '食物',
    Transport: '交通',
    Accommodation: '住宿',
    Activities: '活动',
    Shopping: '购物',
    Other: '其他',
}

export const CATEGORY_COLORS: Record<string, string> = {
    '食物': '#ff7b7b', Meals: '#ff7b7b',
    '交通': '#ffa94d', Transport: '#ffa94d',
    '住宿': '#667eea', Accommodation: '#667eea',
    '活动': '#7bd389', Activities: '#7bd389',
    '购物': '#764ba2', Shopping: '#764ba2',
    '其他': '#9ad0ff', Other: '#9ad0ff',
    '其他(小额)': '#c0c0c0',
}

export const ATTRACTION_TYPES = [
    { value: 'scenic', label: '🏞️ 自然风光', keywords: '风景区|公园|山|湖|海' },
    { value: 'historic', label: '🏛️ 历史古迹', keywords: '古迹|博物馆|寺庙|故居' },
    { value: 'theme', label: '🎡 主题乐园', keywords: '乐园|游乐场|主题公园' },
    { value: 'food', label: '🍜 美食街区', keywords: '美食|小吃街|餐饮' },
    { value: 'shopping', label: '🛍️ 购物中心', keywords: '商场|购物中心|步行街' },
]
