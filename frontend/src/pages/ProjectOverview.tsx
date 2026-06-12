import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import ProjectNav from '../components/ProjectNav'
import OverviewStats from '../components/overview/OverviewStats'
import WorldSettingsPanel from '../components/overview/WorldSettingsPanel'
import CharacterCards from '../components/overview/CharacterCards'
import RelationshipGraph from '../components/overview/RelationshipGraph'
import EventTimeline from '../components/overview/EventTimeline'
import LocationIndex from '../components/overview/LocationIndex'
import { useProjectStore } from '../stores/projectStore'
import { aggregateLocations } from '../types/overview'

const anchors = [
  { id: 'stats', label: '概览' },
  { id: 'world', label: '世界观' },
  { id: 'characters', label: '人物' },
  { id: 'relationships', label: '关系图' },
  { id: 'timeline', label: '时间线' },
  { id: 'locations', label: '地点' },
]

export default function ProjectOverview() {
  const { projectId } = useParams<{ projectId: string }>()
  const {
    overview, isOverviewLoading, overviewError, fetchOverview, fetchProject,
    updateProjectOverview,
    createCharacter, updateCharacter, deleteCharacter,
    createWorldSetting, updateWorldSetting, deleteWorldSetting,
  } = useProjectStore()
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    if (projectId) {
      fetchProject(projectId)
      fetchOverview(projectId)
    }
  }, [projectId])

  const pid = projectId!

  const handleProjectEdit = (data: { genre?: string; style_guide?: string; description?: string }) => {
    updateProjectOverview(pid, data)
  }

  if (isOverviewLoading) {
    return (
      <div className="min-h-screen bg-parchment-gradient flex items-center justify-center">
        <div className="animate-pulse text-warm-gray font-literary">加载设定数据...</div>
      </div>
    )
  }

  if (overviewError || !overview) {
    return (
      <div className="min-h-screen flex flex-col bg-parchment-gradient">
        <ProjectNav projectId={pid} currentTab="overview" />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <p className="text-ink-muted font-literary mb-4">{overviewError || '无法加载设定数据'}</p>
            <button
              onClick={() => fetchOverview(pid)}
              className="px-6 py-3 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors"
            >
              重试
            </button>
          </div>
        </div>
      </div>
    )
  }

  const locations = aggregateLocations(overview.world_settings, overview.outlines)



  return (
    <div className="min-h-screen flex flex-col bg-parchment-gradient">
      <ProjectNav
        projectId={pid}
        currentTab="overview"
        actions={
          <button
            onClick={() => setEditing(!editing)}
            className={`px-4 py-2 rounded-lg text-sm font-literary transition-colors ${
              editing ? 'bg-amber text-white' : 'text-amber-dark hover:bg-amber/5'
            }`}
          >
            {editing ? '完成编辑' : '编辑设定'}
          </button>
        }
      />

      <div className="flex-1 max-w-[1400px] mx-auto w-full px-6 py-6 flex gap-6">
        <aside className="hidden lg:block w-32 shrink-0 sticky top-24 self-start">
          <nav className="space-y-1">
            {anchors.map((a) => (
              <a
                key={a.id}
                href={`#${a.id}`}
                className="block px-3 py-2 text-sm text-ink-muted hover:text-amber font-literary rounded-lg hover:bg-white/50 transition-colors"
              >
                {a.label}
              </a>
            ))}
          </nav>
        </aside>

        <main className="flex-1 space-y-6 min-w-0">
          <OverviewStats overview={overview} editing={editing} onEdit={handleProjectEdit} />
          <WorldSettingsPanel
            settings={overview.world_settings}
            editing={editing}
            onCreate={(category, content) => createWorldSetting(pid, category, content)}
            onUpdate={(id, category, content) => updateWorldSetting(id, category, content)}
            onDelete={deleteWorldSetting}
          />
          <CharacterCards
            characters={overview.characters}
            editing={editing}
            onCreate={(data) => createCharacter(pid, data)}
            onUpdate={(id, data) => updateCharacter(id, data)}
            onDelete={deleteCharacter}
          />
          <RelationshipGraph characters={overview.characters} />
          <EventTimeline outlines={overview.outlines} />
          <LocationIndex locations={locations} />
        </main>
      </div>
    </div>
  )
}
