import { useDiscussionStore } from '../stores/discussionStore'

export default function DiscussionPanel({ chapterId, onClose }: { chapterId: string; onClose: () => void }) {
  const { result, isDiscussing, startDiscussion } = useDiscussionStore()

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[80vh] overflow-y-auto p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">审稿讨论</h2>
          <button onClick={onClose} className="text-gray-500">关闭</button>
        </div>

        {!result && !isDiscussing && (
          <button onClick={() => startDiscussion(chapterId)} className="px-4 py-2 bg-blue-600 text-white rounded-md">开始审稿</button>
        )}

        {isDiscussing && <div className="text-center py-8">正在审稿中，请稍候...</div>}

        {result && (
          <div className="space-y-6">
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="font-semibold mb-2">编辑建议</h3>
              {result.aggregated?.map((s, i) => (
                <div key={i} className="mb-2 p-2 bg-white rounded">
                  <span className="text-xs px-2 py-1 bg-blue-100 text-blue-800 rounded">{s.type}</span>
                  <span className="text-xs px-2 py-1 ml-2 bg-gray-100 rounded">优先级 {s.priority}</span>
                  <p className="mt-1 text-sm">{s.problem}</p>
                  <p className="text-sm text-gray-600">建议：{s.suggestion}</p>
                </div>
              ))}
            </div>
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="font-semibold mb-2">读者反馈</h3>
              <p className="text-sm whitespace-pre-wrap">{result.reader_feedback}</p>
            </div>
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="font-semibold mb-2">评论家分析</h3>
              <p className="text-sm whitespace-pre-wrap">{result.critic_analysis}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
