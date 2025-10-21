import { useState } from 'react'
import './Dashboard.css'
import TripPlanner from './TripPlanner'

interface User {
    id: number
    username: string
}

interface DashboardProps {
    user: User
    onLogout: () => void
}

function Dashboard({ user, onLogout }: DashboardProps) {
    const [isSidebarOpen, setIsSidebarOpen] = useState(false)
    const [currentView, setCurrentView] = useState<'home' | 'planner'>('home')

    const toggleSidebar = () => {
        setIsSidebarOpen(!isSidebarOpen)
    }

    const closeSidebar = () => {
        setIsSidebarOpen(false)
    }

    const navigateToPlanner = () => {
        setCurrentView('planner')
    }

    const navigateToHome = () => {
        setCurrentView('home')
    }

    return (
        <div className="dashboard-container">
            <header className="dashboard-header">
                <div className="header-content">
                    <div className="header-text">
                        <h1>旅行规划助手</h1>
                        <p className="header-subtitle">开始规划您的下一次精彩旅程</p>
                    </div>
                    <button
                        className="menu-button"
                        onClick={toggleSidebar}
                        aria-label="打开菜单"
                    >
                        <span className="menu-icon">☰</span>
                    </button>
                </div>
            </header>

            {/* 侧边栏遮罩层 */}
            {isSidebarOpen && (
                <div
                    className="sidebar-overlay"
                    onClick={closeSidebar}
                />
            )}

            {/* 侧边栏 */}
            <div className={`sidebar ${isSidebarOpen ? 'sidebar-open' : ''}`}>
                <div className="sidebar-header">
                    <h2>账户信息</h2>
                </div>

                <div className="sidebar-content">
                    <div className="user-avatar">
                        <div className="avatar-circle">
                            {user.username.charAt(0).toUpperCase()}
                        </div>
                    </div>

                    <div className="sidebar-user-info">
                        <div className="info-item">
                            <span className="info-label">👤 用户名</span>
                            <span className="info-value">{user.username}</span>
                        </div>
                        <div className="info-item">
                            <span className="info-label">🆔 用户ID</span>
                            <span className="info-value">{user.id}</span>
                        </div>
                    </div>

                    <div className="sidebar-footer">
                        <button onClick={onLogout} className="sidebar-logout-button">
                            <span className="button-icon">🚪</span>
                            <span>退出登录</span>
                        </button>
                    </div>
                </div>
            </div>

            <main className="dashboard-main">
                <div className="content-wrapper">
                    {currentView === 'home' ? (
                        <>
                            <div className="feature-cards">
                                <div className="feature-card" onClick={navigateToPlanner} style={{ cursor: 'pointer' }}>
                                    <div className="card-icon">🗺️</div>
                                    <h3>规划行程</h3>
                                    <p>创建详细的旅行计划，安排每日行程</p>
                                </div>

                                <div className="feature-card">
                                    <div className="card-icon">📍</div>
                                    <h3>探索景点</h3>
                                    <p>发现热门景点和隐藏的宝藏</p>
                                </div>

                                <div className="feature-card">
                                    <div className="card-icon">💰</div>
                                    <h3>预算管理</h3>
                                    <p>追踪旅行开支，控制预算</p>
                                </div>

                                <div className="feature-card">
                                    <div className="card-icon">📝</div>
                                    <h3>旅行日记</h3>
                                    <p>记录旅途中的美好时刻</p>
                                </div>
                            </div>
                        </>
                    ) : (
                        <>
                            <button onClick={navigateToHome} className="back-button">
                                ← 返回首页
                            </button>
                            <TripPlanner user={user} />
                        </>
                    )}
                </div>
            </main>
        </div>
    )
}

export default Dashboard
