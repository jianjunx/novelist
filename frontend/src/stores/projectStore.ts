import { create } from 'zustand'
import api from '../api/client'
import type { Character, ProjectOverview, WorldSetting, OutlineItem } from '../types/overview'
import { parseKeyEvents, parseRelationships } from '../types/overview'

interface Project {
  id: string
  short_id: string
  title: string
  genre: string
  description: string
  style_guide: string
  created_at: string
  updated_at: string
  brainstormed: boolean
  has_chapters: boolean
  has_content: boolean
  first_chapter_id: string | null
}

interface Volume {
  id: string
  project_id: string
  volume_num: number
  title: string
  description: string
  status: string
  created_at: string
  updated_at: string
}

interface Chapter {
  id: string
  project_id: string
  outline_id: string | null
  chapter_num: number
  title: string
  content: string
  word_count: number
  status: string
  outline_summary: string
  volume_id: string | null
  volume_num: number
  volume_title: string
  can_generate: boolean
}

interface Suggestion {
  type: string
  location: string
  problem: string
  suggestion: string
  priority: number
}

interface DiscussionResult {
  editor_suggestions: Suggestion[]
  reader_feedback: string
  critic_analysis: string
  aggregated: Suggestion[]
}

interface ReviewResult {
  discussion: DiscussionResult | null
  revised_content: string
  round_num: number
}

function normalizeOverview(data: ProjectOverview): ProjectOverview {
  return {
    ...data,
    characters: data.characters.map((c) => ({
      ...c,
      relationships: parseRelationships(c.relationships),
    })),
    outlines: data.outlines.map((o) => ({
      ...o,
      key_events: parseKeyEvents(o.key_events),
    })),
  }
}

