import type { LocationEntry } from '../../types/overview'

interface Props {
  locations: LocationEntry[]
}

export default function LocationIndex({ locations }: Props) {
  return (
    <section id="locations" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <h2 className="text-xl font-serif font-semibold text-ink mb-4">地点索引</h2>
      {locations.length === 0 ? (
        <p className="text-sm text-warm-gray font-literary">暂无地点数据，可在世界观设定中添加 category 为「地点」的条目</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {locations.map((loc) => (
            <div key={loc.name} className="border border-parchment-deep/30 rounded-lg p-4">
              <h3 className="font-serif font-semibold text-ink">{loc.name}</h3>
              {loc.description && (
                <p className="text-sm text-ink-muted font-literary mt-1 leading-relaxed">{loc.description}</p>
              )}
              {loc.events.length > 0 && (
                <div className="mt-3">
                  <p className="text-xs text-warm-gray mb-1">关联事件</p>
                  <ul className="space-y-1">
                    {loc.events.map((ev, i) => (
                      <li key={i} className="text-xs text-ink-muted">{ev}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
