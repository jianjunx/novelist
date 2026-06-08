import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useSettingsStore } from '../stores/settingsStore'

function PasswordField({
  label,
  value,
  onChange,
  placeholder,
}: {
  label: string
  value: string
  onChange: (v: string) => void
  placeholder?: string
}) {
  const [show, setShow] = useState(false)

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-ink-light">{label}</label>
      <div className="relative">
        <input
          type={show ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className="w-full px-4 py-3 pr-12 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray transition-all duration-200"
        />
        <button
          type="button"
          onClick={() => setShow(!show)}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-warm-gray hover:text-ink transition-colors"
          tabIndex={-1}
        >
          {show ? (
            <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
              <line x1="1" y1="1" x2="23" y2="23" />
            </svg>
          ) : (
            <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          )}
        </button>
      </div>
    </div>
  )
}

export default function Settings() {
  const { loading, error, models, fetchSettings, fetchModels, updateSettings } = useSettingsStore()
  const [deepseekKey, setDeepseekKey] = useState('')
  const [claudeKey, setClaudeKey] = useState('')
  const [openaiKey, setOpenaiKey] = useState('')
  const [localModelUrl, setLocalModelUrl] = useState('')
  const [defaultModel, setDefaultModel] = useState('deepseek-chat')
  const [discussionRounds, setDiscussionRounds] = useState(1)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    fetchSettings().then(() => {
      const s = useSettingsStore.getState()
      setDeepseekKey(s.deepseekKey)
      setClaudeKey(s.claudeKey)
      setOpenaiKey(s.openaiKey)
      setLocalModelUrl(s.localModelUrl)
      setDefaultModel(s.defaultModel)
      setDiscussionRounds(s.discussionRounds)
    }).catch(() => {})
    fetchModels()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaved(false)
    try {
      await updateSettings({
        deepseekKey,
        claudeKey,
        openaiKey,
        localModelUrl,
        defaultModel,
        discussionRounds,
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    } catch {}
  }

  return (
    <div className="min-h-screen bg-parchment-gradient">
      <header className="bg-white/70 backdrop-blur-md border-b border-parchment-deep/50 sticky top-0 z-40">
        <div className="max-w-3xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <Link to="/" className="flex items-center gap-2 text-warm-gray hover:text-ink transition-colors">
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
              <span className="text-sm">返回</span>
            </Link>
          </div>
          <h1 className="text-lg font-serif font-bold text-ink tracking-wide">设置</h1>
          <div className="w-16" />
        </div>
      </header>

      <main className="max-w-3xl mx-auto px-6 py-10 animate-fade-in">
        <div className="mb-8">
          <h2 className="text-2xl font-serif font-bold text-ink mb-2">创作配置</h2>
          <p className="text-ink-muted font-literary">配置 API 密钥与模型参数，为你的创作之旅做好准备</p>
        </div>

        <div className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-lg shadow-ink/5 border border-parchment-deep/30 p-8">
          {loading && !deepseekKey && !error ? (
            <div className="space-y-5 animate-pulse">
              {[1, 2, 3, 4].map((i) => (
                <div key={i}>
                  <div className="h-4 bg-parchment-dark rounded w-1/4 mb-2" />
                  <div className="h-12 bg-parchment-dark rounded" />
                </div>
              ))}
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-6">
              {error && (
                <div className="bg-terracotta-light border border-terracotta/20 text-terracotta px-4 py-3 rounded-lg text-sm animate-fade-in">
                  {error}
                </div>
              )}
              {saved && (
                <div className="bg-sage-light border border-sage/20 text-sage px-4 py-3 rounded-lg text-sm animate-fade-in">
                  设置已保存
                </div>
              )}

              <div className="space-y-1">
                <h3 className="text-sm font-serif font-semibold text-ink flex items-center gap-2">
                  <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                  </svg>
                  API 密钥
                </h3>
                <p className="text-xs text-warm-gray font-literary mb-4">密钥仅存储在本地服务器，不会上传至第三方</p>
              </div>

              <PasswordField
                label="DeepSeek API Key"
                value={deepseekKey}
                onChange={setDeepseekKey}
                placeholder="sk-..."
              />
              <PasswordField
                label="Claude API Key"
                value={claudeKey}
                onChange={setClaudeKey}
                placeholder="sk-ant-..."
              />
              <PasswordField
                label="OpenAI API Key"
                value={openaiKey}
                onChange={setOpenaiKey}
                placeholder="sk-..."
              />

              <div className="border-t border-parchment-deep/30 pt-6 space-y-1">
                <h3 className="text-sm font-serif font-semibold text-ink flex items-center gap-2">
                  <svg className="w-4 h-4 text-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="3" />
                    <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
                  </svg>
                  模型配置
                </h3>
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-ink-light">本地模型地址</label>
                <input
                  type="text"
                  value={localModelUrl}
                  onChange={(e) => setLocalModelUrl(e.target.value)}
                  placeholder="http://localhost:11434/v1"
                  className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray transition-all duration-200"
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="block text-sm font-medium text-ink-light">默认模型</label>
                  <button
                    type="button"
                    onClick={fetchModels}
                    className="text-xs text-warm-gray hover:text-ink transition-colors"
                  >
                    刷新列表
                  </button>
                </div>
                <input
                  type="text"
                  value={defaultModel}
                  onChange={(e) => setDefaultModel(e.target.value)}
                  list="model-options"
                  placeholder="输入或选择模型名称"
                  className="w-full px-4 py-3 bg-parchment-dark border border-parchment-deep rounded-lg text-ink placeholder-warm-gray transition-all duration-200"
                />
                <datalist id="model-options">
                  {models.map((opt) => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </datalist>
                {models.length > 0 && (
                  <p className="text-xs text-warm-gray font-literary">
                    已根据你的 API Key 自动加载可用模型，也可手动输入任意模型名称
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-ink-light">讨论轮数</label>
                <div className="flex items-center gap-4">
                  <input
                    type="range"
                    min={1}
                    max={5}
                    value={discussionRounds}
                    onChange={(e) => setDiscussionRounds(Number(e.target.value))}
                    className="flex-1 accent-amber"
                  />
                  <span className="w-8 text-center text-lg font-serif font-bold text-amber">{discussionRounds}</span>
                </div>
                <p className="text-xs text-warm-gray font-literary">Agent 讨论修订的轮数，建议 1-3 轮</p>
              </div>

              <div className="pt-4">
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 bg-ink text-parchment font-medium rounded-lg hover:bg-ink-light transition-all duration-300 disabled:opacity-50 relative overflow-hidden"
                >
                  <span className={`inline-block transition-transform duration-300 ${loading ? '-translate-y-8 opacity-0' : 'translate-y-0 opacity-100'}`}>
                    保存设置
                  </span>
                  {loading && (
                    <span className="absolute inset-0 flex items-center justify-center">
                      <svg className="animate-spin h-5 w-5 text-parchment" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                      </svg>
                    </span>
                  )}
                </button>
              </div>
            </form>
          )}
        </div>

        <p className="text-center text-warm-gray text-xs mt-8 font-literary">
          工欲善其事，必先利其器
        </p>
      </main>
    </div>
  )
}
