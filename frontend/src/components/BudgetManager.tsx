import { useState, useEffect } from 'react'
import '../styles/common.css'
import './TripPlanner.css'
import './BudgetManager.css'
import { apiPost, apiGet, formatAmount, useSpeechRecognition } from '../shared/utils'
import type { Expense } from '../shared/types'
import { CATEGORY_MAP, CATEGORY_COLORS, BUDGET_ITEMS_PER_PAGE, CATEGORY_NAMES } from '../shared/constants'

export default function BudgetManager() {
    const [list, setList] = useState<Expense[]>([])
    const [category, setCategory] = useState('食物')
    const [amount, setAmount] = useState<number | ''>('')
    const [note, setNote] = useState('')
    const [date, setDate] = useState<string>(new Date().toISOString().slice(0, 10))
    const [isListening, setIsListening] = useState(false)
    const today = new Date().toISOString().slice(0, 10)
    const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 3600 * 1000).toISOString().slice(0, 10)
    const [filterCategory, setFilterCategory] = useState('')
    const [filterFrom, setFilterFrom] = useState(thirtyDaysAgo)
    const [filterTo, setFilterTo] = useState(today)
    const [analysisQuery, setAnalysisQuery] = useState('')
    const [analysisResult, setAnalysisResult] = useState<string | null>(null)
    const [currentPage, setCurrentPage] = useState(1)
    const [shouldAutoApply, setShouldAutoApply] = useState(false)

    const getCategoryName = (cat: string) => CATEGORY_MAP[cat] || cat
    const getCategoryColor = (cat: string) => CATEGORY_COLORS[cat] || '#999'

    const autoSetFilters = (text: string) => {
        const lowerText = text.toLowerCase()

        if (lowerText.includes('食物') || lowerText.includes('吃') || lowerText.includes('饭') || lowerText.includes('餐')) {
            setFilterCategory('食物')
        } else if (lowerText.includes('交通') || lowerText.includes('打车') || lowerText.includes('地铁') || lowerText.includes('公交')) {
            setFilterCategory('交通')
        } else if (lowerText.includes('住宿') || lowerText.includes('酒店') || lowerText.includes('住')) {
            setFilterCategory('住宿')
        } else if (lowerText.includes('购物') || lowerText.includes('买')) {
            setFilterCategory('购物')
        } else if (lowerText.includes('活动') || lowerText.includes('娱乐') || lowerText.includes('玩')) {
            setFilterCategory('活动')
        }

        const today = new Date()
        if (lowerText.includes('一周') || lowerText.includes('7天') || lowerText.includes('七天') || lowerText.includes('最近一周')) {
            const weekAgo = new Date(today.getTime() - 7 * 24 * 3600 * 1000)
            setFilterFrom(weekAgo.toISOString().slice(0, 10))
            setFilterTo(today.toISOString().slice(0, 10))
        } else if (lowerText.includes('三天') || lowerText.includes('3天')) {
            const threeDaysAgo = new Date(today.getTime() - 3 * 24 * 3600 * 1000)
            setFilterFrom(threeDaysAgo.toISOString().slice(0, 10))
            setFilterTo(today.toISOString().slice(0, 10))
        } else if (lowerText.includes('一个月') || lowerText.includes('30天') || lowerText.includes('三十天') || lowerText.includes('最近一个月')) {
            const monthAgo = new Date(today.getTime() - 30 * 24 * 3600 * 1000)
            setFilterFrom(monthAgo.toISOString().slice(0, 10))
            setFilterTo(today.toISOString().slice(0, 10))
        } else if (lowerText.includes('今天') || lowerText.includes('今日')) {
            setFilterFrom(today.toISOString().slice(0, 10))
            setFilterTo(today.toISOString().slice(0, 10))
        } else if (lowerText.includes('昨天')) {
            const yesterday = new Date(today.getTime() - 24 * 3600 * 1000)
            setFilterFrom(yesterday.toISOString().slice(0, 10))
            setFilterTo(yesterday.toISOString().slice(0, 10))
        }
    }

    const { isListening: srListening, toggle, stop } = useSpeechRecognition({
        onInterim: () => { },
        onFinal: (t: string) => {
            setAnalysisQuery(t)
            autoSetFilters(t)
            setShouldAutoApply(true)
        },
    })

    useEffect(() => {
        setIsListening(srListening)
        if (srListening) {
            setAnalysisQuery('')
        }
    }, [srListening])

    useEffect(() => {
        if (shouldAutoApply) {
            setShouldAutoApply(false)
            setTimeout(() => {
                fetchList()
            }, 100)
        }
    }, [shouldAutoApply])

    const handleFilterFromChange = (newFrom: string) => {
        setFilterFrom(newFrom)
        if (filterTo && newFrom > filterTo) {
            setFilterTo(newFrom)
        }
    }

    const handleFilterToChange = (newTo: string) => {
        setFilterTo(newTo)
        if (filterFrom && newTo < filterFrom) {
            setFilterFrom(newTo)
        }
    }

    const addExpense = async () => {
        if (!amount || Number(amount) <= 0) return
        const noteValue = note.trim() || '消费'
        const rec = { category, amount: Number(amount), currency: 'CNY', note: noteValue, date }
        const data = await apiPost('/api/expenses', rec)
        if (data.success && data.data) {
            fetchList()
            setAmount('')
            setNote('')
        }
    }

    const fetchList = async () => {
        const params: Record<string, string | undefined> = {}
        if (filterCategory) params.category = filterCategory
        if (filterFrom) params.from = filterFrom
        if (filterTo) params.to = filterTo
        const data = await apiGet('/api/expenses', params)
        if (data.success) {
            const sorted = (data.data || []).sort((a: Expense, b: Expense) => {
                const dateA = new Date(a.createdAt || 0).getTime()
                const dateB = new Date(b.createdAt || 0).getTime()
                return dateB - dateA
            })
            setList(sorted)
            setCurrentPage(1)
        }
    }

    const analyze = async (useQuery = false) => {
        const params: Record<string, string | undefined> = {}
        if (filterCategory) params.category = filterCategory
        if (filterFrom) params.from = filterFrom
        if (filterTo) params.to = filterTo
        if (useQuery && analysisQuery) params.q = analysisQuery
        const data = await apiGet('/api/expenses/analyze', params)
        if (data.success) {
            setAnalysisResult(String(data.data.analysis))
            if (srListening) stop()
        }
    }

    useEffect(() => { fetchList() }, [])

    const totalsMap = new Map<string, number>()
    list.forEach(it => totalsMap.set(it.category, (totalsMap.get(it.category) || 0) + Number(it.amount || 0)))
    const totalsArr = Array.from(totalsMap.entries())
    const totalSum = totalsArr.reduce((s, [, v]) => s + v, 0) || 0

    const threshold = totalSum * 0.01
    const mainItems: [string, number][] = []
    let othersSum = 0
    totalsArr.forEach(([k, v]) => {
        if (v >= threshold) {
            mainItems.push([k, v])
        } else {
            othersSum += v
        }
    })
    if (othersSum > 0) {
        mainItems.push(['其他(小额)', othersSum])
    }
    const pieData = mainItems.length > 0 ? mainItems : totalsArr

    const totalPages = Math.ceil(list.length / BUDGET_ITEMS_PER_PAGE)
    const startIndex = (currentPage - 1) * BUDGET_ITEMS_PER_PAGE
    const endIndex = startIndex + BUDGET_ITEMS_PER_PAGE
    const currentItems = list.slice(startIndex, endIndex)

    return (
        <div className="trip-planner-page">
            <div className="trip-planner">
                <div className="planner-header">
                    <h2>💰 预算与开销管理</h2>
                    <p>记录并分析你的旅行支出</p>
                </div>

                <div className="planner-form budget-layout">
                    { }
                    <div className="top-row add-row">
                        <input value={note} onChange={e => setNote(e.target.value)} placeholder="消费事件" />
                        <select value={category} onChange={e => setCategory(e.target.value)}>
                            {CATEGORY_NAMES.map(cat => <option key={cat}>{cat}</option>)}
                        </select>
                        <input type="number" value={amount as any} onChange={e => setAmount(e.target.value === '' ? '' : Number(e.target.value))} placeholder="金额" />
                        <input type="date" value={date} onChange={e => setDate(e.target.value)} placeholder="日期" />
                        <button className="submit-button" onClick={addExpense}>添加</button>
                    </div>

                    { }
                    <div className="voice-row">
                        <input
                            className="analysis-input"
                            value={analysisQuery}
                            onChange={e => setAnalysisQuery(e.target.value)}
                            placeholder='例如"帮我看最近一周吃饭的花销"'
                        />
                        <div className="voice-buttons">
                            <button className={`voice-button ${isListening ? 'listening' : ''}`} onClick={() => toggle()}>
                                {isListening ? '⏹ 停止录音' : '🎤 开始语音'}
                            </button>
                            <button className="analyze-button" onClick={() => analyze(true)}>✨ 开始分析</button>
                        </div>
                    </div>

                    { }
                    <div className="budget-main-box">
                        <div className="left-section">
                            <div className="filter-section">
                                <h4>消费记录</h4>
                                <div className="filter-row">
                                    <select value={filterCategory} onChange={e => setFilterCategory(e.target.value)}>
                                        <option value="">全部种类</option>
                                        {CATEGORY_NAMES.map(cat => <option key={cat}>{cat}</option>)}
                                    </select>
                                    <input type="date" value={filterFrom} onChange={e => handleFilterFromChange(e.target.value)} />
                                    <input type="date" value={filterTo} onChange={e => handleFilterToChange(e.target.value)} />
                                    <button className="submit-button apply-btn" onClick={fetchList}>应用</button>
                                </div>
                            </div>

                            <div className="list-section">
                                <div className="expenses-list">
                                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                                        <thead>
                                            <tr style={{ textAlign: 'left', borderBottom: '1px solid #eee' }}>
                                                <th>消费事件</th>
                                                <th>种类</th>
                                                <th style={{ textAlign: 'right' }}>金额</th>
                                                <th>日期</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {currentItems.map((it, idx) => (
                                                <tr key={idx} style={{ borderBottom: '1px solid #fafafa' }}>
                                                    <td style={{ padding: '8px 6px' }}>{it.note}</td>
                                                    <td style={{ padding: '8px 6px' }}>{getCategoryName(it.category)}</td>
                                                    <td style={{ padding: '8px 6px', textAlign: 'right' }}>¥{formatAmount(it.amount)}</td>
                                                    <td style={{ padding: '8px 6px' }}>{it.createdAt ? it.createdAt.split('T')[0] : ''}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                                {totalPages > 1 && (
                                    <div className="pagination">
                                        <button
                                            className="page-btn"
                                            onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                                            disabled={currentPage === 1}
                                        >
                                            ← 上一页
                                        </button>
                                        <span className="page-info">第 {currentPage} / {totalPages} 页</span>
                                        <button
                                            className="page-btn"
                                            onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                                            disabled={currentPage === totalPages}
                                        >
                                            下一页 →
                                        </button>
                                    </div>
                                )}
                            </div>
                        </div>

                        <div className="right-section">
                            <svg viewBox="0 0 32 32" width="210" height="210" style={{ display: 'block', margin: '12px auto' }}>
                                {pieData.length === 0 ? (
                                    <circle cx="16" cy="16" r="16" fill="#e0e0e0" />
                                ) : pieData.length === 1 ? (
                                    <circle cx="16" cy="16" r="16" fill={getCategoryColor(pieData[0][0])} />
                                ) : (
                                    (() => {
                                        let acc = 0
                                        return pieData.map(([k, v]) => {
                                            const frac = v / (totalSum || 1)
                                            const start = acc
                                            const end = acc + frac
                                            acc = end
                                            const large = frac > 0.5 ? 1 : 0
                                            const a = (start - 0.25) * Math.PI * 2
                                            const b = (end - 0.25) * Math.PI * 2
                                            const x1 = 16 + 16 * Math.cos(a)
                                            const y1 = 16 + 16 * Math.sin(a)
                                            const x2 = 16 + 16 * Math.cos(b)
                                            const y2 = 16 + 16 * Math.sin(b)
                                            const d = `M16 16 L ${x1} ${y1} A 16 16 0 ${large} 1 ${x2} ${y2} Z`
                                            return <path key={k} d={d} fill={getCategoryColor(k)} stroke="#fff" strokeWidth="0.5" />
                                        })
                                    })()
                                )}
                            </svg>
                            {pieData.length > 0 ? (
                                <div className="legend">
                                    {pieData.map(([k, v]) => (
                                        <div key={k} className="legend-item">
                                            <span className="legend-swatch" style={{ background: getCategoryColor(k) }} />
                                            <span className="legend-label">{k === '其他(小额)' ? k : getCategoryName(k)}</span>
                                            <span className="legend-value">¥{formatAmount(v)}</span>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <p style={{ textAlign: 'center', color: '#999', fontSize: 13, marginTop: 12 }}>暂无消费记录</p>
                            )}
                        </div>
                    </div>

                    {analysisResult && (
                        <div className="analysis-box" style={{ marginTop: 12 }}>
                            <h4>AI 分析</h4>
                            <pre className="analysis-pre">{analysisResult}</pre>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
