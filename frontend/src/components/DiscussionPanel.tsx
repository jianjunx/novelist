import { useState } from 'react'
import { useDiscussionStore } from '../stores/discussionStore'
import api from '../api/client'

export default function DiscussionPanel({ chapterId, onClose, onApply }: { chapterId: string; onClose: () => void; onApply?: (revisedContent: string) => void }) {
  const { result, isDiscussing, startDiscussion } = useDiscussionStore()
  const [applying, setApplying] = useState(false)

  const handleApply = async () => {
    if (!result || !onApply) return
    setApplying(true)
    try {
      const { data } = await api.post<{ revised_content: string }>(`/chapters/${chapterId}/apply-feedback`, {
        discussion: result,
      })
      onApply(data.revised_content)
      onClose()
    } catch {
      // silently fail, user can retry
    } finally {
      setApplying(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={onClose}>
      {/* Backdrop */}
      <div className="absolute inset-0 bg-ink/40 backdrop-blur-sm animate-fade-in" />

      {/* Panel */}
      <div
        className="relative bg-parchment rounded-2xl shadow-2xl shadow-ink/20 max-w-4xl w-full max-h-[85vh] overflow-hidden animate-slide-up flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-parchment-deep/30 bg-white/50">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-amber/10 border border-amber/20 flex items-center justify-center">
              <svg className="w-5 h-5 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              </svg>
            </div>
            <div>
              <h2 className="text-lg font-serif font-semibold text-ink">审稿讨论</h2>
              <p className="text-xs text-warm-gray font-literary">多位Agent将从不同角度审阅你的作品</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-lg bg-parchment-dark flex items-center justify-center text-ink-muted hover:text-ink hover:bg-parchment-deep transition-colors"
          >
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Initial state */}
          {!result && !isDiscussing && (
            <div className="text-center py-12 animate-fade-in">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-amber/10 border border-amber/20 mb-6">
                <svg className="w-10 h-10 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                  <circle cx="9" cy="7" r="4" />
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                  <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                </svg>
              </div>
              <h3 className="text-xl font-serif font-semibold text-ink mb-2">开始审稿</h3>
              <p className="text-ink-muted font-literary mb-8 max-w-md mx-auto">
                编辑、读者和评论家将从不同角度审阅你的章节，提供专业建议
              </p>
              <button
                onClick={() => startDiscussion(chapterId)}
                className="px-8 py-3 bg-ink text-parchment font-medium rounded-lg hover:bg-ink-light transition-all duration-300 shadow-lg shadow-ink/10"
              >
                开始审稿
              </button>
            </div>
          )}

          {/* Loading state */}
          {isDiscussing && (
            <div className="text-center py-16 animate-fade-in">
              <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-amber/10 border border-amber/20 mb-6 animate-pulse">
                <svg className="w-8 h-8 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                </svg>
              </div>
              <h3 className="text-lg font-serif font-semibold text-ink mb-2">正在审稿中</h3>
              <p className="text-ink-muted font-literary">Agent们正在审阅你的作品，请稍候...</p>
              <div className="flex justify-center gap-1.5 mt-4">
                <div className="w-2 h-2 bg-amber rounded-full" style={{ animation: 'dotPulse 1.4s infinite ease-in-out' }} />
                <div className="w-2 h-2 bg-amber rounded-full" style={{ animation: 'dotPulse 1.4s infinite ease-in-out 0.2s' }} />
                <div className="w-2 h-2 bg-amber rounded-full" style={{ animation: 'dotPulse 1.4s infinite ease-in-out 0.4s' }} />
              </div>
            </div>
          )}

          {/* Results */}
          {result && (
            <div className="space-y-6 animate-fade-in">
              {/* Editor suggestions */}
              {result.aggregated && result.aggregated.length > 0 && (
                <div className="bg-white rounded-xl border border-parchment-deep/30 overflow-hidden">
                  <div className="px-5 py-3 bg-parchment-dark/50 border-b border-parchment-deep/30 flex items-center gap-2">
                    <div className="w-7 h-7 rounded-md bg-amber/10 flex items-center justify-center">
                      <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                      </svg>
                    </div>
                    <h3 className="font-serif font-semibold text-ink">编辑建议</h3>
                  </div>
                  <div className="divide-y divide-parchment-deep/20">
                    {result.aggregated.map((s, i) => (
                      <div key={i} className="px-5 py-4 hover:bg-parchment-dark/30 transition-colors">
                        <div className="flex items-center gap-2 mb-2">
                          <span className="px-2 py-0.5 bg-amber/10 text-amber-dark text-xs rounded-full font-medium">{s.type}</span>
                          <span className={`px-2 py-0.5 text-xs rounded-full font-medium ${
                            s.priority === 1 ? 'bg-terracotta-light text-terracotta' :
                            s.priority === 2 ? 'bg-amber-glow text-amber' :
                            'bg-sage-light text-sage'
                          }`}>
                            {s.priority === 1 ? '高优先' : s.priority === 2 ? '中优先' : '低优先'}
                          </span>
                        </div>
                        <p className="text-sm text-ink mb-1 font-literary">{s.problem}</p>
                        <p className="text-sm text-ink-muted font-literary">
                          <span className="text-sage font-medium">建议：</span>{s.suggestion}
                        </p>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Reader feedback */}
              {result.reader_feedback && (
                <div className="bg-white rounded-xl border border-parchment-deep/30 overflow-hidden">
                  <div className="px-5 py-3 bg-parchment-dark/50 border-b border-parchment-deep/30 flex items-center gap-2">
                    <div className="w-7 h-7 rounded-md bg-sage-light flex items-center justify-center">
                      <svg className="w-4 h-4 text-sage" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                      </svg>
                    </div>
                    <h3 className="font-serif font-semibold text-ink">读者反馈</h3>
                  </div>
                  <div className="px-5 py-4">
                    <p className="text-sm text-ink-light leading-relaxed font-literary whitespace-pre-wrap">{result.reader_feedback}</p>
                  </div>
                </div>
              )}

              {/* Critic analysis */}
              {result.critic_analysis && (
                <div className="bg-white rounded-xl border border-parchment-deep/30 overflow-hidden">
                  <div className="px-5 py-3 bg-parchment-dark/50 border-b border-parchment-deep/30 flex items-center gap-2">
                    <div className="w-7 h-7 rounded-md bg-terracotta-light flex items-center justify-center">
                      <svg className="w-4 h-4 text-terracotta" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                      </svg>
                    </div>
                    <h3 className="font-serif font-semibold text-ink">评论家分析</h3>
                  </div>
                  <div className="px-5 py-4">
                    <p className="text-sm text-ink-light leading-relaxed font-literary whitespace-pre-wrap">{result.critic_analysis}</p>
                  </div>
                </div>
              )}

              {/* Apply button */}
              {onApply && (
                <div className="flex justify-end pt-2">
                  <button
                    onClick={handleApply}
                    disabled={applying}
                    className="flex items-center gap-2 px-6 py-2.5 bg-sage text-white rounded-lg hover:bg-sage/90 transition-colors disabled:opacity-50 shadow-md shadow-sage/20"
                  >
                    {applying ? (
                      <>
                        <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                        </svg>
                        <span className="text-sm font-medium">应用中...</span>
                      </>
                    ) : (
                      <>
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M20 6L9 17l-5-5" />
                        </svg>
                        <span className="text-sm font-medium">应用建议</span>
                      </>
                    )}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
