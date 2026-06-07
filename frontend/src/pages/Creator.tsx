import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAgentStore } from '../stores/agentStore'
import { useProjectStore } from '../stores/projectStore'

export default function Creator() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const { messages, isStreaming, sendMessage, loadConversations, clearMessages, savedIDs } = useAgentStore()
  const { currentProject, fetchProject } = useProjectStore()
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (projectId) {
      fetchProject(projectId)
      loadConversations(projectId)
    }
    return () => clearMessages()
  }, [projectId])
  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [messages, isStreaming])

  // Auto-navigate to chapter list when brainstorm is saved
  useEffect(() => {
    if (savedIDs?.chapter_ids && savedIDs.chapter_ids.length > 0) {
      const timer = setTimeout(() => {
        navigate(`/projects/${projectId}/chapters`)
      }, 2000)
      return () => clearTimeout(timer)
    }
  }, [savedIDs, navigate, projectId])

  const handleSend = async () => {
    if (!input.trim() || !projectId) return
    const msg = input
    setInput('')
    await sendMessage(projectId, msg)
  }

  const handleOptionClick = async (option: string) => {
    if (!projectId) return
    await sendMessage(projectId, option)
  }

  return (
    <div className="min-h-screen flex flex-col bg-parchment-gradient">
      {/* Header */}
      <header className="bg-white/70 backdrop-blur-md border-b border-parchment-deep/50 sticky top-0 z-40">
        <div className="max-w-4xl mx-auto px-6 py-3 flex justify-between items-center">
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
              <h1 className="text-lg font-serif font-semibold text-ink">{currentProject?.title || '新项目'}</h1>
              <p className="text-xs text-warm-gray font-literary">构思阶段</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {savedIDs?.chapter_ids && savedIDs.chapter_ids.length > 0 && (
              <div className="flex items-center gap-2 px-4 py-2 bg-sage/10 text-sage border border-sage/20 rounded-lg animate-fade-in">
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M20 6L9 17l-5-5" />
                </svg>
                <span className="text-sm font-medium">构思已保存，正在跳转...</span>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Messages area */}
      <div className="flex-1 max-w-4xl mx-auto w-full px-6 py-6 flex flex-col pb-32">
        <div className="flex-1 overflow-y-auto space-y-6">
          {/* Empty state */}
          {messages.length === 0 && (
            <div className="text-center py-16 animate-fade-in">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-amber/10 border border-amber/20 mb-6">
                <svg className="w-10 h-10 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                </svg>
              </div>
              <h2 className="text-2xl font-serif font-semibold text-ink mb-3">开始构思你的小说</h2>
              <p className="text-ink-muted font-literary max-w-md mx-auto leading-relaxed">
                告诉我你想写什么类型的小说，有什么初步的想法或灵感？我会引导你完成世界观、人物和大纲的构建。
              </p>
            </div>
          )}

          {/* Messages */}
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'} animate-fade-in`}
            >
              {msg.role === 'agent' && (
                <div className="w-8 h-8 rounded-full bg-amber/10 border border-amber/20 flex items-center justify-center mr-3 mt-1 flex-shrink-0">
                  <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                </div>
              )}
              <div className={`max-w-[75%] ${msg.role === 'user' ? 'order-1' : ''}`}>
                <div
                  className={`p-4 rounded-2xl ${
                    msg.role === 'user'
                      ? 'bg-ink text-parchment rounded-br-md'
                      : 'bg-white border border-parchment-deep/30 shadow-sm rounded-bl-md'
                  }`}
                >
                  {msg.role === 'user' ? (
                    <div className="whitespace-pre-wrap text-sm leading-relaxed">{msg.content}</div>
                  ) : (
                    <div className="prose prose-sm max-w-none">
                      {msg.content ? (
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>{msg.content}</ReactMarkdown>
                      ) : isStreaming && i === messages.length - 1 ? (
                        <div className="flex items-center gap-1.5 py-1">
                          <div className="w-2 h-2 bg-amber rounded-full" style={{ animation: 'dotPulse 1.4s infinite ease-in-out' }} />
                          <div className="w-2 h-2 bg-amber rounded-full" style={{ animation: 'dotPulse 1.4s infinite ease-in-out 0.2s' }} />
                          <div className="w-2 h-2 bg-amber rounded-full" style={{ animation: 'dotPulse 1.4s infinite ease-in-out 0.4s' }} />
                        </div>
                      ) : null}
                      {isStreaming && i === messages.length - 1 && msg.content && (
                        <span className="inline-block w-0.5 h-4 bg-amber ml-0.5 align-text-bottom animate-pulse" />
                      )}
                    </div>
                  )}
                </div>
                {/* Options */}
                {msg.options && msg.options.length > 0 && i === messages.length - 1 && !isStreaming && (
                  <div className="mt-3 flex flex-wrap gap-2">
                    {msg.options.map((opt, j) => (
                      <button
                        key={j}
                        onClick={() => handleOptionClick(opt)}
                        className="px-4 py-2 bg-white border border-amber/30 text-amber-dark rounded-full text-sm font-literary hover:bg-amber/5 hover:border-amber/50 transition-all duration-200 shadow-sm hover:shadow-md"
                      >
                        {opt}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}

          <div ref={messagesEndRef} />
        </div>

      </div>

      {/* Floating input area */}
      <div className="fixed bottom-0 left-0 right-0 z-30">
        <div className="max-w-4xl mx-auto px-6 pb-4">
          <div className="bg-white/90 backdrop-blur-md rounded-xl border border-parchment-deep/30 shadow-xl shadow-ink/10 p-4">
            <div className="flex gap-3">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() } }}
                placeholder="输入你的想法..."
                className="flex-1 px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray resize-none transition-all duration-200"
                rows={2}
              />
              <button
                onClick={handleSend}
                disabled={isStreaming || !input.trim()}
                className="px-5 py-3 bg-ink text-parchment rounded-lg hover:bg-ink-light transition-all duration-300 disabled:opacity-30 self-end group"
              >
                <svg className="w-5 h-5 transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="22" y1="2" x2="11" y2="13" />
                  <polygon points="22 2 15 22 11 13 2 9 22 2" />
                </svg>
              </button>
            </div>
            <p className="text-xs text-warm-gray mt-2 font-literary">按 Enter 发送，Shift+Enter 换行</p>
          </div>
        </div>
      </div>
    </div>
  )
}
