import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import DiscussionPanel from '../components/DiscussionPanel'
import TipTapEditor from '../components/TipTapEditor'

export default function ChapterEditor() {
  const { chapterId } = useParams<{ chapterId: string }>()
  const navigate = useNavigate()
  const [showDiscussion, setShowDiscussion] = useState(false)
  const [content, setContent] = useState('')

  if (!chapterId) {
    return <div className="p-8 text-center text-gray-500">章节ID无效</div>
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow px-4 py-3">
        <div className="max-w-6xl mx-auto flex justify-between items-center">
          <h1 className="text-xl font-semibold">章节编辑器</h1>
          <div className="flex space-x-4">
            <button
              onClick={() => setShowDiscussion(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              审稿讨论
            </button>
            <button onClick={() => navigate(-1)} className="text-gray-500">返回</button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="bg-white rounded-lg shadow p-6">
          <TipTapEditor
            content={content}
            onChange={setContent}
            placeholder="开始写作..."
          />
          <div className="mt-4 text-sm text-gray-500">
            字数: {content.length}
          </div>
        </div>
      </main>

      {showDiscussion && (
        <DiscussionPanel
          chapterId={chapterId}
          onClose={() => setShowDiscussion(false)}
        />
      )}
    </div>
  )
}
