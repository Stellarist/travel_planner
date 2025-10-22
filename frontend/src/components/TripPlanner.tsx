import { useState, useEffect } from 'react';
import { getApiUrl } from '../shared/config';
import '../shared/common.css';
import './TripPlanner.css';
import speechRecognition from '../shared/speechRecognition';
import type { TripPlan } from '../shared/types';
import { AVAILABLE_PREFERENCES } from '../shared/constants';

export default function TripPlanner() {
    const getDefaultDates = () => {
        const today = new Date();
        const threeDaysLater = new Date();
        threeDaysLater.setDate(today.getDate() + 3);

        return {
            start: today.toISOString().split('T')[0],
            end: threeDaysLater.toISOString().split('T')[0]
        };
    };

    const defaultDates = getDefaultDates();

    const [destination, setDestination] = useState('');
    const [startDate, setStartDate] = useState(defaultDates.start);
    const [endDate, setEndDate] = useState(defaultDates.end);
    const [budget, setBudget] = useState('2000');
    const [travelers, setTravelers] = useState('1');
    const [preferences, setPreferences] = useState<string[]>([]);
    const [specialNeeds, setSpecialNeeds] = useState('');
    const [isListening, setIsListening] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [tripPlan, setTripPlan] = useState<TripPlan | null>(null);
    const [error, setError] = useState('');
    const [recognizedText, setRecognizedText] = useState('');

    const { isListening: srListening, recognizedText: srText, toggle } = speechRecognition({
        onFinal: (t: string) => { setRecognizedText(t); parseVoiceInput(t); },
    });

    useEffect(() => {
        setIsListening(srListening);
    }, [srListening]);

    useEffect(() => {
        setRecognizedText(srText);
    }, [srText]);

    const parseVoiceInput = (text: string) => {
        console.log('语音输入:', text);

        const chineseToNumber = (cn: string): number | null => {
            if (!cn) return null;
            const map: any = { 零: 0, 一: 1, 二: 2, 三: 3, 四: 4, 五: 5, 六: 6, 七: 7, 八: 8, 九: 9 };
            let result = 0;
            let section = 0;
            let number = 0;
            for (let i = 0; i < cn.length; i++) {
                const ch = cn[i];
                if (ch === '万') {
                    section = (section + number) * 10000;
                    result += section;
                    section = 0;
                    number = 0;
                } else if (ch === '千') {
                    section += (number || 1) * 1000;
                    number = 0;
                } else if (ch === '百') {
                    section += (number || 1) * 100;
                    number = 0;
                } else if (ch === '十') {
                    section += (number || 1) * 10;
                    number = 0;
                } else if (map.hasOwnProperty(ch)) {
                    number = map[ch];
                } else {
                    return null;
                }
            }
            return result + section + number;
        };

        const destRegex = /(去|到|想去|我要去|想去到)\s*([A-Za-z0-9\u4e00-\u9fa5·\s]{1,30}?)(?=\s|\d|天|日|预算|带|和|，|,|。|$)/i;
        const destMatch = text.match(destRegex);
        if (destMatch && destMatch[2]) {
            const dest = destMatch[2].trim().replace(/[，,。\.\s]+$/, '');
            if (dest) setDestination(dest);
        } else {
            const lonePlace = text.match(/^[\s]*(?:我想去|我要去|去|到)?\s*([A-Za-z0-9\u4e00-\u9fa5·]{2,30})[\s,，。]?/i);
            if (lonePlace && lonePlace[1]) {
                setDestination(lonePlace[1].trim());
            }
        }

        const daysRegex = /(\d+)\s*天|([一二三四五六七八九十百零]+)\s*天/;
        const daysMatch = text.match(daysRegex);
        if (daysMatch && startDate) {
            let days = 0;
            if (daysMatch[1]) days = parseInt(daysMatch[1]);
            else if (daysMatch[2]) {
                const n = chineseToNumber(daysMatch[2]);
                if (n) days = n;
            }
            if (days > 0) {
                const start = new Date(startDate);
                const end = new Date(start);
                end.setDate(end.getDate() + days - 1);
                setEndDate(end.toISOString().split('T')[0]);
            }
        }

        let budgetValue: number | null = null;
        const budRegex1 = /预算\s*([0-9]+(?:\.[0-9]+)?)(?:\s*(万|万元|元))?/i;
        const budRegex2 = /([0-9]+(?:\.[0-9]+)?)\s*(万|万元|元)?\s*预算/i;
        const budCNA = /预算\s*([一二三四五六七八九十百千万零]+)\s*(万|万元|元)?/;

        let m = text.match(budRegex1);
        if (!m) m = text.match(budRegex2);
        if (m && m[1]) {
            const num = parseFloat(m[1]);
            const unit = m[2] || '';
            if (unit && unit.includes('万')) {
                budgetValue = num * 10000;
            } else {
                budgetValue = num;
            }
        } else {
            const m2 = text.match(budCNA);
            if (m2 && m2[1]) {
                const cnNum = chineseToNumber(m2[1]);
                if (cnNum != null) {
                    const unit = m2[2] || '';
                    if (unit && unit.includes('万')) budgetValue = cnNum * 10000;
                    else budgetValue = cnNum;
                }
            }
        }

        if (budgetValue != null) {
            setBudget(budgetValue.toString());
        }

        const newPrefs: string[] = [];
        AVAILABLE_PREFERENCES.forEach(pref => {
            if (text.includes(pref)) {
                newPrefs.push(pref);
            }
        });
        if (newPrefs.length > 0) {
            setPreferences(prev => [...new Set([...prev, ...newPrefs])]);
        }

        if (text.includes('带孩子') || text.includes('亲子')) {
            setSpecialNeeds(prev => prev ? prev + '、带孩子' : '带孩子');
        }
    };

    const togglePreference = (pref: string) => {
        if (preferences.includes(pref)) {
            setPreferences(preferences.filter(p => p !== pref));
        } else {
            setPreferences([...preferences, pref]);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            const token = localStorage.getItem('token');
            if (!token) {
                setError('请先登录');
                setIsLoading(false);
                return;
            }

            const response = await fetch(getApiUrl('/api/trips/plan'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify({
                    destination,
                    startDate,
                    endDate,
                    budget: parseFloat(budget),
                    travelers: parseInt(travelers),
                    preferences,
                    specialNeeds,
                }),
            });

            const data = await response.json();

            if (data.success && data.trip) {
                setTripPlan(data.trip);
            } else {
                setError(data.message || '生成行程失败');
            }
        } catch (err) {
            setError('网络错误，请稍后重试');
        } finally {
            setIsLoading(false);
        }
    };

    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        const year = date.getFullYear();
        const month = date.getMonth() + 1;
        const day = date.getDate();
        return `${year}-${month}-${day}`;
    };

    const resetForm = () => {
        const newDefaultDates = getDefaultDates();
        setDestination('');
        setStartDate(newDefaultDates.start);
        setEndDate(newDefaultDates.end);
        setBudget('2000');
        setTravelers('1');
        setPreferences([]);
        setSpecialNeeds('');
        setTripPlan(null);
        setError('');
    };

    return (
        <div className="trip-planner-page">
            <div className="trip-planner">
                <div className="planner-header">
                    <h2>🗺️ 智能行程规划</h2>
                    <p>告诉我你的旅行想法，让 AI 为你定制专属行程</p>
                </div>

                {!tripPlan ? (
                    <form className="planner-form" onSubmit={handleSubmit}>
                        <div className="voice-input-section">
                            <button
                                type="button"
                                className={`voice-button ${isListening ? 'listening' : ''}`}
                                onClick={() => toggle()}
                            >
                                {isListening ? '⏹ 停止聆听' : '🎤 语音输入'}
                            </button>
                            <p className="voice-hint">
                                例如："我想去日本，5 天，预算 1 万元，喜欢美食和动漫，带孩子"
                            </p>
                            {recognizedText && (
                                <div className="recognized-text">
                                    <strong>识别结果：</strong>
                                    <span>{recognizedText}</span>
                                </div>
                            )}

                            { }
                        </div>

                        <div className="form-group">
                            <label>目的地 *</label>
                            <input
                                type="text"
                                value={destination}
                                onChange={(e) => setDestination(e.target.value)}
                                placeholder="例如：日本东京"
                                required
                            />
                        </div>

                        <div className="form-row">
                            <div className="form-group">
                                <label>出发日期 *</label>
                                <input
                                    type="date"
                                    value={startDate}
                                    onChange={(e) => setStartDate(e.target.value)}
                                    required
                                />
                            </div>
                            <div className="form-group">
                                <label>返程日期 *</label>
                                <input
                                    type="date"
                                    value={endDate}
                                    onChange={(e) => setEndDate(e.target.value)}
                                    min={startDate}
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="form-group">
                                <label>预算 (元) *</label>
                                <input
                                    type="number"
                                    value={budget}
                                    onChange={(e) => setBudget(e.target.value)}
                                    placeholder="例如：10000"
                                    min="0"
                                    required
                                />
                            </div>
                            <div className="form-group">
                                <label>同行人数 *</label>
                                <input
                                    type="number"
                                    value={travelers}
                                    onChange={(e) => setTravelers(e.target.value)}
                                    min="1"
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="form-group">
                                <label>旅行偏好</label>
                                <div className="preferences-grid">
                                    {AVAILABLE_PREFERENCES.map(pref => (
                                        <button
                                            key={pref}
                                            type="button"
                                            className={`pref-tag ${preferences.includes(pref) ? 'active' : ''}`}
                                            onClick={() => togglePreference(pref)}
                                        >
                                            {pref}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div className="form-group">
                                <label>特殊需求</label>
                                <textarea
                                    className="special-needs"
                                    value={specialNeeds}
                                    onChange={(e) => setSpecialNeeds(e.target.value)}
                                    placeholder="例如：带孩子、需要无障碍设施、素食等"
                                    rows={1}
                                />
                            </div>
                        </div>

                        {error && <div className="error-message">{error}</div>}

                        <button type="submit" className="submit-button" disabled={isLoading}>
                            {isLoading ? '⏳ 正在生成行程...' : '✨ 生成行程'}
                        </button>
                    </form>
                ) : (
                    <div className="trip-result">
                        <div className="result-header">
                            <h3>📋 您的{tripPlan.destination}行程</h3>
                            <button className="new-plan-button" onClick={resetForm}>
                                + 创建新行程
                            </button>
                        </div>

                        <div className="trip-summary">
                            <div className="summary-item">
                                <span className="label">出行日期：</span>
                                <span>{formatDate(tripPlan.startDate)} 至 {formatDate(tripPlan.endDate)}</span>
                            </div>
                            <div className="summary-item">
                                <span className="label">总费用：</span>
                                <span className="price">¥{tripPlan.totalCost.toFixed(2)}</span>
                            </div>
                            <div className="summary-text">
                                {tripPlan.summary}
                            </div>
                        </div>

                        <div className="itinerary-container">
                            {tripPlan.itinerary.map((day, index) => (
                                <div key={index} className="day-card">
                                    <div className="day-header">
                                        <h4>第 {day.day} 天</h4>
                                        <span className="day-date">{formatDate(day.date)}</span>
                                        <span className="day-cost">¥{day.dailyCost.toFixed(2)}</span>
                                    </div>

                                    <div className="activities-list">
                                        {day.activities.map((activity, actIdx) => (
                                            <div key={actIdx} className="activity-item">
                                                <div className="activity-time">{activity.time}</div>
                                                <div className="activity-content">
                                                    <div className="activity-header">
                                                        <span className="activity-type">{activity.type}</span>
                                                        <h5>{activity.name}</h5>
                                                    </div>
                                                    <p className="activity-location">📍 {activity.location}</p>
                                                    <p className="activity-desc">{activity.description}</p>
                                                    {activity.tips && (
                                                        <p className="activity-tips">💡 {activity.tips}</p>
                                                    )}
                                                    <div className="activity-footer">
                                                        <span className="duration">⏱️ {activity.duration}</span>
                                                        <span className="cost">💰 ¥{activity.cost.toFixed(2)}</span>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>

                                    {day.accommodation && (
                                        <div className="accommodation">
                                            <span>🏨 住宿：{day.accommodation}</span>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
