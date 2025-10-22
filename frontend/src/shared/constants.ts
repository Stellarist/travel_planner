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

