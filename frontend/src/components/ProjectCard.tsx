import { useNavigate } from 'react-router-dom'

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
    <div
      onClick={handleClick}
      className="group bg-white rounded-xl border border-parchment-deep/30 p-6 cursor-pointer transition-all duration-300 hover:shadow-xl hover:shadow-ink/5 hover:border-amber/20 hover:-translate-y-1"
    >
      {/* Top accent line */}
      <div className="w-8 h-0.5 bg-gradient-to-r from-amber to-amber-light rounded-full mb-4 group-hover:w-12 transition-all duration-500" />

      <div className="flex items-start justify-between gap-3 mb-2">
        <h3 className="text-lg font-serif font-semibold text-ink group-hover:text-amber transition-colors duration-300">
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
  )
}
