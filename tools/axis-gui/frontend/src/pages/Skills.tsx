import { useEffect, useState } from 'react'

interface SkillInfo {
  name: string
  description?: string
  tags?: string[]
  version?: string
  depends_on?: string[]
  conflicts_with?: string[]
}

export default function Skills() {
  const [skills, setSkills] = useState<SkillInfo[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/skills')
      .then(r => r.json())
      .then(d => setSkills(Array.isArray(d) ? d : []))
      .catch(() => setSkills([]))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="text-xs text-zinc-500">加载中...</div>

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="font-title text-lg font-semibold">Skills</h2>
        <span className="text-xs text-zinc-500">{skills.length} installed</span>
      </div>

      {skills.length === 0 ? (
        <div className="text-sm text-zinc-500">
          No skills found. Create one with: <code className="text-zinc-400">axis skills create &lt;name&gt;</code>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3">
          {skills.map(s => (
            <div key={s.name} className="rounded border border-zinc-800 bg-zinc-900 p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-sm font-medium text-zinc-200">{s.name}</span>
                {s.version && <span className="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400">v{s.version}</span>}
              </div>
              {s.description && <p className="text-xs text-zinc-400 mb-2">{s.description}</p>}
              <div className="flex flex-wrap gap-1 mb-2">
                {s.tags?.map(tag => (
                  <span key={tag} className="text-[10px] px-1.5 py-0.5 rounded bg-blue-950/50 text-blue-400 border border-blue-900/50">{tag}</span>
                ))}
              </div>
              {(s.depends_on?.length || s.conflicts_with?.length) ? (
                <div className="text-[11px] text-zinc-500 space-y-0.5">
                  {s.depends_on?.length ? <div>depends: <span className="text-zinc-400">{s.depends_on.join(', ')}</span></div> : null}
                  {s.conflicts_with?.length ? <div>conflicts: <span className="text-amber-400">{s.conflicts_with.join(', ')}</span></div> : null}
                </div>
              ) : null}
              <div className="mt-2 pt-2 border-t border-zinc-800">
                <code className="text-[10px] text-zinc-500">axis skills show {s.name}</code>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="mt-4 p-3 rounded border border-zinc-800 bg-zinc-900/50">
        <div className="text-xs text-zinc-500 space-y-1">
          <div><code className="text-zinc-400">axis skills list</code> — 列出所有 skills</div>
          <div><code className="text-zinc-400">axis skills show &lt;name&gt;</code> — 查看 skill 详情</div>
          <div><code className="text-zinc-400">axis skills validate</code> — 验证 skill 格式</div>
          <div><code className="text-zinc-400">axis skills create &lt;name&gt;</code> — 创建新 skill</div>
        </div>
      </div>
    </div>
  )
}
