import { useEffect, useState, useCallback } from 'react'
import { useWebSocket } from '../hooks/useWebSocket'

interface TaskEvent {
  task_id: string
  event_type: string
  status: string
  created_at: string
  message?: string
}

export default function Tasks() {
  const [tasks, setTasks] = useState<TaskEvent[]>([])
  const [filter, setFilter] = useState('')

  useEffect(() => {
    fetch('/api/tasks')
      .then(r => r.json())
      .then(d => setTasks(d.tasks || []))
      .catch(() => {})
  }, [])

  const handleWsMessage = useCallback((data: string) => {
    try {
      const event = JSON.parse(data) as TaskEvent
      setTasks(prev => [event, ...prev])
    } catch { /* ignore non-json */ }
  }, [])

  useWebSocket('/ws/events', { onMessage: handleWsMessage })

  const filtered = filter
    ? tasks.filter(t => t.task_id.includes(filter))
    : tasks

  return (
    <div className="space-y-3">
      <header className="flex flex-wrap items-center justify-between gap-2 border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Tasks</h1>
        <div className="flex items-center gap-2">
          <label htmlFor="task-filter" className="sr-only">按 task_id 筛选</label>
          <input
            id="task-filter"
            type="text"
            placeholder="按 task_id 筛选..."
            value={filter}
            onChange={e => setFilter(e.target.value)}
            className="w-56 px-2 py-1 text-xs font-mono rounded-sm border border-zinc-700 bg-zinc-900 text-zinc-200 placeholder-zinc-600 focus:border-zinc-500"
          />
        </div>
      </header>

      <div className="border border-zinc-800 bg-zinc-900/40 overflow-x-auto">
        <table className="w-full text-xs border-collapse">
          <caption className="sr-only">Axis 任务事件流水表</caption>
          <thead>
            <tr className="border-b border-zinc-800 bg-zinc-900 text-zinc-500 font-mono text-[11px]">
              <th scope="col" className="text-left px-3 py-1.5 font-medium">task_id</th>
              <th scope="col" className="text-left px-3 py-1.5 font-medium">event_type</th>
              <th scope="col" className="text-left px-3 py-1.5 font-medium">status</th>
              <th scope="col" className="text-left px-3 py-1.5 font-medium">created_at</th>
              <th scope="col" className="text-left px-3 py-1.5 font-medium">message</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800/50">
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-3 py-4 text-center text-zinc-600 font-mono">暂无任务事件</td>
              </tr>
            ) : (
              filtered.map((t, i) => (
                <tr key={i} className="hover:bg-zinc-800/30">
                  <td className="px-3 py-1.5 font-mono text-zinc-200 whitespace-nowrap">{t.task_id}</td>
                  <td className="px-3 py-1.5 font-mono text-zinc-400 whitespace-nowrap">{t.event_type}</td>
                  <td className="px-3 py-1.5 whitespace-nowrap">
                    <span className={`px-1.5 py-0.5 rounded-sm font-mono text-[10px] border ${
                      t.status === 'completed'
                        ? 'border-emerald-800/70 bg-emerald-950/40 text-emerald-400'
                        : t.status === 'failed'
                          ? 'border-red-800/70 bg-red-950/40 text-red-400'
                          : 'border-zinc-700 bg-zinc-800 text-zinc-400'
                    }`}>
                      {t.status}
                    </span>
                  </td>
                  <td className="px-3 py-1.5 font-mono text-zinc-500 whitespace-nowrap">{t.created_at}</td>
                  <td className="px-3 py-1.5 text-zinc-400 max-w-xs truncate">{t.message || '-'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
