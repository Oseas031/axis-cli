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

  if (loading) return <div className="text-xs font-mono text-zinc-500">加载中...</div>

  return (
    <div className="space-y-3">
      <header className="flex items-center justify-between border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Skills</h1>
        <span className="font-mono text-[11px] text-zinc-500">{skills.length} installed</span>
      </header>

      {skills.length === 0 ? (
        <div className="text-xs font-mono text-zinc-500 border border-zinc-800 bg-zinc-900/40 p-3">
          No skills found. Create one with: <code className="text-zinc-300">axis skills create &lt;name&gt;</code>
        </div>
      ) : (
        <div className="border border-zinc-800 bg-zinc-900/40 divide-y divide-zinc-800">
          {skills.map(s => (
            <section key={s.name} className="p-3">
              <div className="flex items-center justify-between gap-2 mb-1">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs font-medium text-zinc-100">{s.name}</span>
                  {s.version && (
                    <span className="font-mono text-[10px] px-1.5 py-0.5 rounded-sm border border-zinc-700 bg-zinc-800 text-zinc-400">
                      v{s.version}
                    </span>
                  )}
                </div>
                <code className="text-[10px] text-zinc-500">axis skills show {s.name}</code>
              </div>
              {s.description && <p className="text-xs text-zinc-400 mb-2">{s.description}</p>}
              {s.tags && s.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mb-1.5">
                  {s.tags.map(tag => (
                    <span
                      key={tag}
                      className="font-mono text-[10px] px-1.5 py-0.5 rounded-sm bg-blue-950/40 text-blue-400 border border-blue-900/50"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}
              {(s.depends_on?.length || s.conflicts_with?.length) ? (
                <div className="font-mono text-[11px] text-zinc-500 space-y-0.5">
                  {s.depends_on?.length ? <div>depends: <span className="text-zinc-300">{s.depends_on.join(', ')}</span></div> : null}
                  {s.conflicts_with?.length ? <div>conflicts: <span className="text-amber-400">{s.conflicts_with.join(', ')}</span></div> : null}
                </div>
              ) : null}
            </section>
          ))}
        </div>
      )}

      <section className="p-3 border border-zinc-800 bg-zinc-900/30">
        <div className="font-mono text-[11px] text-zinc-500 uppercase mb-1.5">CLI Reference</div>
        <div className="text-xs text-zinc-500 space-y-1">
          <div><code className="text-zinc-300">axis skills list</code> — 列出所有 skills</div>
          <div><code className="text-zinc-300">axis skills show &lt;name&gt;</code> — 查看 skill 详情</div>
          <div><code className="text-zinc-300">axis skills validate</code> — 验证 skill 格式</div>
          <div><code className="text-zinc-300">axis skills create &lt;name&gt;</code> — 创建新 skill</div>
        </div>
      </section>
    </div>
  )
}