interface ProjectState {
  projects: Project[]
  currentProject: Project | null
  chapters: Chapter[]
  volumes: Volume[]
  overview: ProjectOverview | null
  isLoading: boolean
  isOverviewLoading: boolean
  isGenerating: boolean
  isReviewing: boolean
  isExpanding: boolean
  reviewResult: ReviewResult | null
  fetchProjects: () => Promise<void>
  fetchProject: (id: string) => Promise<void>
  fetchOverview: (projectId: string) => Promise<void>
  fetchChapters: (projectId: string) => Promise<void>
  fetchVolumes: (projectId: string) => Promise<void>
  createProject: (d: Partial<Project>) => Promise<Project>
  createVolume: (projectId: string) => Promise<Volume>
  generateChapter: (chapterId: string) => Promise<ReviewResult>
  reviewAndRevise: (chapterId: string) => Promise<ReviewResult>
  expandOutlines: (projectId: string) => Promise<{ volume_complete: boolean }>
  deleteProject: (projectId: string) => Promise<void>
  updateProject: (projectId: string, data: Partial<Project>) => Promise<void>
  updateProjectOverview: (projectId: string, data: Partial<Pick<Project, 'genre' | 'description' | 'style_guide' | 'title'>>) => Promise<void>
  createCharacter: (projectId: string, data: Omit<Character, 'id' | 'project_id' | 'created_at'>) => Promise<void>
  updateCharacter: (id: string, data: Partial<Character>) => Promise<void>
  deleteCharacter: (id: string) => Promise<void>
  createWorldSetting: (projectId: string, category: string, content: string) => Promise<void>
  updateWorldSetting: (id: string, category: string, content: string) => Promise<void>
  deleteWorldSetting: (id: string) => Promise<void>
  setCurrentProject: (p: Project | null) => void
  clearReviewResult: () => void
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [],
  currentProject: null,
  chapters: [],
  volumes: [],
  overview: null,
  isLoading: false,
  isOverviewLoading: false,
  isGenerating: false,
  isReviewing: false,
  isExpanding: false,
  reviewResult: null,
  fetchProjects: async () => {
    set({ isLoading: true })
    const { data } = await api.get('/projects')
    set({ projects: data, isLoading: false })
  },
  fetchProject: async (id) => {
    const { data } = await api.get(`/projects/${id}`)
    set({ currentProject: data })
  },
  fetchOverview: async (projectId) => {
    set({ isOverviewLoading: true })
    try {
      const { data } = await api.get(`/projects/${projectId}/overview`)
      set({ overview: normalizeOverview(data), isOverviewLoading: false })
    } catch {
      set({ isOverviewLoading: false })
    }
  },
  fetchChapters: async (projectId) => {
    const { data } = await api.get(`/projects/${projectId}/chapters`)
    set({ chapters: data })
  },
  fetchVolumes: async (projectId) => {
    const { data } = await api.get(`/projects/${projectId}/volumes`)
    set({ volumes: data })
  },
  createProject: async (d) => {
    const { data } = await api.post('/projects', d)
    set({ projects: [data, ...get().projects] })
    return data
  },
  createVolume: async (projectId) => {
    const { data } = await api.post(`/projects/${projectId}/volumes`)
    set({ volumes: [...get().volumes, data] })
    return data
  },
  generateChapter: async (chapterId) => {
    set({ isGenerating: true, reviewResult: null })
    try {
      const { data } = await api.post(`/chapters/${chapterId}/generate-review`)
      const result: ReviewResult = data
      set({ reviewResult: result })
      const currentProject = get().currentProject
      if (currentProject) {
        await get().fetchChapters(currentProject.short_id)
      }
      return result
    } finally {
      set({ isGenerating: false })
    }
  },
  reviewAndRevise: async (chapterId) => {
    set({ isReviewing: true, reviewResult: null })
    try {
      const { data } = await api.post(`/chapters/${chapterId}/review-revise`)
      const result: ReviewResult = data
      set({ reviewResult: result })
      const currentProject = get().currentProject
      if (currentProject) {
        await get().fetchChapters(currentProject.short_id)
      }
      return result
    } finally {
      set({ isReviewing: false })
    }
  },
  expandOutlines: async (projectId) => {
    set({ isExpanding: true })
    try {
      const { data } = await api.post(`/projects/${projectId}/expand-outlines`)
      await get().fetchChapters(projectId)
      return { volume_complete: data.volume_complete || false }
    } finally {
      set({ isExpanding: false })
    }
  },
  deleteProject: async (projectId) => {
    await api.delete(`/projects/${projectId}`)
    set({ projects: get().projects.filter((p) => p.id !== projectId && p.short_id !== projectId) })
  },
  updateProject: async (projectId, data) => {
    const { data: updated } = await api.put(`/projects/${projectId}`, data)
    set({
      projects: get().projects.map((p) => (p.id === projectId || p.short_id === projectId ? { ...p, ...updated } : p)),
      currentProject: get().currentProject?.id === projectId || get().currentProject?.short_id === projectId
        ? { ...get().currentProject!, ...updated }
        : get().currentProject,
    })
  },
  updateProjectOverview: async (projectId, data) => {
    const current = get().overview
    if (!current) return
    const payload = {
      title: current.project.title,
      genre: data.genre ?? current.project.genre,
      description: data.description ?? current.project.description,
      style_guide: data.style_guide ?? current.project.style_guide,
    }
    await api.put(`/projects/${projectId}`, payload)
    set({
      overview: {
        ...current,
        project: { ...current.project, ...data },
      },
      currentProject: get().currentProject
        ? { ...get().currentProject!, ...data }
        : get().currentProject,
    })
  },
  createCharacter: async (projectId, data) => {
    const { data: created } = await api.post(`/projects/${projectId}/characters`, {
      ...data,
      relationships: data.relationships ?? [],
    })
    const overview = get().overview
    if (overview) {
      set({
        overview: {
          ...overview,
          characters: [...overview.characters, { ...created, relationships: parseRelationships(created.relationships) }],
        },
      })
    }
  },
  updateCharacter: async (id, data) => {
    await api.put(`/characters/${id}`, {
      name: data.name,
      role: data.role,
      personality: data.personality,
      background: data.background,
      appearance: data.appearance,
      relationships: data.relationships ?? [],
    })
    const overview = get().overview
    if (overview) {
      set({
        overview: {
          ...overview,
          characters: overview.characters.map((c) =>
            c.id === id ? { ...c, ...data, relationships: data.relationships ?? c.relationships } : c,
          ),
        },
      })
    }
  },
  deleteCharacter: async (id) => {
    await api.delete(`/characters/${id}`)
    const overview = get().overview
    if (overview) {
      set({
        overview: {
          ...overview,
          characters: overview.characters.filter((c) => c.id !== id),
        },
      })
    }
  },
  createWorldSetting: async (projectId, category, content) => {
    const { data: created } = await api.post(`/projects/${projectId}/world-settings`, { category, content })
    const overview = get().overview
    if (overview) {
      set({ overview: { ...overview, world_settings: [...overview.world_settings, created] } })
    }
  },
  updateWorldSetting: async (id, category, content) => {
    await api.put(`/world-settings/${id}`, { category, content })
    const overview = get().overview
    if (overview) {
      set({
        overview: {
          ...overview,
          world_settings: overview.world_settings.map((s) =>
            s.id === id ? { ...s, category, content } : s,
          ),
        },
      })
    }
  },
  deleteWorldSetting: async (id) => {
    await api.delete(`/world-settings/${id}`)
    const overview = get().overview
    if (overview) {
      set({
        overview: {
          ...overview,
          world_settings: overview.world_settings.filter((s) => s.id !== id),
        },
      })
    }
  },
  setCurrentProject: (p) => set({ currentProject: p }),
  clearReviewResult: () => set({ reviewResult: null }),
}))

export type { ReviewResult, DiscussionResult, Suggestion, ProjectOverview, Character, WorldSetting, OutlineItem }
