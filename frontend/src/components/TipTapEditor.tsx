import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import { Markdown } from 'tiptap-markdown'
import { forwardRef, useEffect, useImperativeHandle, useRef } from 'react'

export interface TipTapEditorHandle {
  getSelectedText: () => string
  replaceSelection: (text: string) => void
}

interface TipTapEditorProps {
  content: string
  onChange: (content: string) => void
  placeholder?: string
}

const TipTapEditor = forwardRef<TipTapEditorHandle, TipTapEditorProps>(function TipTapEditor(
  { content, onChange, placeholder = '开始写作...' },
  ref
) {
  // Track the last content we received from outside (not from our own onChange)
  // to avoid unnecessary round-trips through the markdown parser/serializer
  const lastExternalContent = useRef(content)

  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
      Markdown.configure({
        html: false,
        transformPastedText: true,
        transformCopiedText: true,
      }),
    ],
    content,
    onUpdate: ({ editor }) => {
      const html = editor.getHTML()
      lastExternalContent.current = html
      onChange(html)
    },
  })

  // Only sync content when it changes from an external source
  // (e.g., server load, discussion panel apply), not from our own edits.
  // This prevents the markdown round-trip from degrading content (losing blank lines).
  useEffect(() => {
    if (editor && content !== lastExternalContent.current) {
      editor.commands.setContent(content)
      lastExternalContent.current = content
    }
  }, [content, editor])

  useImperativeHandle(ref, () => ({
    getSelectedText: () => {
      if (!editor) return ''
      const { from, to } = editor.state.selection
      if (from === to) return ''
      return editor.state.doc.textBetween(from, to)
    },
    replaceSelection: (text: string) => {
      if (!editor) return
      editor.chain().focus().deleteSelection().insertContent(text).run()
    },
  }), [editor])

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
    { label: '•', action: () => editor.chain().focus().toggleBulletList().run(), active: editor.isActive('bulletList'), title: '无序列表' },
    { label: '1.', action: () => editor.chain().focus().toggleOrderedList().run(), active: editor.isActive('orderedList'), title: '有序列表' },
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
          [&_.tiptap_h2]:font-serif [&_.tiptap_h2]:text-2xl [&_.tiptap_h2]:font-bold [&_.tiptap_h2]:text-ink [&_.tiptap_h2]:mt-8 [&_.tiptap_h2]:mb-4
          [&_.tiptap_h3]:font-serif [&_.tiptap_h3]:text-xl [&_.tiptap_h3]:font-semibold [&_.tiptap_h3]:text-ink [&_.tiptap_h3]:mt-6 [&_.tiptap_h3]:mb-3
          [&_.tiptap_p]:text-ink-light [&_.tiptap_p]:leading-relaxed [&_.tiptap_p]:mb-4
          [&_.tiptap_strong]:text-ink [&_.tiptap_strong]:font-semibold
          [&_.tiptap_em]:text-ink-muted
          [&_.tiptap_blockquote]:border-l-4 [&_.tiptap_blockquote]:border-amber/40 [&_.tiptap_blockquote]:pl-4 [&_.tiptap_blockquote]:italic [&_.tiptap_blockquote]:text-ink-muted
          [&_.tiptap_hr]:border-parchment-deep/30 [&_.tiptap_hr]:my-8
          [&_.tiptap_ul]:list-disc [&_.tiptap_ul]:pl-6 [&_.tiptap_ul]:mb-4
          [&_.tiptap_ol]:list-decimal [&_.tiptap_ol]:pl-6 [&_.tiptap_ol]:mb-4
          [&_.tiptap_li]:text-ink-light [&_.tiptap_li]:mb-1
          [&_.tiptap_p.is-editor-empty:first-child::before]:text-warm-gray
          [&_.tiptap_p.is-editor-empty:first-child::before]:font-literary
          [&_.tiptap_p.is-editor-empty:first-child::before]:content-[attr(data-placeholder)]
          [&_.tiptap_p.is-editor-empty:first-child::before]:float-left
          [&_.tiptap_p.is-editor-empty:first-child::before]:pointer-events-none
          [&_.tiptap_p.is-editor-empty:first-child::before]:h-0"
      />
    </div>
  )
})

export default TipTapEditor
