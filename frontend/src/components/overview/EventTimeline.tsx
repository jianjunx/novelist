import { useNavigate } from 'react-router-dom'
import type { OutlineItem } from '../../types/overview'
import { parseKeyEvents } from '../../types/overview'

interface Props {
  outlines: OutlineItem[]
}

export default function EventTimeline({ outlines }: Props) {
  const navigate = useNavigate()

  const byAct = outlines.reduce<Record<number, OutlineItem[]>>((acc, o) => {
    const act = o.act || 1
    if (!acc[act]) acc[act] = []
    acc[act].push(o)
    return acc
  }, {})

  const acts = Object.keys(byAct).map(Number).sort((a, b) => a - b)

  if (outlines.length === 0) {
    return (
      <section id="timeline" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
        <h2 className="text-xl font-serif font-semibold text-ink mb-4">事件时间线</h2>
        <p className="text-sm text-warm-gray font-literary">暂无大纲数据</p>
      </section>
    )
  }

  return (
    <section id="timeline" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <h2 className="text-xl font-serif font-semibold text-ink mb-4">事件时间线</h2>
      <div className="space-y-6">
        {acts.map((act) => (
          <div key={act}>
            <h3 className="text-sm font-medium text-amber-dark mb-3">第 {act} 幕</h3>
            <div className="relative pl-6 border-l-2 border-parchment-deep space-y-4">
              {byAct[act].map((outline) => {
                const events = parseKeyEvents(outline.key_events)
                return (
                  <div key={outline.id} className="relative">
                    <div className="absolute -left-[25px] w-3 h-3 rounded-full bg-amber border-2 border-white" />
                    <div className="bg-parchment-dark/50 rounded-lg p-4">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <span className="text-xs text-warm-gray">第 {outline.chapter_num} 章</span>
                          <p className="text-sm text-ink font-literary mt-1">{outline.summary}</p>
                        </div>
                        {outline.chapter_id && (
                          <button
                            onClick={() => navigate(`/chapters/${outline.chapter_id}/edit`)}
                            className="text-xs text-amber hover:underline shrink-0"
                          >
                            编辑章节 →
                          </button>
                        )}
                      </div>
                      {events.length > 0 && (
                        <ul className="mt-3 space-y-2">
                          {events.map((ev, i) => (
                            <li key={i} className="text-xs text-ink-muted flex flex-wrap gap-x-2">
                              <span className="text-ink">• {ev.event}</span>
                              {ev.location && <span className="text-warm-gray">@{ev.location}</span>}
                              {ev.characters && ev.characters.length > 0 && (
                                <span className="text-sage">{ev.characters.join('、')}</span>
                              )}
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
