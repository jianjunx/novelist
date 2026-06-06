import { useNavigate } from 'react-router-dom'

export default function ProjectCard({ project }: { project: any }) {
  const navigate = useNavigate()
  return (
    <div
      className="bg-white p-6 rounded-lg shadow hover:shadow-md cursor-pointer"
      onClick={() => navigate(`/projects/${project.id}`)}
    >
      <h3 className="text-xl font-semibold mb-2">{project.title}</h3>
      {project.genre && (
        <span className="inline-block bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded mb-2">
          {project.genre}
        </span>
      )}
      <p className="text-gray-600 text-sm line-clamp-3">{project.description}</p>
    </div>
  )
}
