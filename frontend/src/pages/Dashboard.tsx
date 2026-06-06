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
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Novelist</h1>
          <div className="flex items-center space-x-4">
            <span className="text-gray-600">{user?.username}</span>
            <button onClick={logout} className="text-gray-500">退出</button>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <h2 className="text-xl font-semibold">我的项目</h2>
          <button
            onClick={() => setShowCreate(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-md"
          >
            新建项目
          </button>
        </div>
        {showCreate && (
          <div className="bg-white p-6 rounded-lg shadow mb-8">
            <input
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              placeholder="项目标题"
              className="w-full px-3 py-2 border rounded-md mb-4"
            />
            <button
              onClick={handleCreate}
              className="px-4 py-2 bg-blue-600 text-white rounded-md"
            >
              创建并开始构思
            </button>
          </div>
        )}
        {isLoading ? (
          <div className="text-center py-12">加载中...</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((p) => (
              <ProjectCard key={p.id} project={p} />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
