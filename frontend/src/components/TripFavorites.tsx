import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getApiUrl, formatDate } from '../shared/utils';
import type { TripPlan } from '../shared/types';
import './TripFavorites.css';

export default function TripFavorites() {
    const navigate = useNavigate();
    const [favoriteTrips, setFavoriteTrips] = useState<TripPlan[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        loadFavoriteTrips();
    }, []);

    const loadFavoriteTrips = async () => {
        const token = localStorage.getItem('token');
        if (!token) {
            setError('请先登录');
            setIsLoading(false);
            return;
        }

        try {
            const response = await fetch(getApiUrl('/api/trips/favorites/list'), {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            const data = await response.json();

            if (data.success) {
                setFavoriteTrips(data.data || []);
            } else {
                setError(data.message || '加载失败');
            }
        } catch (err) {
            console.error('加载收藏失败:', err);
            setError('网络错误，请稍后重试');
        } finally {
            setIsLoading(false);
        }
    };

    const removeFavorite = async (tripId: string) => {
        const token = localStorage.getItem('token');
        if (!token) return;

        try {
            const response = await fetch(getApiUrl(`/api/trips/favorites/${tripId}`), {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            const data = await response.json();

            if (data.success) {
                // 从列表中移除
                setFavoriteTrips(prev => prev.filter(trip => trip.id !== tripId));
            } else {
                alert(data.message || '取消收藏失败');
            }
        } catch (err) {
            console.error('取消收藏失败:', err);
            alert('操作失败，请稍后重试');
        }
    };

    const viewTrip = (trip: TripPlan) => {
        // 通过 state 传递行程数据到行程规划页面
        navigate('/trips/plan', { state: { tripPlan: trip, isFavorited: true } });
    };

    return (
        <div className="trip-favorites-page">
            <div className="favorites-container">
                <div className="favorites-header">
                    <button className="back-button" onClick={() => navigate('/dashboard')}>
                        ← 返回
                    </button>
                    <h2>⭐ 我的收藏行程</h2>
                </div>

                {isLoading ? (
                    <div className="loading-state">
                        <p>加载中...</p>
                    </div>
                ) : error ? (
                    <div className="error-state">
                        <p>{error}</p>
                    </div>
                ) : favoriteTrips.length === 0 ? (
                    <div className="empty-state">
                        <div className="empty-icon">📋</div>
                        <p>还没有收藏的行程</p>
                        <button
                            className="create-trip-button"
                            onClick={() => navigate('/trips/plan')}
                        >
                            创建新行程
                        </button>
                    </div>
                ) : (
                    <div className="favorites-grid">
                        {favoriteTrips.map((trip) => (
                            <div key={trip.id} className="favorite-trip-card">
                                <div className="card-header">
                                    <h3>{trip.destination}</h3>
                                    <button
                                        className="remove-favorite-btn"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            if (confirm('确定取消收藏此行程？')) {
                                                removeFavorite(trip.id);
                                            }
                                        }}
                                        title="取消收藏"
                                    >
                                        ★
                                    </button>
                                </div>

                                <div className="card-info">
                                    <div className="info-item">
                                        <span className="icon">📅</span>
                                        <span>{formatDate(trip.startDate)} - {formatDate(trip.endDate)}</span>
                                    </div>
                                    <div className="info-item">
                                        <span className="icon">💰</span>
                                        <span>¥{trip.totalCost.toFixed(2)}</span>
                                    </div>
                                    <div className="info-item">
                                        <span className="icon">📍</span>
                                        <span>{trip.itinerary.length} 天行程</span>
                                    </div>
                                </div>

                                <p className="trip-summary">{trip.summary}</p>

                                <button
                                    className="view-trip-button"
                                    onClick={() => viewTrip(trip)}
                                >
                                    查看详情
                                </button>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
