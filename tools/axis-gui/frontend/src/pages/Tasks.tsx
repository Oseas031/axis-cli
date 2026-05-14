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
      <h2 className="font-title text-lg font-semibold">Tasks</h2>

      <input
        type="text"
        placeholder="按 task_id 筛选..."
        value={filter}
        onChange={e => setFilter(e.target.value)}
        className="w-full max-w-xs px-2 py-1 text-xs rounded border border-zinc-700 bg-zinc-900 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
      />

      <div className="rounded border border-zinc-800 bg-zinc-900 overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-500">
              <th className="text-left px-3 py-1.5 font-medium">task_id</th>
              <th className="text-left px-3 py-1.5 font-medium">event_type</th>
              <th className="text-left px-3 py-1.5 font-medium">status</th>
              <th className="text-left px-3 py-1.5 font-medium">created_at</th>
              <th className="text-left px-3 py-1.5 font-medium">message</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr><td colSpan={5} className="px-3 py-4 text-center text-zinc-600">暂无任务事件</td></tr>
            ) : (
              filtered.map((t, i) => (
                <tr key={i} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                  <td className="px-3 py-1.5 font-mono text-zinc-300">{t.task_id}</td>
                  <td className="px-3 py-1.5 text-zinc-400">{t.event_type}</td>
                  <td className="px-3 py-1.5">
                    <span className={`px-1 rounded text-[10px] ${
                      t.status === 'completed' ? 'bg-emerald-900/50 text-emerald-400' :
                      t.status === 'failed' ? 'bg-red-900/50 text-red-400' :
                      'bg-zinc-800 text-zinc-400'
                    }`}>{t.status}</span>
                  </td>
                  <td className="px-3 py-1.5 text-zinc-500">{t.created_at}</td>
                  <td className="px-3 py-1.5 text-zinc-500 max-w-xs truncate">{t.message || '-'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
