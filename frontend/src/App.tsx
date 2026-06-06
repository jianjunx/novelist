import { Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuthStore } from './stores/authStore'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Creator from './pages/Creator'
import ChapterEditor from './pages/ChapterEditor'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuthStore()
  return token ? <>{children}</> : <Navigate to="/login" />
}

export default function App() {
  const { checkAuth } = useAuthStore()
  useEffect(() => { checkAuth() }, [])
  return (
    <div className="min-h-screen bg-gray-50">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<PrivateRoute><Dashboard /></PrivateRoute>} />
        <Route path="/projects/:projectId/creator" element={<PrivateRoute><Creator /></PrivateRoute>} />
        <Route path="/chapters/:chapterId/edit" element={<PrivateRoute><ChapterEditor /></PrivateRoute>} />
      </Routes>
    </div>
  )
}
