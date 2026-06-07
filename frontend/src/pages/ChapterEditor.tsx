import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import api from '../api/client'
import DiscussionPanel from '../components/DiscussionPanel'
import TipTapEditor, { type TipTapEditorHandle } from '../components/TipTapEditor'

interface Chapter {
  id: string
  project_id: string
  chapter_num: number
  title: string
  content: string
  word_count: number
  status: string
}

export default function ChapterEditor() {
  const { chapterId } = useParams<{ chapterId: string }>()
  const navigate = useNavigate()
  const editorRef = useRef<TipTapEditorHandle>(null)

  const [showDiscussion, setShowDiscussion] = useState(false)
  const [chapter, setChapter] = useState<Chapter | null>(null)
  const [content, setContent] = useState('')
  const [saved, setSaved] = useState(true)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [continuing, setContinuing] = useState(false)
  const [polishing, setPolishing] = useState(false)

  useEffect(() => {
    if (!chapterId) return

    let cancelled = false
    setLoading(true)
    setError(null)

    api.get<Chapter>(`/chapters/${chapterId}`)
      .then(({ data }) => {
        if (cancelled) return
        setChapter(data)
        setContent(data.content)
        setSaved(true)
      })
      .catch(() => {
        if (cancelled) return
        setError('加载章节失败')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [chapterId])

  if (!chapterId) {
    return (
      <div className="min-h-screen bg-parchment-gradient flex items-center justify-center">
        <div className="text-center animate-fade-in">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-terracotta-light mb-4">
            <svg className="w-8 h-8 text-terracotta" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="12" cy="12" r="10" /><line x1="15" y1="9" x2="9" y2="15" /><line x1="9" y1="9" x2="15" y2="15" />
            </svg>
          </div>
          <p className="text-ink-muted font-literary">章节ID无效</p>
        </div>
      </div>
    )
  }

  const handleContentChange = (newContent: string) => {
    setContent(newContent)
    setSaved(false)
    setError(null)
  }

  const handleSave = async () => {
    if (!chapterId || saving) return
    setSaving(true)
    try {
      await api.put(`/chapters/${chapterId}`, {
        title: chapter?.title,
        content,
      })
      setSaved(true)
      setError(null)
    } catch {
      setError('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleContinue = async () => {
    if (!chapterId || continuing) return
    setContinuing(true)
    setError(null)
    try {
      const { data } = await api.post<{ content: string }>(`/chapters/${chapterId}/continue`, {
        content,
      })
      const newContent = content + data.content
      setContent(newContent)
      setSaved(false)
    } catch {
      setError('续写失败')
    } finally {
      setContinuing(false)
    }
  }

  const handlePolish = async () => {
    if (!chapterId || polishing) return
    const selected = editorRef.current?.getSelectedText() ?? ''
    if (!selected.trim()) {
      setError('请先选中要润色的文本')
      return
    }
    setPolishing(true)
    setError(null)
    try {
      const { data } = await api.post<{ content: string }>(`/chapters/${chapterId}/polish`, {
        content: selected,
      })
      editorRef.current?.replaceSelection(data.content)
      setSaved(false)
    } catch {
      setError('润色失败')
    } finally {
      setPolishing(false)
    }
  }

  const busy = saving || continuing || polishing

  if (loading) {
    return (
      <div className="min-h-screen bg-parchment-gradient flex items-center justify-center">
        <div className="animate-pulse text-warm-gray font-literary">加载中...</div>
      </div>
    )
  }

  if (!chapter) {
    return (
      <div className="min-h-screen bg-parchment-gradient flex items-center justify-center">
        <div className="text-center animate-fade-in">
          <p className="text-ink-muted font-literary mb-4">{error || '章节不存在'}</p>
          <button
            onClick={() => navigate(-1)}
            className="px-4 py-2 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors text-sm"
          >
            返回
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-parchment-gradient">
      {/* Header */}
      <header className="bg-white/70 backdrop-blur-md border-b border-parchment-deep/50 sticky top-0 z-40">
        <div className="max-w-5xl mx-auto px-6 py-3 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate(-1)}
              className="w-8 h-8 rounded-lg bg-parchment-dark flex items-center justify-center text-ink-muted hover:text-ink hover:bg-parchment-deep transition-colors"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
            </button>
            <div>
              <h1 className="text-lg font-serif font-semibold text-ink">
                {chapter ? chapter.title : '章节编辑'}
              </h1>
              <div className="flex items-center gap-2 text-xs text-warm-gray font-literary">
                <span>{content.length} 字</span>
                <span className="text-parchment-deep">|</span>
                <span className={saved ? 'text-sage' : 'text-amber'}>
                  {saved ? '已保存' : '未保存'}
                </span>
                {error && (
                  <>
                    <span className="text-parchment-deep">|</span>
                    <span className="text-terracotta">{error}</span>
                  </>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleSave}
              disabled={busy || saved}
              className="flex items-center gap-2 px-4 py-2 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors shadow-md shadow-ink/10 disabled:opacity-50 text-sm font-medium"
            >
              {saving ? (
                <>
                  <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  保存中...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                    <polyline points="17 21 17 13 7 13 7 21" />
                    <polyline points="7 3 7 8 15 8" />
                  </svg>
                  保存
                </>
              )}
            </button>
            <button
              onClick={handleContinue}
              disabled={busy}
              className="flex items-center gap-2 px-4 py-2 bg-amber text-white rounded-lg hover:bg-amber-dark transition-colors shadow-md shadow-amber/20 disabled:opacity-50 text-sm font-medium"
            >
              {continuing ? (
                <>
                  <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  续写中...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 5v14M5 12h14" />
                  </svg>
                  续写
                </>
              )}
            </button>
            <button
              onClick={handlePolish}
              disabled={busy}
              className="flex items-center gap-2 px-4 py-2 bg-white border border-parchment-deep/30 text-ink rounded-lg hover:bg-parchment-dark transition-colors disabled:opacity-50 text-sm font-medium"
            >
              {polishing ? (
                <>
                  <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  润色中...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                  润色
                </>
              )}
            </button>
            <button
              onClick={() => setShowDiscussion(true)}
              disabled={busy}
              className="flex items-center gap-2 px-4 py-2 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors shadow-md shadow-ink/10 group disabled:opacity-50"
            >
              <svg className="w-4 h-4 transition-transform group-hover:scale-110" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              </svg>
              <span className="text-sm font-medium">审稿讨论</span>
            </button>
          </div>
        </div>
      </header>

      {/* Editor area */}
      <main className="max-w-5xl mx-auto px-6 py-8">
        <div className="animate-fade-in">
          <TipTapEditor
            ref={editorRef}
            content={content}
            onChange={handleContentChange}
            placeholder="开始写作..."
          />
        </div>
      </main>

      {/* Discussion Panel */}
      {showDiscussion && (
        <DiscussionPanel
          chapterId={chapterId}
          onClose={() => setShowDiscussion(false)}
        />
      )}
    </div>
  )
}
