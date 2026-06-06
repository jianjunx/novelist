import { create } from 'zustand'
import api from '../api/client'

interface Project {
  id: string
  title: string
  genre: string
  description: string
  style_guide: string
  created_at: string
  updated_at: string
}
interface ProjectState {
  projects: Project[]
  currentProject: Project | null
  isLoading: boolean
  fetchProjects: () => Promise<void>
  fetchProject: (id: string) => Promise<void>
  createProject: (d: Partial<Project>) => Promise<Project>
  setCurrentProject: (p: Project | null) => void
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [],
  currentProject: null,
  isLoading: false,
  fetchProjects: async () => {
    set({ isLoading: true })
    const { data } = await api.get('/projects')
    set({ projects: data, isLoading: false })
  },
  fetchProject: async (id) => {
    const { data } = await api.get(`/projects/${id}`)
    set({ currentProject: data })
  },
  createProject: async (d) => {
    const { data } = await api.post('/projects', d)
    set({ projects: [data, ...get().projects] })
    return data
  },
  setCurrentProject: (p) => set({ currentProject: p }),
}))
