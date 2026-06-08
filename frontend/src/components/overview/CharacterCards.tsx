import { useState } from 'react'
import type { Character, CharacterRelationship } from '../../types/overview'
import { parseRelationships } from '../../types/overview'

interface Props {
  characters: Character[]
  editing: boolean
  onCreate: (data: Omit<Character, 'id' | 'project_id' | 'created_at'>) => Promise<void>
  onUpdate: (id: string, data: Partial<Character>) => Promise<void>
  onDelete: (id: string) => Promise<void>
}

const emptyChar = (): Omit<Character, 'id' | 'project_id' | 'created_at'> => ({
  name: '',
  role: '',
  personality: '',
  background: '',
  appearance: '',
  relationships: [],
})

export default function CharacterCards({ characters, editing, onCreate, onUpdate, onDelete }: Props) {
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState(emptyChar())
  const [editId, setEditId] = useState<string | null>(null)
  const [editForm, setEditForm] = useState(emptyChar())

  const roleColor = (role: string) => {
    if (role.includes('主')) return 'bg-amber/10 text-amber-dark'
    if (role.includes('反')) return 'bg-terracotta/10 text-terracotta'
    return 'bg-sage/10 text-sage'
  }

  const handleCreate = async () => {
    if (!form.name.trim()) return
    await onCreate(form)
    setForm(emptyChar())
    setShowForm(false)
  }

  const startEdit = (c: Character) => {
    setEditId(c.id)
    setEditForm({
      name: c.name,
      role: c.role,
      personality: c.personality,
      background: c.background,
      appearance: c.appearance,
      relationships: parseRelationships(c.relationships),
    })
  }

  const saveEdit = async () => {
    if (!editId) return
    await onUpdate(editId, editForm)
    setEditId(null)
  }

  const renderRelationships = (rels: CharacterRelationship[]) => {
    if (rels.length === 0) return null
    return (
      <div className="mt-2 pt-2 border-t border-parchment-deep/20">
        <p className="text-xs text-warm-gray mb-1">关系</p>
        <div className="flex flex-wrap gap-1">
          {rels.map((r, i) => (
            <span key={i} className="text-xs px-2 py-0.5 bg-parchment-dark rounded-full text-ink-muted">
              {r.type} · {r.target}
            </span>
          ))}
        </div>
      </div>
    )
  }

  const charForm = (
    data: typeof form,
    setData: (d: typeof form) => void,
    onSave: () => void,
    onCancel: () => void,
  ) => (
    <div className="space-y-2 p-4 bg-parchment-dark rounded-lg">
      {(['name', 'role', 'personality', 'background', 'appearance'] as const).map((field) => (
        <input
          key={field}
          value={data[field]}
          onChange={(e) => setData({ ...data, [field]: e.target.value })}
          placeholder={{ name: '姓名', role: '角色（主角/配角/反派）', personality: '性格', background: '背景', appearance: '外貌' }[field]}
          className="w-full px-3 py-2 bg-white border border-parchment-deep rounded text-sm"
        />
      ))}
      <div className="flex gap-2">
        <button onClick={onSave} className="px-3 py-1.5 bg-ink text-parchment rounded text-sm">保存</button>
        <button onClick={onCancel} className="px-3 py-1.5 text-sm text-warm-gray">取消</button>
      </div>
    </div>
  )

  return (
    <section id="characters" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-serif font-semibold text-ink">人物档案</h2>
        {editing && !showForm && (
          <button onClick={() => setShowForm(true)} className="text-sm text-amber hover:underline">+ 添加人物</button>
        )}
      </div>
      {showForm && charForm(form, setForm, handleCreate, () => setShowForm(false))}
      {characters.length === 0 && !showForm ? (
        <p className="text-sm text-warm-gray font-literary">暂无人物设定</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
          {characters.map((c) => (
            <div key={c.id} className="group border border-parchment-deep/30 rounded-lg p-4 hover:border-amber/20 transition-colors">
              {editId === c.id ? (
                charForm(editForm, setEditForm, saveEdit, () => setEditId(null))
              ) : (
                <>
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <h3 className="font-serif font-semibold text-ink">{c.name}</h3>
                      {c.role && (
                        <span className={`inline-block mt-1 px-2 py-0.5 text-xs rounded-full ${roleColor(c.role)}`}>
                          {c.role}
                        </span>
                      )}
                    </div>
                    {editing && (
                      <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button onClick={() => startEdit(c)} className="text-xs text-amber hover:underline">编辑</button>
                        <button onClick={() => onDelete(c.id)} className="text-xs text-terracotta hover:underline">删除</button>
                      </div>
                    )}
                  </div>
                  {c.personality && <p className="text-sm text-ink-muted mt-2"><span className="text-warm-gray">性格：</span>{c.personality}</p>}
                  {c.background && <p className="text-sm text-ink-muted mt-1 font-literary"><span className="text-warm-gray">背景：</span>{c.background}</p>}
                  {c.appearance && <p className="text-sm text-ink-muted mt-1"><span className="text-warm-gray">外貌：</span>{c.appearance}</p>}
                  {renderRelationships(parseRelationships(c.relationships))}
                </>
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
