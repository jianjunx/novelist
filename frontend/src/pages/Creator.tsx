import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAgentStore } from '../stores/agentStore'
import { useProjectStore } from '../stores/projectStore'

export default function Creator() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const { messages, isStreaming, sendMessage, clearMessages } = useAgentStore()
  const { currentProject, fetchProjects } = useProjectStore()
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => { if (projectId) fetchProjects(); return () => clearMessages() }, [projectId])
  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [messages])

  const handleSend = async () => {
    if (!input.trim() || !projectId) return
    const msg = input; setInput(''); await sendMessage(projectId, msg)
  }

  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <header className="bg-white shadow px-4 py-3"><div className="max-w-4xl mx-auto flex justify-between items-center">
        <h1 className="text-xl font-semibold">构思对话 - {currentProject?.title || '新项目'}</h1>
        <button onClick={() => navigate('/')} className="text-gray-500">返回</button>
      </div></header>
      <div className="flex-1 max-w-4xl mx-auto w-full p-4 flex flex-col">
        <div className="flex-1 overflow-y-auto mb-4 space-y-4">
          {messages.length === 0 && <div className="text-center text-gray-500 py-12"><p className="text-lg mb-2">你好！请告诉我你想写什么类型的小说？</p><p>有什么初步的想法或灵感吗？</p></div>}
          {messages.map((msg, i) => (
            <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div className={`max-w-[80%] p-4 rounded-lg ${msg.role === 'user' ? 'bg-blue-600 text-white' : 'bg-white shadow'}`}>
                <div className="whitespace-pre-wrap">{msg.content}</div>
              </div>
            </div>
          ))}
          {isStreaming && <div className="flex justify-start"><div className="max-w-[80%] p-4 rounded-lg bg-white shadow"><div className="whitespace-pre-wrap">思考中...</div></div></div>}
          <div ref={messagesEndRef} />
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="flex space-x-4">
            <textarea value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() } }} placeholder="输入你的想法..." className="flex-1 px-3 py-2 border rounded-md resize-none" rows={2} />
            <button onClick={handleSend} disabled={isStreaming || !input.trim()} className="px-6 py-2 bg-blue-600 text-white rounded-md disabled:opacity-50 self-end">发送</button>
          </div>
        </div>
      </div>
    </div>
  )
}
