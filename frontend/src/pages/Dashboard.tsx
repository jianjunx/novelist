import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useProjectStore } from '../stores/projectStore'
import { useAuthStore } from '../stores/authStore'
import ProjectCard from '../components/ProjectCard'

export default function Dashboard() {
  const { projects, fetchProjects, createProject, isLoading } = useProjectStore()
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const [showCreate, setShowCreate] = useState(false)
  const [newTitle, setNewTitle] = useState('')

  useEffect(() => { fetchProjects() }, [])

  const handleCreate = async () => {
    if (!newTitle.trim()) return
    const project = await createProject({ title: newTitle })
    setShowCreate(false)
    setNewTitle('')
    navigate(`/projects/${project.id}/creator`)
  }

  return (
    <div className="min-h-screen bg-parchment-gradient">
      {/* Header */}
      <header className="bg-white/70 backdrop-blur-md border-b border-parchment-deep/50 sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-ink flex items-center justify-center">
              <svg className="w-5 h-5 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
              </svg>
            </div>
            <h1 className="text-xl font-serif font-bold text-ink tracking-wide">Novelist</h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 px-3 py-1.5 bg-parchment-dark rounded-full">
              <div className="w-6 h-6 rounded-full bg-amber/20 flex items-center justify-center">
                <span className="text-xs font-bold text-amber">{user?.username?.[0]?.toUpperCase()}</span>
              </div>
              <span className="text-sm text-ink-light font-medium">{user?.username}</span>
            </div>
            <button
              onClick={logout}
              className="text-sm text-warm-gray hover:text-terracotta transition-colors"
            >
              退出
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-10">
        {/* Hero section */}
        <div className="mb-10 animate-fade-in">
          <h2 className="text-3xl font-serif font-bold text-ink mb-2">我的书房</h2>
          <p className="text-ink-muted font-literary">在这里开始你的创作之旅</p>
        </div>

        {/* Actions bar */}
        <div className="flex justify-between items-center mb-8 animate-fade-in delay-1">
          <div className="flex items-center gap-2 text-sm text-warm-gray">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
            </svg>
            <span>{projects.length} 个项目</span>
          </div>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-5 py-2.5 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-all duration-300 shadow-lg shadow-ink/10 hover:shadow-ink/20 group"
          >
            <svg className="w-4 h-4 transition-transform group-hover:rotate-90 duration-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            <span className="font-medium">新建项目</span>
          </button>
        </div>

        {/* Create modal */}
        {showCreate && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => { setShowCreate(false); setNewTitle('') }}>
            <div className="absolute inset-0 bg-ink/40 backdrop-blur-sm animate-fade-in" />
            <div
              className="relative bg-parchment rounded-2xl shadow-2xl shadow-ink/20 w-full max-w-md animate-slide-up"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="px-6 py-5 border-b border-parchment-deep/30">
                <h3 className="text-xl font-serif font-semibold text-ink">创建新项目</h3>
                <p className="text-sm text-ink-muted font-literary mt-1">为你的小说起个名字，开始构思之旅</p>
              </div>
              <div className="px-6 py-5">
                <input
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') handleCreate() }}
                  placeholder="例如：长安夜雨、星河彼岸..."
                  className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray text-lg font-literary"
                  autoFocus
                />
              </div>
              <div className="px-6 py-4 border-t border-parchment-deep/30 flex justify-end gap-3">
                <button
                  onClick={() => { setShowCreate(false); setNewTitle('') }}
                  className="px-5 py-2.5 text-warm-gray hover:text-ink transition-colors rounded-lg"
                >
                  取消
                </button>
                <button
                  onClick={handleCreate}
                  disabled={!newTitle.trim()}
                  className="px-6 py-2.5 bg-amber text-white font-medium rounded-lg hover:bg-amber-light transition-colors disabled:opacity-40 shadow-md shadow-amber/20"
                >
                  开始构思
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Projects grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="bg-white rounded-xl border border-parchment-deep/30 p-6 animate-pulse">
                <div className="h-6 bg-parchment-dark rounded w-2/3 mb-4" />
                <div className="h-4 bg-parchment-dark rounded w-1/4 mb-4" />
                <div className="space-y-2">
                  <div className="h-3 bg-parchment-dark rounded" />
                  <div className="h-3 bg-parchment-dark rounded w-4/5" />
                </div>
              </div>
            ))}
          </div>
        ) : projects.length === 0 ? (
          <div className="text-center py-20 animate-fade-in">
            <div className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-parchment-dark mb-6">
              <svg className="w-12 h-12 text-warm-gray-light" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1">
                <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
              </svg>
            </div>
            <h3 className="text-xl font-serif text-ink mb-2">还没有项目</h3>
            <p className="text-warm-gray font-literary mb-6">点击「新建项目」开始你的第一部小说</p>
            <button
              onClick={() => setShowCreate(true)}
              className="px-6 py-3 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors"
            >
              创建第一个项目
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((p, i) => (
              <div key={p.id} className="animate-fade-in" style={{ animationDelay: `${i * 0.1}s`, opacity: 0 }}>
                <ProjectCard project={p} />
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
