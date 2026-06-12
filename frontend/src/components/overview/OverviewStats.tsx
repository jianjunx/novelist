import { useState } from 'react'
import type { ProjectOverview } from '../../types/overview'

interface Props {
  overview: ProjectOverview
  editing: boolean
  onEdit: (data: { genre?: string; style_guide?: string; description?: string }) => void
}

export default function OverviewStats({ overview, editing, onEdit }: Props) {
  const { stats, project } = overview
  const progress = stats.chapter_count > 0
    ? Math.round((stats.written_count / stats.chapter_count) * 100)
    : 0

  const [genre, setGenre] = useState(project.genre)
  const [styleGuide, setStyleGuide] = useState(project.style_guide)
  const [description, setDescription] = useState(project.description)

  return (
    <section id="stats" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <h2 className="text-xl font-serif font-semibold text-ink mb-4">项目概览</h2>
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

      {editing && (
        <div className="mt-6 space-y-4 border-t border-parchment-deep/20 pt-4">
          <h3 className="text-lg font-serif font-semibold text-ink">项目信息</h3>
          <div>
            <label className="block text-sm font-literary text-warm-gray mb-1">题材</label>
            <input
              type="text"
              value={genre}
              onChange={(e) => setGenre(e.target.value)}
              onBlur={() => onEdit({ genre })}
              className="w-full px-3 py-2 border border-parchment-deep/30 rounded-lg bg-parchment-dark/30 text-ink font-literary focus:outline-none focus:ring-2 focus:ring-amber/40"
            />
          </div>
          <div>
            <label className="block text-sm font-literary text-warm-gray mb-1">风格指南</label>
            <input
              type="text"
              value={styleGuide}
              onChange={(e) => setStyleGuide(e.target.value)}
              onBlur={() => onEdit({ style_guide: styleGuide })}
              className="w-full px-3 py-2 border border-parchment-deep/30 rounded-lg bg-parchment-dark/30 text-ink font-literary focus:outline-none focus:ring-2 focus:ring-amber/40"
            />
          </div>
          <div>
            <label className="block text-sm font-literary text-warm-gray mb-1">简介</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              onBlur={() => onEdit({ description })}
              rows={3}
              className="w-full px-3 py-2 border border-parchment-deep/30 rounded-lg bg-parchment-dark/30 text-ink font-literary focus:outline-none focus:ring-2 focus:ring-amber/40 resize-none"
            />
          </div>
        </div>
      )}
    </section>
  )
}
