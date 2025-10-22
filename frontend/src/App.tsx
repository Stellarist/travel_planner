import { useState, useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Login from './components/Login'
import Dashboard from './components/Dashboard'
import TripPlanner from './components/TripPlanner'
import './App.css'
import type { User } from './types'

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const initApp = async () => {
      // 检查本地存储
      const storedUser = localStorage.getItem('user')
      const storedToken = localStorage.getItem('token')

      if (storedUser && storedToken) {
        try {
          setUser(JSON.parse(storedUser))
        } catch {
          localStorage.clear()
        }
      }
      setLoading(false)
    }

    initApp()
  }, [])

  const handleLoginSuccess = (userData: User) => setUser(userData)

  const handleLogout = () => {
    localStorage.clear()
    setUser(null)
  }

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        fontSize: '18px',
        color: '#666'
      }}>
        加载中...
      </div>
    )
  }

  if (loading) return null

  return (
    <Routes>
      <Route path="/" element={user ? <Dashboard user={user} onLogout={handleLogout} /> : <Login onLoginSuccess={handleLoginSuccess} />} />
      <Route path="/planner" element={user ? <TripPlanner /> : <Navigate to="/" replace />} />
      {/* catch-all: redirect to root */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
