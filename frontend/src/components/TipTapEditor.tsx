import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import { useEffect } from 'react'

interface TipTapEditorProps {
  content: string
  onChange: (content: string) => void
  placeholder?: string
}

export default function TipTapEditor({ content, onChange, placeholder = '开始写作...' }: TipTapEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
    ],
    content,
    onUpdate: ({ editor }) => {
      onChange(editor.getText())
    },
  })

  useEffect(() => {
    if (editor && content !== editor.getText()) {
      editor.commands.setContent(content)
    }
  }, [content, editor])

  if (!editor) return null

  const tools = [
    { icon: 'B', action: () => editor.chain().focus().toggleBold().run(), active: editor.isActive('bold'), title: '粗体' },
    { icon: 'I', action: () => editor.chain().focus().toggleItalic().run(), active: editor.isActive('italic'), title: '斜体', italic: true },
    { divider: true },
    { label: 'H2', action: () => editor.chain().focus().toggleHeading({ level: 2 }).run(), active: editor.isActive('heading', { level: 2 }), title: '标题2' },
    { label: 'H3', action: () => editor.chain().focus().toggleHeading({ level: 3 }).run(), active: editor.isActive('heading', { level: 3 }), title: '标题3' },
    { label: 'P', action: () => editor.chain().focus().setParagraph().run(), active: editor.isActive('paragraph'), title: '正文' },
    { divider: true },
    { icon: '—', action: () => editor.chain().focus().setHorizontalRule().run(), title: '分割线' },
    { icon: '"', action: () => editor.chain().focus().toggleBlockquote().run(), active: editor.isActive('blockquote'), title: '引用' },
  ]

  return (
    <div className="bg-white rounded-xl border border-parchment-deep/30 shadow-lg shadow-ink/5 overflow-hidden">
      {/* Toolbar */}
      <div className="border-b border-parchment-deep/30 bg-parchment-dark/50 px-3 py-2 flex items-center gap-1">
        {tools.map((tool, i) =>
          tool.divider ? (
            <div key={i} className="w-px h-5 bg-parchment-deep mx-1" />
          ) : (
            <button
              key={i}
              onClick={tool.action}
              title={tool.title}
              className={`w-8 h-8 rounded-md flex items-center justify-center text-sm font-medium transition-all duration-150 ${
                tool.active
                  ? 'bg-ink text-parchment shadow-sm'
                  : 'text-ink-muted hover:text-ink hover:bg-parchment-deep/50'
              } ${tool.italic ? 'italic' : ''}`}
            >
              {tool.icon || tool.label}
            </button>
          )
        )}
      </div>

      {/* Editor content */}
      <EditorContent
        editor={editor}
        className="prose prose-lg max-w-none p-8 min-h-[500px] font-literary leading-loose focus:outline-none
          [&_.tiptap]:focus:outline-none
          [&_.tiptap]:min-h-[480px]
          [&_.tiptap_p.is-editor-empty:first-child::before]:text-warm-gray
          [&_.tiptap_p.is-editor-empty:first-child::before]:font-literary
          [&_.tiptap_p.is-editor-empty:first-child::before]:content-[attr(data-placeholder)]
          [&_.tiptap_p.is-editor-empty:first-child::before]:float-left
          [&_.tiptap_p.is-editor-empty:first-child::before]:pointer-events-none
          [&_.tiptap_p.is-editor-empty:first-child::before]:h-0"
      />
    </div>
  )
}
