import type { ProjectOverview } from '../../types/overview'

interface Props {
  overview: ProjectOverview
  editing: boolean
  onEdit: (field: 'genre' | 'style_guide' | 'description', value: string) => void
}

export default function OverviewStats({ overview, editing, onEdit }: Props) {
  const { project, stats } = overview
  const progress = stats.chapter_count > 0
    ? Math.round((stats.written_count / stats.chapter_count) * 100)
    : 0

  return (
    <section id="stats" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <h2 className="text-xl font-serif font-semibold text-ink mb-4">项目概览</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-3">
          {editing ? (
            <>
              <div>
                <label className="text-xs text-warm-gray font-literary">类型</label>
                <input
                  defaultValue={project.genre}
                  onBlur={(e) => onEdit('genre', e.target.value)}
                  className="w-full mt-1 px-3 py-2 bg-parchment-dark border border-parchment-deep rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="text-xs text-warm-gray font-literary">风格</label>
                <input
                  defaultValue={project.style_guide}
                  onBlur={(e) => onEdit('style_guide', e.target.value)}
                  className="w-full mt-1 px-3 py-2 bg-parchment-dark border border-parchment-deep rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="text-xs text-warm-gray font-literary">简介</label>
                <textarea
                  defaultValue={project.description}
                  onBlur={(e) => onEdit('description', e.target.value)}
                  rows={3}
                  className="w-full mt-1 px-3 py-2 bg-parchment-dark border border-parchment-deep rounded-lg text-sm resize-none"
                />
              </div>
            </>
          ) : (
            <>
              {project.genre && (
                <p className="text-sm"><span className="text-warm-gray">类型：</span>{project.genre}</p>
              )}
              {project.style_guide && (
                <p className="text-sm"><span className="text-warm-gray">风格：</span>{project.style_guide}</p>
              )}
              {project.description && (
                <p className="text-sm text-ink-muted font-literary leading-relaxed">{project.description}</p>
              )}
              {!project.genre && !project.style_guide && !project.description && (
                <p className="text-sm text-warm-gray font-literary">暂无项目描述，可在编辑模式下补充</p>
              )}
            </>
          )}
        </div>
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-parchment-dark rounded-lg p-4 text-center">
            <div className="text-2xl font-serif font-bold text-ink">{stats.chapter_count}</div>
            <div className="text-xs text-warm-gray mt-1">总章节</div>
          </div>
          <div className="bg-parchment-dark rounded-lg p-4 text-center">
            <div className="text-2xl font-serif font-bold text-sage">{stats.written_count}</div>
            <div className="text-xs text-warm-gray mt-1">已写作</div>
          </div>
          <div className="bg-parchment-dark rounded-lg p-4 text-center">
            <div className="text-2xl font-serif font-bold text-amber">{stats.total_words.toLocaleString()}</div>
            <div className="text-xs text-warm-gray mt-1">总字数</div>
          </div>
        </div>
      </div>
      {stats.chapter_count > 0 && (
        <div className="mt-4">
          <div className="flex justify-between text-xs text-warm-gray mb-1">
            <span>写作进度</span>
            <span>{progress}%</span>
          </div>
          <div className="h-2 bg-parchment-dark rounded-full overflow-hidden">
            <div className="h-full bg-sage rounded-full transition-all" style={{ width: `${progress}%` }} />
          </div>
        </div>
      )}
    </section>
  )
}
