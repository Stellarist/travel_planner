import { useState, useEffect } from 'react'
import speechRecognition from '../shared/speechRecognition'
import '../shared/common.css'
import './TripPlanner.css'
import './BudgetManager.css'
import { getApiUrl } from '../config'

interface Expense {
    id?: string
    category: string
    amount: number
    currency?: string
    note?: string
    createdAt?: string
}

export default function BudgetManager() {
    const [list, setList] = useState<Expense[]>([])
    const [category, setCategory] = useState('Meals')
    const [amount, setAmount] = useState<number | ''>('')
    const [note, setNote] = useState('')
    const [isListening, setIsListening] = useState(false)

    const { isListening: srListening, toggle } = speechRecognition({
        onInterim: () => { },
        onFinal: (t) => setNote(prev => prev ? prev + ' ' + t : t),
    })

    useEffect(() => { setIsListening(srListening) }, [srListening])

    const addExpense = async () => {
        if (!amount || Number(amount) <= 0) return
        const rec = { category, amount: Number(amount), currency: 'CNY', note }
        const token = localStorage.getItem('token')
        const res = await fetch(getApiUrl('/api/expenses'), {
            method: 'POST', headers: {
                'Content-Type': 'application/json', 'Authorization': `Bearer ${token}`
            }, body: JSON.stringify(rec)
        })
        const data = await res.json()
        if (data.success && data.data) {
            setList(prev => [data.data, ...prev])
            setAmount('')
            setNote('')
        }
    }

    const fetchList = async () => {
        const token = localStorage.getItem('token')
        const res = await fetch(getApiUrl('/api/expenses'), { headers: { 'Authorization': `Bearer ${token}` } })
        const data = await res.json()
        if (data.success) setList(data.data || [])
    }

    const analyze = async () => {
        const token = localStorage.getItem('token')
        const res = await fetch(getApiUrl('/api/expenses/analyze'), { headers: { 'Authorization': `Bearer ${token}` } })
        const data = await res.json()
        if (data.success) alert('AI analysis:\n' + JSON.stringify(data.data.analysis).slice(0, 1000))
    }

    useEffect(() => { fetchList() }, [])

    return (
        <div className="trip-planner-page">
            <div className="trip-planner">
                <div className="planner-header">
                    <h2>💰 预算与开销管理</h2>
                    <p>记录并分析你的旅行支出（支持语音输入）</p>
                </div>

                <div className="planner-form">
                    <div className="form-row">
                        <div className="form-group">
                            <label>Category</label>
                            <input value={category} onChange={e => setCategory(e.target.value)} />
                        </div>
                        <div className="form-group">
                            <label>Amount</label>
                            <input type="number" value={amount as any} onChange={e => setAmount(e.target.value === '' ? '' : Number(e.target.value))} />
                        </div>
                    </div>

                    <div className="form-group">
                        <label>Note (you can use voice)</label>
                        <div className="voice-row">
                            <textarea className="special-needs" value={note} onChange={e => setNote(e.target.value)} rows={1}></textarea>
                            <button type="button" className={`voice-button ${isListening ? 'listening' : ''}`} onClick={() => toggle()}>{isListening ? 'Stop' : 'Voice'}</button>
                        </div>
                    </div>

                    <div className="form-row">
                        <button className="submit-button" onClick={addExpense}>Add</button>
                        <button className="submit-button" onClick={analyze}>Analyze</button>
                    </div>

                    <div className="expenses-list">
                        {list.map((it, idx) => (
                            <div key={idx} className="day-card" style={{ padding: '10px' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <div style={{ fontWeight: 600 }}>{it.category}</div>
                                    <div style={{ color: '#ff4757' }}>¥{it.amount}</div>
                                </div>
                                {it.note && <div style={{ marginTop: 6, color: '#666', fontSize: 13 }}>{it.note}</div>}
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
}
