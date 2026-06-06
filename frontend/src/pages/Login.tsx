import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

export default function Login() {
  const [isLogin, setIsLogin] = useState(true)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const { login, register, isLoading, error } = useAuthStore()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      isLogin ? await login(username, password) : await register(username, password)
      navigate('/')
    } catch {}
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-ink-wash relative overflow-hidden">
      {/* Decorative ink wash circles */}
      <div className="absolute top-[-20%] right-[-10%] w-[500px] h-[500px] rounded-full bg-amber/5 blur-3xl" />
      <div className="absolute bottom-[-15%] left-[-5%] w-[400px] h-[400px] rounded-full bg-terracotta/5 blur-3xl" />

      {/* Floating brush strokes */}
      <div className="absolute top-20 left-20 w-32 h-0.5 bg-gradient-to-r from-amber/30 to-transparent rotate-12 animate-float" />
      <div className="absolute bottom-32 right-24 w-24 h-0.5 bg-gradient-to-l from-amber/20 to-transparent -rotate-6 animate-float delay-2" />

      <div className="max-w-md w-full mx-4 animate-ink-drop">
        {/* Logo area */}
        <div className="text-center mb-10">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-amber/10 border border-amber/20 mb-6">
            <svg className="w-10 h-10 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
            </svg>
          </div>
          <h1 className="text-4xl font-serif font-bold text-parchment tracking-wider">Novelist</h1>
          <div className="mt-3 flex items-center justify-center gap-3">
            <div className="w-12 h-px bg-gradient-to-r from-transparent to-amber/40" />
            <p className="text-amber-light text-sm font-literary tracking-widest">AI小说创作平台</p>
            <div className="w-12 h-px bg-gradient-to-l from-transparent to-amber/40" />
          </div>
        </div>

        {/* Card */}
        <div className="bg-parchment/95 backdrop-blur-sm rounded-2xl shadow-2xl shadow-black/20 p-8">
          {/* Tabs */}
          <div className="flex mb-8 bg-parchment-dark rounded-lg p-1">
            <button
              onClick={() => setIsLogin(true)}
              className={`flex-1 py-2.5 rounded-md text-sm font-medium transition-all duration-300 ${
                isLogin
                  ? 'bg-white text-ink shadow-sm'
                  : 'text-ink-muted hover:text-ink'
              }`}
            >
              登录
            </button>
            <button
              onClick={() => setIsLogin(false)}
              className={`flex-1 py-2.5 rounded-md text-sm font-medium transition-all duration-300 ${
                !isLogin
                  ? 'bg-white text-ink shadow-sm'
                  : 'text-ink-muted hover:text-ink'
              }`}
            >
              注册
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="bg-terracotta-light border border-terracotta/20 text-terracotta px-4 py-3 rounded-lg text-sm animate-fade-in">
                {error}
              </div>
            )}
            <div className="space-y-2">
              <label className="block text-sm font-medium text-ink-light">用户名</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray transition-all duration-200"
                placeholder="请输入用户名"
                required
                minLength={3}
              />
            </div>
            <div className="space-y-2">
              <label className="block text-sm font-medium text-ink-light">密码</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray transition-all duration-200"
                placeholder="请输入密码"
                required
                minLength={6}
              />
            </div>
            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-3 bg-ink text-parchment font-medium rounded-lg hover:bg-ink-light transition-all duration-300 disabled:opacity-50 relative overflow-hidden group"
            >
              <span className={`inline-block transition-transform duration-300 ${isLoading ? '-translate-y-8 opacity-0' : 'translate-y-0 opacity-100'}`}>
                {isLogin ? '登录' : '注册'}
              </span>
              {isLoading && (
                <span className="absolute inset-0 flex items-center justify-center">
                  <svg className="animate-spin h-5 w-5 text-parchment" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                </span>
              )}
            </button>
          </form>
        </div>

        {/* Footer */}
        <p className="text-center text-warm-gray text-xs mt-8 font-literary">
          以笔为剑，以墨为锋
        </p>
      </div>
    </div>
  )
}
