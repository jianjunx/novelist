import { create } from 'zustand'
import api from '../api/client'

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

interface ProjectState {
  projects: Project[]
  currentProject: Project | null
  chapters: Chapter[]
  isLoading: boolean
  isGenerating: boolean
  isReviewing: boolean
  reviewResult: ReviewResult | null
  fetchProjects: () => Promise<void>
  fetchProject: (id: string) => Promise<void>
  fetchChapters: (projectId: string) => Promise<void>
  createProject: (d: Partial<Project>) => Promise<Project>
  generateChapter: (chapterId: string) => Promise<ReviewResult>
  reviewAndRevise: (chapterId: string) => Promise<ReviewResult>
  setCurrentProject: (p: Project | null) => void
  clearReviewResult: () => void
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [],
  currentProject: null,
  chapters: [],
  isLoading: false,
  isGenerating: false,
  isReviewing: false,
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
  fetchChapters: async (projectId) => {
    const { data } = await api.get(`/projects/${projectId}/chapters`)
    set({ chapters: data })
  },
  createProject: async (d) => {
    const { data } = await api.post('/projects', d)
    set({ projects: [data, ...get().projects] })
    return data
  },
  generateChapter: async (chapterId) => {
    set({ isGenerating: true, reviewResult: null })
    try {
      const { data } = await api.post(`/chapters/${chapterId}/generate-review`)
      const result: ReviewResult = data
      set({ reviewResult: result })
      // Refresh chapters list
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
      // Refresh chapters list
      const currentProject = get().currentProject
      if (currentProject) {
        await get().fetchChapters(currentProject.short_id)
      }
      return result
    } finally {
      set({ isReviewing: false })
    }
  },
  setCurrentProject: (p) => set({ currentProject: p }),
  clearReviewResult: () => set({ reviewResult: null }),
}))

export type { ReviewResult, DiscussionResult, Suggestion }
