import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import DiscussionPanel from '../components/DiscussionPanel'
import TipTapEditor from '../components/TipTapEditor'

export default function ChapterEditor() {
  const { chapterId } = useParams<{ chapterId: string }>()
  const navigate = useNavigate()
  const [showDiscussion, setShowDiscussion] = useState(false)
  const [content, setContent] = useState('')
  const [saved, setSaved] = useState(true)

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
              <h1 className="text-lg font-serif font-semibold text-ink">章节编辑</h1>
              <div className="flex items-center gap-2 text-xs text-warm-gray font-literary">
                <span>{content.length} 字</span>
                <span className="text-parchment-deep">|</span>
                <span className={saved ? 'text-sage' : 'text-amber'}>
                  {saved ? '已保存' : '未保存'}
                </span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => setShowDiscussion(true)}
              className="flex items-center gap-2 px-4 py-2 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-colors shadow-md shadow-ink/10 group"
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
