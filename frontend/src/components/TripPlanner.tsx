import { useState } from 'react';
import { getApiUrl } from '../config';
import '../styles/common.css';
import './TripPlanner.css';

interface Activity {
    time: string;
    type: string;
    name: string;
    location: string;
    duration: string;
    cost: number;
    description: string;
    tips: string;
}

interface DayItinerary {
    day: number;
    date: string;
    activities: Activity[];
    accommodation: string;
    dailyCost: number;
}

interface TripPlan {
    id: string;
    destination: string;
    startDate: string;
    endDate: string;
    itinerary: DayItinerary[];
    totalCost: number;
    summary: string;
    createdAt: string;
}

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
    const [budget, setBudget] = useState('');
    const [travelers, setTravelers] = useState('1');
    const [preferences, setPreferences] = useState<string[]>([]);
    const [specialNeeds, setSpecialNeeds] = useState('');
    const [isListening, setIsListening] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [tripPlan, setTripPlan] = useState<TripPlan | null>(null);
    const [error, setError] = useState('');

    const availablePreferences = ['美食', '动漫', '亲子', '历史', '自然', '购物', '冒险'];

    // 语音识别
    const startVoiceInput = () => {
        if (!('webkitSpeechRecognition' in window) && !('SpeechRecognition' in window)) {
            setError('您的浏览器不支持语音识别功能，请使用 Chrome 浏览器');
            return;
        }

        const SpeechRecognition = (window as any).webkitSpeechRecognition || (window as any).SpeechRecognition;
        const recognition = new SpeechRecognition();
        recognition.lang = 'zh-CN';
        recognition.continuous = false;
        recognition.interimResults = false;

        recognition.onstart = () => {
            setIsListening(true);
            setError('');
        };

        recognition.onresult = (event: any) => {
            const transcript = event.results[0][0].transcript;
            parseVoiceInput(transcript);
        };

        recognition.onerror = (event: any) => {
            setIsListening(false);
            setError('语音识别失败: ' + event.error);
        };

        recognition.onend = () => {
            setIsListening(false);
        };

        recognition.start();
    };

    // 解析语音输入
    const parseVoiceInput = (text: string) => {
        console.log('语音输入:', text);

        // 提取目的地
        const destMatch = text.match(/去([\u4e00-\u9fa5]+)/);
        if (destMatch) {
            setDestination(destMatch[1]);
        }

        // 提取天数
        const daysMatch = text.match(/(\d+)\s*天/);
        if (daysMatch && startDate) {
            const days = parseInt(daysMatch[1]);
            const start = new Date(startDate);
            const end = new Date(start);
            end.setDate(end.getDate() + days - 1);
            setEndDate(end.toISOString().split('T')[0]);
        }

        // 提取预算
        const budgetMatch = text.match(/预算\s*(\d+\.?\d*)\s*[万元]/);
        if (budgetMatch) {
            const amount = parseFloat(budgetMatch[1]);
            if (text.includes('万')) {
                setBudget((amount * 10000).toString());
            } else {
                setBudget(amount.toString());
            }
        }

        // 提取偏好
        const newPrefs: string[] = [];
        availablePreferences.forEach(pref => {
            if (text.includes(pref)) {
                newPrefs.push(pref);
            }
        });
        if (newPrefs.length > 0) {
            setPreferences(prev => [...new Set([...prev, ...newPrefs])]);
        }

        // 提取特殊需求
        if (text.includes('带孩子') || text.includes('亲子')) {
            setSpecialNeeds(prev => prev ? prev + '、带孩子' : '带孩子');
        }
    };

    // 切换偏好
    const togglePreference = (pref: string) => {
        if (preferences.includes(pref)) {
            setPreferences(preferences.filter(p => p !== pref));
        } else {
            setPreferences([...preferences, pref]);
        }
    };

    // 提交表单
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

    // 格式化日期为 Y-M-D 格式
    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        const year = date.getFullYear();
        const month = date.getMonth() + 1;
        const day = date.getDate();
        return `${year}-${month}-${day}`;
    };

    // 重置表单
    const resetForm = () => {
        const newDefaultDates = getDefaultDates();
        setDestination('');
        setStartDate(newDefaultDates.start);
        setEndDate(newDefaultDates.end);
        setBudget('');
        setTravelers('1');
        setPreferences([]);
        setSpecialNeeds('');
        setTripPlan(null);
        setError('');
    };

    return (
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
                            onClick={startVoiceInput}
                            disabled={isListening}
                        >
                            {isListening ? '🎤 正在聆听...' : '🎤 语音输入'}
                        </button>
                        <p className="voice-hint">
                            例如："我想去日本，5 天，预算 1 万元，喜欢美食和动漫，带孩子"
                        </p>
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

                    <div className="form-group">
                        <label>旅行偏好</label>
                        <div className="preferences-grid">
                            {availablePreferences.map(pref => (
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
                            value={specialNeeds}
                            onChange={(e) => setSpecialNeeds(e.target.value)}
                            placeholder="例如：带孩子、需要无障碍设施、素食等"
                            rows={3}
                        />
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
    );
}
