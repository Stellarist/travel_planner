import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import '../styles/common.css'
import './TripPlanner.css'
import './BudgetManager.css'
import { apiPost, apiGet, formatAmount, useSpeechRecognition, getTodayDate, getDateDaysAgo } from '../shared/utils'
import type { Expense, ParsedExpenseQuery } from '../shared/types'
import { CATEGORY_MAP, CATEGORY_COLORS, BUDGET_ITEMS_PER_PAGE, CATEGORY_NAMES } from '../shared/constants'

export default function BudgetManager() {
    const navigate = useNavigate()
    const [list, setList] = useState<Expense[]>([])
    const [category, setCategory] = useState('食物')
    const [amount, setAmount] = useState<number | ''>('')
    const [note, setNote] = useState('')
    const [date, setDate] = useState<string>(getTodayDate())
    const [isListening, setIsListening] = useState(false)
    const [filterCategory, setFilterCategory] = useState('')
    const [filterFrom, setFilterFrom] = useState(getDateDaysAgo(30))
    const [filterTo, setFilterTo] = useState(getTodayDate())
    const [analysisQuery, setAnalysisQuery] = useState('')
    const [currentPage, setCurrentPage] = useState(1)
    const [shouldAutoApply, setShouldAutoApply] = useState(false)
    const [isAnalyzing, setIsAnalyzing] = useState(false)

    const getCategoryName = (cat: string) => CATEGORY_MAP[cat] || cat
    const getCategoryColor = (cat: string) => CATEGORY_COLORS[cat] || '#999'

    const parseExpenseQueryWithBackend = async (text: string) => {
        console.log('开销查询语音输入:', text);

        try {
            const response: any = await apiPost('/api/parser/parse-expense', { text });

            if (response.success && response.data) {
                const parsed: ParsedExpenseQuery = response.data;
                console.log('后端解析结果:', parsed);

                setFilterCategory('');
                setFilterFrom(getDateDaysAgo(30));
                setFilterTo(getTodayDate());
                setAnalysisQuery('');

                if (parsed.category) {
                    setFilterCategory(parsed.category);
                }
                if (parsed.startDate) {
                    setFilterFrom(parsed.startDate);
                }
                if (parsed.endDate) {
                    setFilterTo(parsed.endDate);
                }
                setAnalysisQuery(parsed.query || text);

                setShouldAutoApply(true);

                if (parsed.confidence === 'low') {
                    console.warn('解析置信度较低，请检查过滤条件');
                }
            } else {
                console.warn('后端解析失败');
            }
        } catch (error) {
            console.error('调用后端解析API失败:', error);
        }
    };

    const { isListening: srListening, toggle, stop } = useSpeechRecognition({
        onFinal: (t: string) => {
            setAnalysisQuery(t)
            parseExpenseQueryWithBackend(t)
        },
    })

    useEffect(() => {
        setIsListening(srListening)
        if (srListening) setAnalysisQuery('')
    }, [srListening])

    useEffect(() => {
        if (shouldAutoApply) {
            setShouldAutoApply(false)
            setTimeout(fetchList, 100)
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
        const rec = {
            category,
            amount: Number(amount),
            currency: 'CNY',
            note: note.trim() || '消费',
            date
        }
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
        // 如果有查询文本，先解析它
        if (useQuery && analysisQuery) {
            await parseExpenseQueryWithBackend(analysisQuery)
            // 等待解析完成后再继续分析
            await new Promise(resolve => setTimeout(resolve, 500))
        }

        setIsAnalyzing(true)
        try {
            const body: Record<string, any> = {}
            if (filterCategory) body.category = filterCategory
            if (filterFrom) body.from = filterFrom
            if (filterTo) body.to = filterTo
            if (useQuery && analysisQuery) body.q = analysisQuery

            const data = await apiPost('/api/expenses/analyze', body)
            if (data.success) {
                const analysis = String(data.data.analysis)
                // 导航到分析结果页面
                navigate('/budget/analysis', {
                    state: {
                        analysis: {
                            analysis,
                            query: useQuery ? analysisQuery : undefined,
                            category: filterCategory,
                            from: filterFrom,
                            to: filterTo
                        }
                    }
                })
                if (srListening) stop()
            }
        } catch (error) {
            console.error('分析失败:', error)
            alert('分析失败，请稍后重试')
        } finally {
            setIsAnalyzing(false)
        }
    }

    useEffect(() => { fetchList() }, [])

    const totalsMap = new Map<string, number>()
    list.forEach(it => totalsMap.set(it.category, (totalsMap.get(it.category) || 0) + Number(it.amount || 0)))
    const totalSum = Array.from(totalsMap.values()).reduce((s, v) => s + v, 0) || 0

    const threshold = totalSum * 0.01
    const tempData = Array.from(totalsMap.entries())
        .reduce((acc, [k, v]) => {
            if (v >= threshold) {
                acc.main.push([k, v])
            } else {
                acc.othersSum += v
            }
            return acc
        }, { main: [] as [string, number][], othersSum: 0 })

    if (tempData.othersSum > 0) tempData.main.push(['其他(小额)', tempData.othersSum])
    const pieData = tempData.main.length > 0 ? tempData.main : Array.from(totalsMap.entries())

    const totalPages = Math.ceil(list.length / BUDGET_ITEMS_PER_PAGE)
    const currentItems = list.slice(
        (currentPage - 1) * BUDGET_ITEMS_PER_PAGE,
        currentPage * BUDGET_ITEMS_PER_PAGE
    )

    return (
        <div className="trip-planner-page">
            <div className="trip-planner">
                <div className="planner-header">
                    <button className="back-home-button" onClick={() => navigate('/dashboard')}>
                        ← 返回主页
                    </button>
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
                            <button
                                className="analyze-button"
                                onClick={() => analyze(true)}
                                disabled={isAnalyzing}
                            >
                                {isAnalyzing ? '⏳ AI 正在分析...' : '✨ 开始分析'}
                            </button>
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
                </div>
            </div>
        </div>
    )
}
