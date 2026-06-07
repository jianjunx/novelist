import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useProjectStore } from '../stores/projectStore'
import type { ReviewResult } from '../stores/projectStore'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

export default function ChapterList() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const {
    currentProject, chapters, volumes, isLoading, isGenerating, isReviewing, isExpanding,
    reviewResult, fetchProject, fetchChapters, fetchVolumes, generateChapter,
    reviewAndRevise, expandOutlines, createVolume, clearReviewResult,
  } = useProjectStore()
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [expandedVolumes, setExpandedVolumes] = useState<Set<string>>(new Set())
  const [volumeComplete, setVolumeComplete] = useState(false)
  const [showReview, setShowReview] = useState(false)
  const [showRename, setShowRename] = useState(false)
  const [newTitle, setNewTitle] = useState('')
  const [renaming, setRenaming] = useState(false)
  const updateProject = useProjectStore((s) => s.updateProject)

  useEffect(() => {
    if (projectId) {
      fetchProject(projectId)
      fetchChapters(projectId)
      fetchVolumes(projectId)
    }
    return () => clearReviewResult()
  }, [projectId])

  // Auto-expand the latest volume
  useEffect(() => {
    if (volumes.length > 0 && expandedVolumes.size === 0) {
      const latest = volumes[volumes.length - 1]
      setExpandedVolumes(new Set([latest.id]))
    }
  }, [volumes])

  // Auto-select first chapter when chapters load
  useEffect(() => {
    if (chapters.length > 0 && !selectedId) {
      setSelectedId(chapters[0].id)
    }
  }, [chapters])

  // Show review panel when result arrives
  useEffect(() => {
    if (reviewResult) setShowReview(true)
  }, [reviewResult])

  const selected = chapters.find(c => c.id === selectedId)
  const selectedHasContent = selected?.content && selected.content.length > 0
  const busy = isGenerating || isReviewing

  // Group chapters by volume
  const chaptersByVolume = chapters.reduce<Record<string, typeof chapters>>((acc, ch) => {
    const key = ch.volume_id || '__default'
    if (!acc[key]) acc[key] = []
    acc[key].push(ch)
    return acc
  }, {})

  const toggleVolume = (volId: string) => {
    setExpandedVolumes(prev => {
      const next = new Set(prev)
      if (next.has(volId)) next.delete(volId)
      else next.add(volId)
      return next
    })
  }

  const handleNewVolume = async () => {
    if (!projectId) return
    await createVolume(projectId)
    await fetchVolumes(projectId)
  }

  const latestVolume = volumes[volumes.length - 1]
  const latestVolumeChapters = latestVolume ? (chaptersByVolume[latestVolume.id] || []) : []
  // Use backend-reported volume_complete, or fall back to chapter count heuristic
  const latestVolumeComplete = volumeComplete || latestVolumeChapters.length >= 6

  const handleGenerate = async (chapterId: string) => {
    setShowReview(false)
    await generateChapter(chapterId)
  }

  const handleReview = async (chapterId: string) => {
    setShowReview(false)
    await reviewAndRevise(chapterId)
  }

  const handleRename = async () => {
    if (!currentProject || !newTitle.trim() || newTitle === currentProject.title) return
    setRenaming(true)
    try {
      await updateProject(currentProject.short_id || currentProject.id, { title: newTitle.trim() })
      setShowRename(false)
    } finally {
      setRenaming(false)
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-parchment-gradient flex items-center justify-center">
        <div className="animate-pulse text-warm-gray font-literary">加载中...</div>
      </div>
    )
  }

  return (
    <div className="h-screen flex flex-col bg-parchment-gradient">
      {/* Header */}
      <header className="bg-white/70 backdrop-blur-md border-b border-parchment-deep/50 shrink-0 z-40">
        <div className="max-w-[1400px] mx-auto px-6 py-3 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate('/')}
              className="w-8 h-8 rounded-lg bg-parchment-dark flex items-center justify-center text-ink-muted hover:text-ink hover:bg-parchment-deep transition-colors"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
            </button>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-lg font-serif font-semibold text-ink">{currentProject?.title || '项目'}</h1>
                <button
                  onClick={() => { setNewTitle(currentProject?.title || ''); setShowRename(true) }}
                  className="w-6 h-6 rounded flex items-center justify-center text-ink-muted hover:text-amber hover:bg-amber/10 transition-colors"
                  title="修改名称"
                >
                  <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                </button>
              </div>
              <p className="text-xs text-warm-gray font-literary">章节管理</p>
            </div>
          </div>
          <button
            onClick={() => navigate(`/projects/${projectId}/creator`)}
            className="flex items-center gap-2 px-4 py-2 text-amber-dark hover:bg-amber/5 rounded-lg transition-colors text-sm font-literary"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
            </svg>
            继续构思
          </button>
        </div>
      </header>

      {chapters.length === 0 ? (
        <div className="flex-1 flex items-center justify-center animate-fade-in">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-parchment-dark mb-6">
              <svg className="w-10 h-10 text-warm-gray-light" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
              </svg>
            </div>
            <h2 className="text-xl font-serif text-ink mb-2">还没有章节</h2>
            <p className="text-warm-gray font-literary mb-6">先完成构思，章节大纲会自动生成</p>
            <button
              onClick={() => navigate(`/projects/${projectId}/creator`)}
              className="px-6 py-3 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors"
            >
              开始构思
            </button>
          </div>
        </div>
      ) : (
        <div className="flex-1 flex overflow-hidden">
          {/* Left: Chapter list sidebar */}
          <aside className="w-80 border-r border-parchment-deep/30 bg-white/50 overflow-y-auto shrink-0">
            <div className="p-4 space-y-3">
              {/* Volume-grouped chapters */}
              {volumes.map((vol) => {
                const volChapters = chaptersByVolume[vol.id] || []
                const isExpanded = expandedVolumes.has(vol.id)
                const isLatest = vol.id === latestVolume?.id
                return (
                  <div key={vol.id}>
                    {/* Volume header */}
                    <button
                      onClick={() => toggleVolume(vol.id)}
                      className="w-full flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-parchment-dark transition-colors"
                    >
                      <svg
                        className={`w-3.5 h-3.5 text-warm-gray transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                        viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
                      >
                        <path d="M9 18l6-6-6-6" />
                      </svg>
                      <span className="text-sm font-serif font-semibold text-ink flex-1 text-left">{vol.title}</span>
                      <span className="text-xs text-warm-gray">{volChapters.length}章</span>
                      {isLatest && <span className="text-xs px-1.5 py-0.5 bg-amber/10 text-amber-dark rounded">当前</span>}
                    </button>

                    {/* Chapter list */}
                    {isExpanded && (
                      <div className="ml-2 mt-1 space-y-0.5 border-l-2 border-parchment-deep/20 pl-3">
                        {volChapters.map((ch) => (
                          <ChapterItem
                            key={ch.id}
                            ch={ch}
                            isActive={ch.id === selectedId}
                            onSelect={() => { setSelectedId(ch.id); clearReviewResult(); setShowReview(false) }}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}

              {/* Fallback for chapters without volume (backward compat) */}
              {chaptersByVolume['__default'] && chaptersByVolume['__default'].length > 0 && (
                <div>
                  <div className="px-3 py-2 text-sm font-serif font-semibold text-ink">未分篇</div>
                  <div className="ml-2 space-y-0.5 border-l-2 border-parchment-deep/20 pl-3">
                    {chaptersByVolume['__default'].map((ch) => (
                      <ChapterItem
                        key={ch.id}
                        ch={ch}
                        isActive={ch.id === selectedId}
                        onSelect={() => { setSelectedId(ch.id); clearReviewResult(); setShowReview(false) }}
                      />
                    ))}
                  </div>
                </div>
              )}

              {/* Expand outlines button */}
              <button
                onClick={async () => {
                  const result = await expandOutlines(projectId!)
                  if (result?.volume_complete) setVolumeComplete(true)
                }}
                disabled={isExpanding || latestVolumeComplete}
                className="w-full mt-3 px-4 py-3 border-2 border-dashed border-parchment-deep/40 rounded-lg text-warm-gray hover:text-amber-dark hover:border-amber/30 transition-all duration-200 disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {isExpanding ? (
                  <>
                    <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    <span className="text-sm">生成中...</span>
                  </>
                ) : latestVolumeComplete ? (
                  <>
                    <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M20 6L9 17l-5-5" />
                    </svg>
                    <span className="text-sm font-literary">当前篇已完成</span>
                  </>
                ) : (
                  <>
                    <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
                    </svg>
                    <span className="text-sm font-literary">扩写大纲</span>
                  </>
                )}
              </button>

              {/* New volume button */}
              {latestVolumeComplete && (
                <button
                  onClick={handleNewVolume}
                  className="w-full mt-2 px-4 py-3 border-2 border-dashed border-sage/30 rounded-lg text-sage hover:bg-sage/5 hover:border-sage/50 transition-all duration-200 flex items-center justify-center gap-2"
                >
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                  <span className="text-sm font-literary">开始新的一篇</span>
                </button>
              )}
            </div>
          </aside>

          {/* Right: Chapter detail */}
          <main className="flex-1 overflow-y-auto">
            {selected ? (
              <div className="max-w-4xl mx-auto px-8 py-8">
                {/* Chapter header + action buttons */}
                <div className="flex items-start justify-between gap-4 mb-6">
                  <div className="flex items-center gap-3 min-w-0">
                    <span className="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-amber/10 text-amber font-serif font-bold text-lg shrink-0">
                      {selected.chapter_num}
                    </span>
                    <div className="min-w-0">
                      <h2 className="text-2xl font-serif font-bold text-ink truncate">{selected.title}</h2>
                      {selectedHasContent && (
                        <span className="text-sm text-warm-gray">{selected.word_count} 字</span>
                      )}
                    </div>
                  </div>

                  {/* Action buttons */}
                  <div className="flex items-center gap-2 shrink-0">
                  {selectedHasContent ? (
                    <>
                      <button
                        onClick={() => navigate(`/chapters/${selected.id}/edit`)}
                        className="flex items-center gap-2 px-5 py-2.5 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors shadow-md shadow-ink/10"
                      >
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                        </svg>
                        编辑内容
                      </button>
                      <button
                        onClick={() => handleReview(selected.id)}
                        disabled={busy}
                        className="flex items-center gap-2 px-5 py-2.5 bg-sage text-white rounded-lg hover:bg-sage/90 transition-colors disabled:opacity-50 shadow-md shadow-sage/20"
                      >
                        {isReviewing ? (
                          <>
                            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                            </svg>
                            评审中...
                          </>
                        ) : (
                          <>
                            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                            </svg>
                            再次评审
                          </>
                        )}
                      </button>
                      <button
                        onClick={() => handleGenerate(selected.id)}
                        disabled={busy}
                        className="flex items-center gap-2 px-5 py-2.5 bg-white border border-parchment-deep/30 text-ink rounded-lg hover:bg-parchment-dark transition-colors disabled:opacity-50"
                      >
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
                        </svg>
                        重新生成
                      </button>
                    </>
                  ) : selected.can_generate ? (
                    <button
                      onClick={() => handleGenerate(selected.id)}
                      disabled={busy}
                      className="flex items-center gap-2 px-6 py-3 bg-amber text-white rounded-lg hover:bg-amber-dark transition-colors disabled:opacity-50 shadow-lg shadow-amber/20"
                    >
                      {isGenerating ? (
                        <>
                          <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                          </svg>
                          生成并评审中...
                        </>
                      ) : (
                        <>
                          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                          </svg>
                          生成内容
                        </>
                      )}
                    </button>
                  ) : (
                    <div className="flex items-center gap-2 px-5 py-2.5 text-warm-gray">
                      <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 8v4M12 16h.01" />
                      </svg>
                      <span className="text-sm">请先完成前置章节</span>
                    </div>
                  )}
                  </div>
                </div>

                {/* Outline summary */}
                {selected.outline_summary && (
                  <div className="bg-amber/5 border border-amber/10 rounded-xl p-4 mb-6">
                    <div className="flex items-center gap-2 mb-2">
                      <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M9 18l6-6-6-6" />
                      </svg>
                      <span className="text-sm font-medium text-amber-dark">大纲摘要</span>
                    </div>
                    <p className="text-sm text-ink-muted font-literary leading-relaxed">{selected.outline_summary}</p>
                  </div>
                )}

                {/* Review panel */}
                {showReview && reviewResult && reviewResult.discussion && (
                  <ReviewPanel result={reviewResult} onClose={() => setShowReview(false)} />
                )}

                {/* Chapter content */}
                <div className="bg-white rounded-xl border border-parchment-deep/30 p-6 min-h-[400px]">
                  {selectedHasContent ? (
                    <div className="prose prose-sm max-w-none font-literary leading-relaxed">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>{selected.content}</ReactMarkdown>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center h-64 text-center">
                      <div className="w-16 h-16 rounded-full bg-parchment-dark flex items-center justify-center mb-4">
                        <svg className="w-8 h-8 text-warm-gray-light" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                          <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                        </svg>
                      </div>
                      <p className="text-warm-gray font-literary">内容尚未生成</p>
                      <p className="text-xs text-warm-gray-light mt-1">点击上方按钮生成章节内容</p>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex-1 flex items-center justify-center h-full text-warm-gray font-literary">
                选择一个章节查看详情
              </div>
            )}
          </main>
        </div>
      )}

      {/* Rename modal */}
      {showRename && currentProject && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 backdrop-blur-sm"
          onClick={() => { setShowRename(false); setNewTitle('') }}
        >
          <div
            className="bg-white rounded-2xl shadow-2xl w-full max-w-md mx-4 p-6 animate-slide-up"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-lg font-serif font-semibold text-ink mb-4">修改项目名称</h2>
            <input
              type="text"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') { handleRename() } }}
              className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray text-sm mb-4"
              autoFocus
            />
            <div className="flex items-center justify-end gap-3">
              <button
                onClick={() => { setShowRename(false); setNewTitle('') }}
                className="px-4 py-2 text-sm text-ink-muted hover:text-ink transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleRename}
                disabled={!newTitle.trim() || newTitle === currentProject.title || renaming}
                className="px-4 py-2 bg-ink text-parchment rounded-lg text-sm font-medium shadow-lg shadow-ink/10 hover:bg-ink-light disabled:opacity-40 disabled:cursor-not-allowed transition-all"
              >
                {renaming ? '保存中...' : '确认'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function ChapterItem({ ch, isActive, onSelect }: { ch: { id: string; chapter_num: number; title: string; content: string; can_generate: boolean }; isActive: boolean; onSelect: () => void }) {
  const hasContent = ch.content && ch.content.length > 0
  return (
    <button
      onClick={onSelect}
      className={`w-full text-left px-3 py-2 rounded-lg transition-all duration-200 ${
        isActive ? 'bg-amber/10 border border-amber/20 shadow-sm' : 'hover:bg-parchment-dark border border-transparent'
      }`}
    >
      <div className="flex items-center gap-2">
        <span className={`inline-flex items-center justify-center w-6 h-6 rounded text-xs font-serif font-semibold shrink-0 ${
          isActive ? 'bg-amber text-white' : 'bg-parchment-dark text-ink-muted'
        }`}>
          {ch.chapter_num}
        </span>
        <span className={`text-sm truncate flex-1 ${isActive ? 'text-amber-dark font-medium' : 'text-ink'}`}>
          {ch.title}
        </span>
        {hasContent ? (
          <span className="w-1.5 h-1.5 rounded-full bg-sage shrink-0" />
        ) : ch.can_generate ? (
          <span className="w-1.5 h-1.5 rounded-full bg-amber shrink-0" />
        ) : (
          <span className="w-1.5 h-1.5 rounded-full bg-warm-gray-light shrink-0" />
        )}
      </div>
    </button>
  )
}

function ReviewPanel({ result, onClose }: { result: ReviewResult; onClose: () => void }) {
  const d = result.discussion!
  const priorityLabel = (p: number) => {
    if (p === 1) return { text: '高', color: 'bg-terracotta/10 text-terracotta' }
    if (p === 2) return { text: '中', color: 'bg-amber/10 text-amber-dark' }
    return { text: '低', color: 'bg-parchment-dark text-warm-gray' }
  }

  return (
    <div className="mb-6 bg-white rounded-xl border border-sage/20 shadow-lg shadow-sage/5 overflow-hidden animate-fade-in">
      {/* Header */}
      <div className="bg-sage/5 border-b border-sage/10 px-5 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <svg className="w-5 h-5 text-sage" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
          <span className="text-sm font-semibold text-sage">评审结果</span>
          <span className="text-xs text-warm-gray">第 {result.round_num} 轮</span>
        </div>
        <button onClick={onClose} className="text-warm-gray hover:text-ink transition-colors">
          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="p-5 space-y-5 max-h-[500px] overflow-y-auto">
        {/* Editor suggestions */}
        {d.aggregated && d.aggregated.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-ink mb-3 flex items-center gap-2">
              <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
              </svg>
              编辑建议
            </h4>
            <div className="space-y-2">
              {d.aggregated.map((s, i) => (
                <div key={i} className="flex gap-3 p-3 bg-parchment-dark/50 rounded-lg">
                  <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium shrink-0 ${priorityLabel(s.priority).color}`}>
                    {priorityLabel(s.priority).text}
                  </span>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs text-warm-gray">{s.type}</span>
                      {s.location && <span className="text-xs text-warm-gray">· {s.location}</span>}
                    </div>
                    <p className="text-sm text-ink">{s.problem}</p>
                    <p className="text-sm text-sage mt-1">→ {s.suggestion}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Reader feedback */}
        {d.reader_feedback && (
          <div>
            <h4 className="text-sm font-semibold text-ink mb-2 flex items-center gap-2">
              <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
              读者反馈
            </h4>
            <div className="text-sm text-ink-muted font-literary leading-relaxed bg-parchment-dark/30 rounded-lg p-3">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{d.reader_feedback}</ReactMarkdown>
            </div>
          </div>
        )}

        {/* Critic analysis */}
        {d.critic_analysis && (
          <div>
            <h4 className="text-sm font-semibold text-ink mb-2 flex items-center gap-2">
              <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
                <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
              </svg>
              评论家分析
            </h4>
            <div className="text-sm text-ink-muted font-literary leading-relaxed bg-parchment-dark/30 rounded-lg p-3">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{d.critic_analysis}</ReactMarkdown>
            </div>
          </div>
        )}

        {/* Revised content notice */}
        {result.revised_content && (
          <div className="flex items-center gap-2 px-4 py-3 bg-sage/5 border border-sage/10 rounded-lg text-sm text-sage">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M20 6L9 17l-5-5" />
            </svg>
            内容已根据评审意见自动修改
          </div>
        )}
      </div>
    </div>
  )
}
