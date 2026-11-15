import { useEffect, useState } from 'react'

interface RuntimeStatus {
  connected: boolean
  address?: string
  hint?: string
}

interface TaskEvent {
  task_id: string
  event_type: string
  status: string
  created_at: string
  message?: string
}

interface ProviderProfile {
  name: string
  type: string
  model: string
  active?: boolean
}

export default function Dashboard() {
  const [runtime, setRuntime] = useState<RuntimeStatus | null>(null)
  const [tasks, setTasks] = useState<TaskEvent[]>([])
  const [providers, setProviders] = useState<Record<string, ProviderProfile> | null>(null)

  useEffect(() => {
    fetch('/api/runtime/status')
      .then(r => r.json())
      .then(setRuntime)
      .catch(() => setRuntime({ connected: false, hint: '无法连接运行时' }))

    fetch('/api/tasks')
      .then(r => r.json())
      .then(d => setTasks((d.tasks || []).slice(0, 5)))
      .catch(() => {})

    fetch('/api/providers')
      .then(r => r.json())
      .then(setProviders)
      .catch(() => {})
  }, [])

  const activeProvider = providers
    ? Object.entries(providers).find(([, p]) => p.active)
    : null

  return (
    <div className="space-y-3">
      <header className="flex items-center justify-between border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Dashboard</h1>
        <span className="font-mono text-[11px] text-zinc-500">
          {runtime?.connected ? 'RUNTIME: ONLINE' : 'RUNTIME: OFFLINE'}
        </span>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 border border-zinc-800 bg-zinc-900/50 divide-y md:divide-y-0 md:divide-x divide-zinc-800">
        <section className="p-3">
          <div className="font-mono text-[11px] text-zinc-500 uppercase mb-1">运行时状态</div>
          <div className="flex items-center gap-2">
            <span
              aria-hidden="true"
              className={`w-2 h-2 rounded-sm ${runtime?.connected ? 'bg-emerald-500' : 'bg-red-500'}`}
            />
            <span className="font-mono text-xs font-medium text-zinc-100">
              {runtime?.connected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
          {runtime?.address && <div className="text-[11px] text-zinc-400 mt-1 font-mono">{runtime.address}</div>}
          {runtime?.hint && <div className="text-[11px] text-zinc-500 mt-1">{runtime.hint}</div>}
        </section>

        <section className="p-3">
          <div className="font-mono text-[11px] text-zinc-500 uppercase mb-1">当前 Provider</div>
          {activeProvider ? (
            <div>
              <div className="font-mono text-xs font-medium text-zinc-100">{activeProvider[0]}</div>
              <div className="font-mono text-[11px] text-zinc-400 mt-1">
                {activeProvider[1].type} / {activeProvider[1].model}
              </div>
            </div>
          ) : (
            <div className="text-xs text-zinc-500">无活跃 Provider</div>
          )}
        </section>

        <section className="p-3">
          <div className="font-mono text-[11px] text-zinc-500 uppercase mb-1">最近任务</div>
          <div className="font-mono text-xs font-medium text-zinc-100">{tasks.length} 条事件</div>
        </section>
      </div>

      <section className="border border-zinc-800 bg-zinc-900/40">
        <div className="px-3 py-1.5 border-b border-zinc-800 font-mono text-[11px] text-zinc-400 uppercase">
          最近任务事件
        </div>
        {tasks.length === 0 ? (
          <div className="px-3 py-3 text-xs text-zinc-600 font-mono">暂无任务</div>
        ) : (
          <ul className="divide-y divide-zinc-800/60">
            {tasks.map((t, i) => (
              <li key={i} className="px-3 py-1.5 flex items-center gap-3 text-xs">
                <span className="font-mono text-zinc-300 w-36 truncate">{t.task_id}</span>
                <span className="font-mono text-zinc-500 flex-1 truncate">{t.event_type}</span>
                <span className={`px-1.5 py-0.5 rounded-sm font-mono text-[10px] border ${
                  t.status === 'completed'
                    ? 'border-emerald-800/70 bg-emerald-950/40 text-emerald-400'
                    : t.status === 'failed'
                      ? 'border-red-800/70 bg-red-950/40 text-red-400'
                      : 'border-zinc-700 bg-zinc-800 text-zinc-400'
                }`}>
                  {t.status}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
