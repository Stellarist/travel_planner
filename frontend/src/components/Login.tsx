import { useState } from 'react'
import type { FormEvent } from 'react'
import { getApiUrl } from '../config'
import '../shared/common.css'
import './Login.css'

interface LoginProps {
    onLoginSuccess: (user: any) => void
}

function Login({ onLoginSuccess }: LoginProps) {
    const [isLogin, setIsLogin] = useState(true)
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault()
        setError('')
        setLoading(true)

        try {
            const endpoint = isLogin ? '/api/auth/login' : '/api/auth/register'
            const response = await fetch(getApiUrl(endpoint), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            })

            const data = await response.json()

            if (data.success) {
                if (isLogin) {
                    localStorage.setItem('token', data.token)
                    localStorage.setItem('user', JSON.stringify(data.user))
                    onLoginSuccess(data.user)
                } else {
                    alert('注册成功！请登录')
                    setIsLogin(true)
                    setPassword('')
                }
            } else {
                setError(data.message || '操作失败')
            }
        } catch {
            setError('网络错误，请稍后重试')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="login-container">
            <div className="login-box">
                <h1 className="login-title">旅行规划助手</h1>
                <h2 className="login-subtitle">{isLogin ? '用户登录' : '用户注册'}</h2>

                <form onSubmit={handleSubmit} className="login-form">
                    <div className="form-group">
                        <label htmlFor="username">用户名</label>
                        <input
                            id="username"
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="请输入用户名"
                            required
                            disabled={loading}
                        />
                    </div>

                    <div className="form-group">
                        <label htmlFor="password">密码</label>
                        <input
                            id="password"
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="请输入密码"
                            required
                            disabled={loading}
                        />
                    </div>

                    {error && <div className="error-message">{error}</div>}

                    <button type="submit" className="submit-button" disabled={loading}>
                        {loading ? '处理中...' : (isLogin ? '登录' : '注册')}
                    </button>
                </form>

                <div className="toggle-mode">
                    <button
                        type="button"
                        onClick={() => {
                            setIsLogin(!isLogin)
                            setError('')
                        }}
                        className="toggle-button"
                        disabled={loading}
                    >
                        {isLogin ? '没有账号？去注册' : '已有账号？去登录'}
                    </button>
                </div>
            </div>
        </div>
    )
}

export default Login
