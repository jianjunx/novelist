import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useProjectStore } from '../stores/projectStore'

interface Project {
  id: string
  short_id: string
  title: string
  genre?: string
  description?: string
  created_at?: string
  updated_at?: string
  brainstormed?: boolean
  has_chapters?: boolean
  has_content?: boolean
  first_chapter_id?: string | null
}

export default function ProjectCard({ project }: { project: Project }) {
  const navigate = useNavigate()
  const deleteProject = useProjectStore((s) => s.deleteProject)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [confirmName, setConfirmName] = useState('')
  const [deleting, setDeleting] = useState(false)

  const date = project.updated_at || project.created_at
  const formattedDate = date ? new Date(date).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }) : ''

  const handleClick = () => {
    const sid = project.short_id || project.id
    if (project.has_chapters) {
      navigate(`/projects/${sid}/chapters`)
    } else {
      navigate(`/projects/${sid}/creator`)
    }
  }

  const handleDelete = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setShowDeleteModal(true)
  }

  const confirmDelete = async () => {
    setDeleting(true)
    try {
      await deleteProject(project.short_id || project.id)
      setShowDeleteModal(false)
    } finally {
      setDeleting(false)
    }
  }

  const statusLabel = () => {
    if (project.has_content) {
      return { text: '写作中', color: 'bg-sage/10 text-sage', dot: 'bg-sage' }
    }
    if (project.brainstormed) {
      return { text: '已构思', color: 'bg-amber/10 text-amber-dark', dot: 'bg-amber' }
    }
    return { text: '待构思', color: 'bg-parchment-dark text-warm-gray', dot: 'bg-warm-gray-light' }
  }

  const status = statusLabel()

  return (
    <>
      <div
        onClick={handleClick}
        className="group bg-white rounded-xl border border-parchment-deep/30 p-6 cursor-pointer transition-all duration-300 hover:shadow-xl hover:shadow-ink/5 hover:border-amber/20 hover:-translate-y-1 relative"
      >
        {/* Delete button */}
        <button
          onClick={handleDelete}
          className="absolute top-3 right-3 w-7 h-7 rounded-lg bg-parchment-dark flex items-center justify-center text-ink-muted hover:text-terracotta hover:bg-terracotta/10 opacity-0 group-hover:opacity-100 transition-all duration-300"
          title="删除项目"
        >
          <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14M10 11v6M14 11v6" />
          </svg>
        </button>

        {/* Top accent line */}
        <div className="w-8 h-0.5 bg-gradient-to-r from-amber to-amber-light rounded-full mb-4 group-hover:w-12 transition-all duration-500" />

        <div className="flex items-start justify-between gap-3 mb-2">
          <h3 className="text-lg font-serif font-semibold text-ink group-hover:text-amber transition-colors duration-300 pr-8">
            {project.title}
          </h3>
          <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 ${status.color} text-xs rounded-full flex-shrink-0`}>
            <span className={`w-1.5 h-1.5 rounded-full ${status.dot}`} />
            {status.text}
          </span>
        </div>

        {project.genre && (
          <span className="inline-block px-2.5 py-0.5 bg-parchment-dark text-ink-muted text-xs rounded-full mb-3 font-literary">
            {project.genre}
          </span>
        )}

        <p className="text-sm text-ink-muted leading-relaxed line-clamp-3 mb-4 font-literary">
          {project.description || '尚未开始构思...'}
        </p>

        <div className="flex items-center justify-between pt-3 border-t border-parchment-deep/30">
          <span className="text-xs text-warm-gray">{formattedDate}</span>
          <div className="flex items-center gap-1 text-xs text-amber opacity-0 group-hover:opacity-100 transition-opacity duration-300">
            <span>{project.has_chapters ? '进入' : '开始构思'}</span>
            <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </div>
        </div>
      </div>

      {/* Delete confirmation modal */}
      {showDeleteModal && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 backdrop-blur-sm"
          onClick={(e) => { e.stopPropagation(); setShowDeleteModal(false); setConfirmName('') }}
        >
          <div
            className="bg-white rounded-2xl shadow-2xl w-full max-w-md mx-4 p-6 animate-slide-up"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-xl bg-terracotta/10 flex items-center justify-center">
                <svg className="w-5 h-5 text-terracotta" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14M10 11v6M14 11v6" />
                </svg>
              </div>
              <div>
                <h2 className="text-lg font-serif font-semibold text-ink">删除项目</h2>
                <p className="text-xs text-warm-gray">此操作不可撤销</p>
              </div>
            </div>

            <p className="text-sm text-ink-muted mb-4 font-literary">
              你即将删除项目 <span className="font-semibold text-ink">「{project.title}」</span>，所有章节、人物、世界观设定等数据都将被永久删除。
            </p>

            <p className="text-sm text-ink-muted mb-2 font-literary">
              请输入项目名称 <span className="font-semibold text-ink">「{project.title}」</span> 以确认删除：
            </p>

            <input
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              placeholder={project.title}
              className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray text-sm mb-4"
              autoFocus
            />

            <div className="flex items-center justify-end gap-3">
              <button
                onClick={(e) => { e.stopPropagation(); setShowDeleteModal(false); setConfirmName('') }}
                className="px-4 py-2 text-sm text-ink-muted hover:text-ink transition-colors"
              >
                取消
              </button>
              <button
                onClick={(e) => { e.stopPropagation(); confirmDelete() }}
                disabled={confirmName !== project.title || deleting}
                className="px-4 py-2 bg-terracotta text-white rounded-lg text-sm font-medium shadow-md shadow-terracotta/20 hover:bg-terracotta-light disabled:opacity-40 disabled:cursor-not-allowed transition-all"
              >
                {deleting ? '删除中...' : '确认删除'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
