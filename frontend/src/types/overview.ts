export interface CharacterRelationship {
  target: string
  type: string
  description?: string
}

export interface KeyEvent {
  event: string
  location?: string
  characters?: string[]
}

export interface Character {
  id: string
  project_id: string
  name: string
  role: string
  personality: string
  background: string
  appearance: string
  relationships: CharacterRelationship[] | null
  created_at: string
}

export interface WorldSetting {
  id: string
  project_id: string
  category: string
  content: string
  created_at: string
}

export interface OutlineItem {
  id: string
  project_id: string
  volume_id: string | null
  act: number
  chapter_num: number
  summary: string
  key_events: KeyEvent[] | null
  status: string
  chapter_id: string | null
  created_at: string
}

export interface ProjectStats {
  chapter_count: number
  written_count: number
  total_words: number
}

export interface ProjectOverview {
  project: {
    id: string
    short_id: string
    title: string
    genre: string
    description: string
    style_guide: string
    brainstormed: boolean
    has_chapters: boolean
    has_content: boolean
  }
  characters: Character[]
  world_settings: WorldSetting[]
  outlines: OutlineItem[]
  stats: ProjectStats
}

export function parseRelationships(raw: unknown): CharacterRelationship[] {
  if (!raw) return []
  if (Array.isArray(raw)) return raw as CharacterRelationship[]
  return []
}

export function parseKeyEvents(raw: unknown): KeyEvent[] {
  if (!raw) return []
  if (Array.isArray(raw)) return raw as KeyEvent[]
  return []
}

export interface LocationEntry {
  name: string
  description: string
  events: string[]
}

export function aggregateLocations(
  worldSettings: WorldSetting[],
  outlines: OutlineItem[],
): LocationEntry[] {
  const map = new Map<string, LocationEntry>()

  for (const ws of worldSettings) {
    if (ws.category.includes('地点')) {
      const name = ws.content.split(/[，,。]/)[0].trim() || ws.content
      map.set(name, {
        name,
        description: ws.content,
        events: map.get(name)?.events ?? [],
      })
    }
  }

  for (const outline of outlines) {
    const events = parseKeyEvents(outline.key_events)
    for (const ev of events) {
      if (!ev.location) continue
      const existing = map.get(ev.location)
      const eventLabel = `第${outline.chapter_num}章：${ev.event}`
      if (existing) {
        if (!existing.events.includes(eventLabel)) {
          existing.events.push(eventLabel)
        }
      } else {
        map.set(ev.location, { name: ev.location, description: '', events: [eventLabel] })
      }
    }
  }

  return Array.from(map.values())
}
