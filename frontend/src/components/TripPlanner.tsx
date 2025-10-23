import { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { getApiUrl, useSpeechRecognition, getDefaultDateRange, formatDate, apiPost } from '../shared/utils';
import '../styles/common.css';
import './TripPlanner.css';
import type { TripPlan, ParsedTripInfo } from '../shared/types';
import { AVAILABLE_PREFERENCES } from '../shared/constants';

export default function TripPlanner() {
    const location = useLocation();
    const navigate = useNavigate();
    const defaultDates = getDefaultDateRange(3)

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
    const [parseLoading, setParseLoading] = useState(false);
    const [isFavorited, setIsFavorited] = useState(false);
    const [showFavorites, setShowFavorites] = useState(false);
    const [favoriteTrips, setFavoriteTrips] = useState<TripPlan[]>([]);

    // 从收藏夹进入时加载行程数据
    useEffect(() => {
        const state = location.state as { tripPlan?: TripPlan; isFavorited?: boolean };
        if (state?.tripPlan) {
            setTripPlan(state.tripPlan);
            setIsFavorited(state.isFavorited || false);
            // 填充目的地和日期以便用户查看
            setDestination(state.tripPlan.destination);
            setStartDate(state.tripPlan.startDate);
            setEndDate(state.tripPlan.endDate);
        }
    }, [location]);

    // 加载收藏的行程
    const loadFavoriteTrips = async () => {
        const token = localStorage.getItem('token');
        if (!token) return;

        try {
            const response = await fetch(getApiUrl('/api/trips/favorites/list'), {
                headers: { 'Authorization': `Bearer ${token}` },
            });
            const data = await response.json();
            if (data.success) {
                setFavoriteTrips(data.data || []);
            }
        } catch (err) {
            console.error('加载收藏失败:', err);
        }
    };

    useEffect(() => {
        loadFavoriteTrips();
    }, []);

    const { isListening: srListening, recognizedText: srText, toggle } = useSpeechRecognition({
        onFinal: (t: string) => {
            setRecognizedText(t);
            parseVoiceInputWithBackend(t);
        },
    });

    useEffect(() => setIsListening(srListening), [srListening]);
    useEffect(() => setRecognizedText(srText), [srText]);

    const parseVoiceInputWithBackend = async (text: string) => {
        console.log('语音输入:', text);

        try {
            setParseLoading(true);
            const response: any = await apiPost('/api/parser/parse', { text });

            if (response.success && response.data) {
                const parsed: ParsedTripInfo = response.data;
                console.log('后端解析结果:', parsed);

                setDestination('');
                setStartDate(defaultDates.start);
                setEndDate(defaultDates.end);
                setBudget('');
                setTravelers('1');
                setPreferences([]);
                setSpecialNeeds('');

                if (parsed.destination) {
                    setDestination(parsed.destination);
                }
                if (parsed.startDate) {
                    setStartDate(parsed.startDate);
                }
                if (parsed.endDate) {
                    setEndDate(parsed.endDate);
                } else if (parsed.startDate && parsed.duration > 0) {
                    const start = new Date(parsed.startDate);
                    const end = new Date(start);
                    end.setDate(end.getDate() + parsed.duration - 1);
                    setEndDate(end.toISOString().split('T')[0]);
                }
                if (parsed.budget > 0) {
                    setBudget(parsed.budget.toString());
                }
                if (parsed.travelers > 0) {
                    setTravelers(parsed.travelers.toString());
                }
                if (parsed.preferences && parsed.preferences.length > 0) {
                    setPreferences(parsed.preferences);
                }

                if (parsed.confidence === 'low') {
                    console.warn('解析置信度较低，请检查填充的信息');
                }
            } else {
                console.warn('后端解析失败');
            }
        } catch (error) {
            console.error('调用后端解析API失败:', error);
        } finally {
            setParseLoading(false);
        }
    };

    const togglePreference = (pref: string) => {
        setPreferences(prev =>
            prev.includes(pref) ? prev.filter(p => p !== pref) : [...prev, pref]
        );
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        const token = localStorage.getItem('token');
        if (!token) {
            setError('请先登录');
            return;
        }

        setIsLoading(true);
        try {
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
                setIsFavorited(false); // 新生成的行程默认未收藏
            } else {
                setError(data.message || '生成行程失败');
            }
        } catch (err) {
            console.error('生成行程失败:', err);
            setError('网络错误或服务器响应超时，请稍后重试。AI 生成行程可能需要 1-2 分钟，请耐心等待。');
        } finally {
            setIsLoading(false);
        }
    };

    const resetForm = () => {
        const newDefaultDates = getDefaultDateRange(3);
        setDestination('');
        setStartDate(newDefaultDates.start);
        setEndDate(newDefaultDates.end);
        setBudget('2000');
        setTravelers('1');
        setPreferences([]);
        setSpecialNeeds('');
        setTripPlan(null);
        setError('');
        setIsFavorited(false);
    };

    // 收藏/取消收藏行程
    const toggleFavorite = async () => {
        if (!tripPlan) return;

        const token = localStorage.getItem('token');
        if (!token) {
            setError('请先登录');
            return;
        }

        try {
            const url = isFavorited
                ? getApiUrl(`/api/trips/favorites/${tripPlan.id}`)
                : getApiUrl(`/api/trips/favorites/${tripPlan.id}`);

            const response = await fetch(url, {
                method: isFavorited ? 'DELETE' : 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            const data = await response.json();

            if (data.success) {
                setIsFavorited(!isFavorited);
                // 刷新收藏列表
                loadFavoriteTrips();
            } else {
                setError(data.message || '操作失败');
            }
        } catch (err) {
            console.error('收藏操作失败:', err);
            setError('操作失败，请稍后重试');
        }
    };

    // 从收藏夹中移除行程
    const removeFavoriteTrip = async (tripId: string) => {
        const token = localStorage.getItem('token');
        if (!token) return;

        try {
            const response = await fetch(getApiUrl(`/api/trips/favorites/${tripId}`), {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` },
            });

            const data = await response.json();
            if (data.success) {
                setFavoriteTrips(prev => prev.filter(trip => trip.id !== tripId));
                // 如果删除的是当前显示的行程，更新收藏状态
                if (tripPlan?.id === tripId) {
                    setIsFavorited(false);
                }
            }
        } catch (err) {
            console.error('删除收藏失败:', err);
        }
    };

    // 查看收藏的行程
    const viewFavoriteTrip = (trip: TripPlan) => {
        setTripPlan(trip);
        setIsFavorited(true);
        setDestination(trip.destination);
        setStartDate(trip.startDate);
        setEndDate(trip.endDate);
        setShowFavorites(false);
    };

    return (
        <div className="trip-planner-page">
            <div className="trip-planner">
                <div className="planner-header">
                    <button className="back-home-button" onClick={() => navigate('/dashboard')}>
                        ← 返回主页
                    </button>
                    <h2>🗺️ 智能行程规划</h2>
                    <p>告诉我你的旅行想法，让 AI 为你定制专属行程</p>
                    <button
                        className="favorites-toggle-btn"
                        onClick={() => setShowFavorites(!showFavorites)}
                        title={showFavorites ? '隐藏收藏夹' : '显示收藏夹'}
                    >
                        ⭐ 我的收藏 ({favoriteTrips.length})
                    </button>
                </div>

                {/* 收藏夹面板 */}
                {showFavorites && (
                    <div className="favorites-panel">
                        <div className="favorites-panel-header">
                            <h3>⭐ 收藏的行程</h3>
                            <button
                                className="close-favorites-btn"
                                onClick={() => setShowFavorites(false)}
                            >
                                ✕
                            </button>
                        </div>
                        <div className="favorites-list">
                            {favoriteTrips.length === 0 ? (
                                <div className="empty-favorites">
                                    <p>还没有收藏的行程</p>
                                </div>
                            ) : (
                                favoriteTrips.map(trip => (
                                    <div key={trip.id} className="favorite-trip-item">
                                        <div className="favorite-trip-info" onClick={() => viewFavoriteTrip(trip)}>
                                            <h4>{trip.destination}</h4>
                                            <div className="trip-meta">
                                                <span>📅 {formatDate(trip.startDate)} - {formatDate(trip.endDate)}</span>
                                                <span>💰 ¥{trip.totalCost.toFixed(0)}</span>
                                            </div>
                                            <p className="trip-summary-short">{trip.summary}</p>
                                        </div>
                                        <button
                                            className="remove-favorite-btn"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                if (confirm('确定取消收藏此行程？')) {
                                                    removeFavoriteTrip(trip.id);
                                                }
                                            }}
                                            title="取消收藏"
                                        >
                                            ✕
                                        </button>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                )}

                {!tripPlan ? (
                    <form className="planner-form" onSubmit={handleSubmit}>
                        <div className="voice-input-section">
                            <div className="voice-controls">
                                <button
                                    type="button"
                                    className={`voice-button ${isListening ? 'listening' : ''}`}
                                    onClick={() => toggle()}
                                >
                                    {isListening ? '⏹ 停止聆听' : '🎤 语音输入'}
                                </button>
                                {!recognizedText && (
                                    <p className="voice-hint">
                                        例如："我想去上海，5 天，预算 1 万元，喜欢美食和动漫，带孩子"
                                    </p>
                                )}
                            </div>
                            {recognizedText && (
                                <div className="recognized-text-editor">
                                    <div className="text-editor-row">
                                        <input
                                            className="recognized-input"
                                            type="text"
                                            value={recognizedText}
                                            onChange={(e) => setRecognizedText(e.target.value)}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter') {
                                                    e.preventDefault();
                                                    if (recognizedText.trim() && !parseLoading) {
                                                        parseVoiceInputWithBackend(recognizedText);
                                                    }
                                                }
                                            }}
                                            placeholder="语音识别的文字会显示在这里，你可以修改后再解析"
                                        />
                                        <button
                                            type="button"
                                            className="parse-button"
                                            onClick={() => parseVoiceInputWithBackend(recognizedText)}
                                            disabled={!recognizedText.trim() || parseLoading}
                                        >
                                            {parseLoading ? '解析中…' : '✨ 智能解析'}
                                        </button>
                                    </div>
                                    { }
                                </div>
                            )}
                        </div>

                        <div className="form-group">
                            <label>目的地 *</label>
                            <input
                                type="text"
                                value={destination}
                                onChange={(e) => setDestination(e.target.value)}
                                placeholder="例如：上海"
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
                            {isLoading ? '⏳ AI 正在规划行程，请稍候...' : '✨ 生成行程'}
                        </button>
                    </form>
                ) : (
                    <div className="trip-result">
                        <div className="result-header">
                            <h3>📋 您的{tripPlan.destination}行程</h3>
                            <div style={{ display: 'flex', gap: '10px' }}>
                                <button
                                    className={`favorite-trip-button ${isFavorited ? 'favorited' : ''}`}
                                    onClick={toggleFavorite}
                                    title={isFavorited ? '取消收藏' : '收藏行程'}
                                >
                                    {isFavorited ? '★ 已收藏' : '☆ 收藏'}
                                </button>
                                <button className="new-plan-button" onClick={resetForm}>
                                    + 创建新行程
                                </button>
                            </div>
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
