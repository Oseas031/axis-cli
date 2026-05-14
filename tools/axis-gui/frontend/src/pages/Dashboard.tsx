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
    <div className="space-y-4">
      <h2 className="font-title text-lg font-semibold">Dashboard</h2>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        {/* Runtime Status */}
        <div className="rounded border border-zinc-800 bg-zinc-900 p-3">
          <div className="text-xs text-zinc-500 mb-1">运行时状态</div>
          <div className="flex items-center gap-2">
            <span className={`w-2 h-2 rounded-full ${runtime?.connected ? 'bg-emerald-500' : 'bg-red-500'}`} />
            <span className="text-sm font-medium">{runtime?.connected ? 'Connected' : 'Disconnected'}</span>
          </div>
          {runtime?.address && <div className="text-xs text-zinc-500 mt-1 font-mono">{runtime.address}</div>}
          {runtime?.hint && <div className="text-xs text-zinc-400 mt-1">{runtime.hint}</div>}
        </div>

        {/* Active Provider */}
        <div className="rounded border border-zinc-800 bg-zinc-900 p-3">
          <div className="text-xs text-zinc-500 mb-1">当前 Provider</div>
          {activeProvider ? (
            <div>
              <div className="text-sm font-medium">{activeProvider[0]}</div>
              <div className="text-xs text-zinc-500">{activeProvider[1].type} / {activeProvider[1].model}</div>
            </div>
          ) : (
            <div className="text-sm text-zinc-500">无活跃 Provider</div>
          )}
        </div>

        {/* Task Count */}
        <div className="rounded border border-zinc-800 bg-zinc-900 p-3">
          <div className="text-xs text-zinc-500 mb-1">最近任务</div>
          <div className="text-sm font-medium">{tasks.length} 条事件</div>
        </div>
      </div>

      {/* Recent Tasks */}
      <div className="rounded border border-zinc-800 bg-zinc-900 p-3">
        <div className="text-xs text-zinc-500 mb-2">最近任务事件</div>
        {tasks.length === 0 ? (
          <div className="text-xs text-zinc-600">暂无任务</div>
        ) : (
          <div className="space-y-1">
            {tasks.map((t, i) => (
              <div key={i} className="flex items-center gap-2 text-xs">
                <span className="font-mono text-zinc-400 w-28 truncate">{t.task_id}</span>
                <span className="text-zinc-500">{t.event_type}</span>
                <span className={`px-1 rounded text-[10px] ${
                  t.status === 'completed' ? 'bg-emerald-900/50 text-emerald-400' :
                  t.status === 'failed' ? 'bg-red-900/50 text-red-400' :
                  'bg-zinc-800 text-zinc-400'
                }`}>{t.status}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
