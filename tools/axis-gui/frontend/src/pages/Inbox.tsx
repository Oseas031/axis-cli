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

  if (loading) return <div className="text-xs font-mono text-zinc-500">加载中...</div>

  return (
    <div className="space-y-3">
      <header className="flex items-center justify-between border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Inbox</h1>
        <button
          type="button"
          onClick={load}
          className="font-mono text-[11px] text-zinc-300 hover:text-zinc-100 border border-zinc-700 bg-zinc-900 rounded-sm px-2 py-0.5"
        >
          refresh
        </button>
      </header>

      {error && (
        <div role="alert" className="font-mono text-xs text-red-400 border border-red-900/70 bg-red-950/20 px-2.5 py-1.5">
          {error}
        </div>
      )}

      {!records || records.length === 0 ? (
        <div className="text-xs font-mono text-zinc-500 border border-zinc-800 bg-zinc-900/40 p-3">
          No human requests. Agents will drop decision/info requests here.
        </div>
      ) : (
        <div className="space-y-2">
          {records.map(rec => {
            const req = rec.message.payload?.request
            const pending = rec.status === 'pending'
            return (
              <section
                key={rec.message.id}
                className={`border p-3 ${
                  pending ? 'border-amber-700/80 bg-amber-950/15' : 'border-zinc-800 bg-zinc-900/40'
                }`}
              >
                <div className="flex items-center justify-between pb-1.5 mb-1.5 border-b border-zinc-800/70">
                  <span className="font-mono text-xs text-zinc-300">{rec.message.id}</span>
                  <span className={`font-mono text-[10px] px-1.5 py-0.5 rounded-sm border ${
                    pending
                      ? 'border-amber-800/70 bg-amber-950/40 text-amber-400'
                      : rec.status === 'completed'
                        ? 'border-emerald-800/70 bg-emerald-950/40 text-emerald-400'
                        : 'border-red-800/70 bg-red-950/40 text-red-400'
                  }`}>
                    {rec.status}
                  </span>
                </div>

                <div className="text-xs text-zinc-100 mb-1.5">
                  <span className="font-mono text-[10px] uppercase tracking-wider text-zinc-500 border border-zinc-800 px-1 py-0.5 mr-2">
                    {req?.kind ?? 'info'}
                  </span>
                  {req?.prompt ?? '(no prompt)'}
                </div>

                <div className="flex flex-wrap gap-3 font-mono text-[11px] text-zinc-500">
                  {req?.task_id && <span>task: {req.task_id}</span>}
                  {rec.nudges ? <span>nudged {rec.nudges}×</span> : null}
                </div>

                {pending && req?.options && req.options.length > 0 && (
                  <div className="flex flex-wrap gap-2 mt-2.5 pt-2 border-t border-zinc-800/70">
                    {req.options.map(opt => (
                      <button
                        key={opt}
                        type="button"
                        disabled={busy === rec.message.id}
                        onClick={() => respond(rec.message.id, { choice: opt })}
                        className="font-mono text-xs px-3 py-1 rounded-sm border border-zinc-700 bg-zinc-800 hover:bg-zinc-700 text-zinc-100 disabled:opacity-50"
                      >
                        {opt}
                      </button>
                    ))}
                  </div>
                )}

                {pending && (!req?.options || req.options.length === 0) && (
                  <form
                    className="flex gap-2 mt-2.5 pt-2 border-t border-zinc-800/70"
                    onSubmit={e => {
                      e.preventDefault()
                      const input = e.currentTarget.elements.namedItem('answer') as HTMLInputElement | null
                      if (input && input.value.trim()) respond(rec.message.id, { text: input.value })
                    }}
                  >
                    <label htmlFor={`answer-${rec.message.id}`} className="sr-only">
                      Response for {rec.message.id}
                    </label>
                    <input
                      id={`answer-${rec.message.id}`}
                      name="answer"
                      placeholder="your answer..."
                      disabled={busy === rec.message.id}
                      className="flex-1 text-xs font-mono bg-zinc-900 border border-zinc-700 rounded-sm px-2 py-1 text-zinc-200 placeholder-zinc-600"
                    />
                    <button
                      type="submit"
                      disabled={busy === rec.message.id}
                      className="font-mono text-xs px-3 py-1 rounded-sm border border-zinc-700 bg-zinc-800 hover:bg-zinc-700 text-zinc-100 disabled:opacity-50"
                    >
                      send
                    </button>
                  </form>
                )}
              </section>
            )
          })}
        </div>
      )}
    </div>
  )
}
