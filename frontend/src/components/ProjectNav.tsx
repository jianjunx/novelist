import { useNavigate } from 'react-router-dom'
import { useProjectStore } from '../stores/projectStore'

export type ProjectTab = 'chapters' | 'overview' | 'creator'

interface ProjectNavProps {
  projectId: string
  currentTab: ProjectTab
  actions?: React.ReactNode
}

const tabs: { id: ProjectTab; label: string; path: string }[] = [
  { id: 'chapters', label: '章节', path: 'chapters' },
  { id: 'overview', label: '设定', path: 'overview' },
  { id: 'creator', label: '构思', path: 'creator' },
]

export default function ProjectNav({ projectId, currentTab, actions }: ProjectNavProps) {
  const navigate = useNavigate()
  const currentProject = useProjectStore((s) => s.currentProject)

  return (
    <header className="bg-white/70 backdrop-blur-md border-b border-parchment-deep/50 sticky top-0 z-40 shrink-0">
      <div className="max-w-[1400px] mx-auto px-6 py-3">
        <div className="flex justify-between items-center gap-4">
          <div className="flex items-center gap-3 min-w-0">
            <button
              onClick={() => navigate('/')}
              className="w-8 h-8 rounded-lg bg-parchment-dark flex items-center justify-center text-ink-muted hover:text-ink hover:bg-parchment-deep transition-colors shrink-0"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
            </button>
            <div className="min-w-0">
              <h1 className="text-lg font-serif font-semibold text-ink truncate">
                {currentProject?.title || '项目'}
              </h1>
            </div>
          </div>

          <nav className="flex items-center gap-1 bg-parchment-dark/60 rounded-lg p-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => navigate(`/projects/${projectId}/${tab.path}`)}
                className={`px-4 py-1.5 rounded-md text-sm font-literary transition-colors ${
                  currentTab === tab.id
                    ? 'bg-white text-ink shadow-sm'
                    : 'text-ink-muted hover:text-ink'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>

          <div className="flex items-center gap-2 shrink-0 min-w-[80px] justify-end">
            {actions}
          </div>
        </div>
      </div>
    </header>
  )
}
