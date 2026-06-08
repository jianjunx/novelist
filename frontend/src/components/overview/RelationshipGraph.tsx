import { useMemo } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  MarkerType,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { useNavigate, useParams } from 'react-router-dom'
import type { Character } from '../../types/overview'
import { parseRelationships } from '../../types/overview'

interface Props {
  characters: Character[]
}

function CharacterNode({ data }: { data: { label: string; role: string } }) {
  const roleColor = data.role.includes('主') ? 'border-amber/50 bg-amber/5'
    : data.role.includes('反') ? 'border-terracotta/50 bg-terracotta/5'
    : 'border-sage/50 bg-sage/5'

  return (
    <div className={`px-4 py-3 rounded-lg border-2 ${roleColor} min-w-[100px] text-center shadow-sm`}>
      <div className="font-serif font-semibold text-ink text-sm">{data.label}</div>
      {data.role && <div className="text-xs text-warm-gray mt-0.5">{data.role}</div>}
    </div>
  )
}

const nodeTypes = { character: CharacterNode }

export default function RelationshipGraph({ characters }: Props) {
  const navigate = useNavigate()
  const { projectId } = useParams<{ projectId: string }>()

  const { nodes, edges } = useMemo(() => {
    const n: Node[] = []
    const e: Edge[] = []
    const nameToId = new Map<string, string>()

    characters.forEach((c, i) => {
      const angle = (2 * Math.PI * i) / Math.max(characters.length, 1)
      const radius = 180
      const id = c.id
      nameToId.set(c.name, id)
      n.push({
        id,
        type: 'character',
        position: {
          x: 250 + radius * Math.cos(angle),
          y: 200 + radius * Math.sin(angle),
        },
        data: { label: c.name, role: c.role },
      })
    })

    characters.forEach((c) => {
      const rels = parseRelationships(c.relationships)
      for (const rel of rels) {
        const targetId = nameToId.get(rel.target)
        if (!targetId || targetId === c.id) continue
        const edgeId = `${c.id}-${targetId}-${rel.type}`
        if (e.some((edge) => edge.id === edgeId)) continue
        e.push({
          id: edgeId,
          source: c.id,
          target: targetId,
          label: rel.type,
          type: 'default',
          markerEnd: { type: MarkerType.ArrowClosed, width: 16, height: 16 },
          style: { stroke: '#b8860b', strokeWidth: 1.5 },
          labelStyle: { fill: '#6b6b6b', fontSize: 11 },
        })
      }
    })

    return { nodes: n, edges: e }
  }, [characters])

  const hasRelationships = edges.length > 0

  return (
    <section id="relationships" className="bg-white rounded-xl border border-parchment-deep/30 p-6">
      <h2 className="text-xl font-serif font-semibold text-ink mb-4">人物关系图</h2>
      {characters.length === 0 ? (
        <p className="text-sm text-warm-gray font-literary">暂无人物，无法绘制关系图</p>
      ) : !hasRelationships ? (
        <div className="text-center py-8">
          <p className="text-sm text-warm-gray font-literary mb-3">暂无人物关系数据</p>
          <button
            onClick={() => navigate(`/projects/${projectId}/creator`)}
            className="text-sm text-amber hover:underline"
          >
            前往构思补充人物关系
          </button>
        </div>
      ) : (
        <div className="h-[400px] rounded-lg border border-parchment-deep/30 overflow-hidden">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            fitView
            nodesDraggable
            nodesConnectable={false}
            elementsSelectable
            proOptions={{ hideAttribution: true }}
          >
            <Background color="#e8e2d8" gap={16} />
            <Controls showInteractive={false} />
            <MiniMap nodeColor="#b8860b" maskColor="rgba(250,248,245,0.8)" />
          </ReactFlow>
        </div>
      )}
    </section>
  )
}
