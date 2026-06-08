import { useState } from 'react'
import type { WorldSetting } from '../../types/overview'

interface Props {
  settings: WorldSetting[]
  editing: boolean
  onCreate: (category: string, content: string) => Promise<void>
  onUpdate: (id: string, category: string, content: string) => Promise<void>
  onDelete: (id: string) => Promise<void>
}

export default function WorldSettingsPanel({ settings, editing, onCreate, onUpdate, onDelete }: Props) {
  const [newCategory, setNewCategory] = useState('体系')
  const [newContent, setNewContent] = useState('')
  const [editId, setEditId] = useState<string | null>(null)
  const [editCategory, setEditCategory] = useState('')
  const [editContent, setEditContent] = useState('')

  const grouped = settings.reduce<Record<string, WorldSetting[]>>((acc, s) => {
    const cat = s.category || '其他'
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(s)
    return acc
  }, {})

  const handleCreate = async () => {
    if (!newContent.trim()) return
    await onCreate(newCategory, newContent.trim())
    setNewContent('')
  }

  const startEdit = (s: WorldSetting) => {
    setEditId(s.id)
    setEditCategory(s.category)
    setEditContent(s.content)
  }

  const saveEdit = async () => {
    if (!editId) return
    await onUpdate(editId, editCategory, editContent)
    setEditId(null)
  }

  return (
    <section id="world" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <h2 className="text-xl font-serif font-semibold text-ink mb-4">世界观设定</h2>
      {Object.keys(grouped).length === 0 ? (
        <p className="text-sm text-warm-gray font-literary">暂无世界观设定</p>
      ) : (
        <div className="space-y-5">
          {Object.entries(grouped).map(([category, items]) => (
            <div key={category}>
              <h3 className="text-sm font-medium text-amber-dark mb-2">{category}</h3>
              <ul className="space-y-2">
                {items.map((s) => (
                  <li key={s.id} className="group flex gap-2 items-start">
                    {editId === s.id ? (
                      <div className="flex-1 space-y-2">
                        <input
                          value={editCategory}
                          onChange={(e) => setEditCategory(e.target.value)}
                          className="w-full px-3 py-1.5 bg-parchment-dark border border-parchment-deep rounded text-sm"
                        />
                        <textarea
                          value={editContent}
                          onChange={(e) => setEditContent(e.target.value)}
                          rows={2}
                          className="w-full px-3 py-1.5 bg-parchment-dark border border-parchment-deep rounded text-sm resize-none"
                        />
                        <div className="flex gap-2">
                          <button onClick={saveEdit} className="text-xs text-sage hover:underline">保存</button>
                          <button onClick={() => setEditId(null)} className="text-xs text-warm-gray hover:underline">取消</button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <p className="flex-1 text-sm text-ink-muted font-literary leading-relaxed">{s.content}</p>
                        {editing && (
                          <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                            <button onClick={() => startEdit(s)} className="text-xs text-amber hover:underline">编辑</button>
                            <button onClick={() => onDelete(s.id)} className="text-xs text-terracotta hover:underline">删除</button>
                          </div>
                        )}
                      </>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      )}
      {editing && (
        <div className="mt-4 pt-4 border-t border-parchment-deep/30 space-y-2">
          <div className="flex gap-2">
            <select
              value={newCategory}
              onChange={(e) => setNewCategory(e.target.value)}
              className="px-3 py-2 bg-parchment-dark border border-parchment-deep rounded-lg text-sm"
            >
              <option value="体系">体系</option>
              <option value="规则">规则</option>
              <option value="历史">历史</option>
              <option value="地点">地点</option>
              <option value="其他">其他</option>
            </select>
          </div>
          <textarea
            value={newContent}
            onChange={(e) => setNewContent(e.target.value)}
            placeholder="添加世界观设定..."
            rows={2}
            className="w-full px-3 py-2 bg-parchment-dark border border-parchment-deep rounded-lg text-sm resize-none"
          />
          <button
            onClick={handleCreate}
            disabled={!newContent.trim()}
            className="px-4 py-2 bg-ink text-parchment rounded-lg text-sm disabled:opacity-40"
          >
            添加设定
          </button>
        </div>
      )}
    </section>
  )
}
