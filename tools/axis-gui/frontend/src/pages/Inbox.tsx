import { useEffect, useState } from 'react'

interface InboxRecord {
  message: {
    id: string
    payload?: {
      request?: {
        kind?: string
        prompt?: string
        task_id?: string
        options?: string[]
        deadline?: string
      }
    }
  }
  status: string
  output?: Record<string, unknown>
  nudges?: number
}

export default function Inbox() {
  const [records, setRecords] = useState<InboxRecord[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    fetch('/api/inbox')
      .then(r => r.json())
      .then(d => setRecords(d.records ?? []))
      .catch(() => setRecords(null))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const respond = async (id: string, body: { choice?: string; text?: string }) => {
    setBusy(id)
    setError(null)
    try {
      const res = await fetch(`/api/inbox/${encodeURIComponent(id)}/respond`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        setError(data.error ?? `respond failed (${res.status})`)
      }
    } catch {
      setError('respond failed: runtime unreachable')
    } finally {
      setBusy(null)
      load()
    }
  }

  if (loading) return <div className="text-xs text-zinc-500">加载中...</div>

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="font-title text-lg font-semibold">Inbox</h2>
        <button onClick={load} className="text-xs text-zinc-400 hover:text-zinc-200 border border-zinc-800 rounded px-2 py-1">
          refresh
        </button>
      </div>
      {error && <div className="text-xs text-red-400">{error}</div>}
      {!records || records.length === 0 ? (
        <div className="text-sm text-zinc-500">No human requests. Agents will drop decision/info requests here.</div>
      ) : (
        <div className="space-y-2">
          {records.map(rec => {
            const req = rec.message.payload?.request
            const pending = rec.status === 'pending'
            return (
              <div key={rec.message.id} className={`rounded border p-3 ${pending ? 'border-amber-800 bg-amber-950/10' : 'border-zinc-800 bg-zinc-900'}`}>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-mono text-zinc-400">{rec.message.id}</span>
                  <span className={`text-[10px] px-1.5 py-0.5 rounded ${
                    pending ? 'bg-amber-900/50 text-amber-400' :
                    rec.status === 'completed' ? 'bg-emerald-900/50 text-emerald-400' : 'bg-red-900/50 text-red-400'
                  }`}>{rec.status}</span>
                </div>
                <div className="text-sm text-zinc-200 mb-1">
                  <span className="text-[10px] uppercase tracking-wide text-zinc-500 mr-2">{req?.kind ?? 'info'}</span>
                  {req?.prompt ?? '(no prompt)'}
                </div>
                {req?.task_id && <div className="text-xs text-zinc-500 font-mono mb-1">task: {req.task_id}</div>}
                {rec.nudges ? <div className="text-[10px] text-zinc-600 mb-1">nudged {rec.nudges}×</div> : null}
                {pending && req?.options && req.options.length > 0 && (
                  <div className="flex gap-2 mt-2">
                    {req.options.map(opt => (
                      <button
                        key={opt}
                        disabled={busy === rec.message.id}
                        onClick={() => respond(rec.message.id, { choice: opt })}
                        className="text-xs px-3 py-1 rounded border border-zinc-700 bg-zinc-800 hover:bg-zinc-700 text-zinc-100 disabled:opacity-50"
                      >
                        {opt}
                      </button>
                    ))}
                  </div>
                )}
                {pending && (!req?.options || req.options.length === 0) && (
                  <form
                    className="flex gap-2 mt-2"
                    onSubmit={e => {
                      e.preventDefault()
                      const input = e.currentTarget.elements.namedItem('answer') as HTMLInputElement | null
                      if (input && input.value.trim()) respond(rec.message.id, { text: input.value })
                    }}
                  >
                    <input name="answer" placeholder="your answer..." disabled={busy === rec.message.id}
                      className="flex-1 text-xs bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-200" />
                    <button type="submit" disabled={busy === rec.message.id}
                      className="text-xs px-3 py-1 rounded border border-zinc-700 bg-zinc-800 hover:bg-zinc-700 text-zinc-100 disabled:opacity-50">
                      send
                    </button>
                  </form>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
