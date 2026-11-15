import { useEffect, useState } from 'react'

interface ProviderProfile {
  name?: string
  type: string
  model: string
  active?: boolean
  api_key?: string
}

function maskKey(key: string | undefined): string {
  if (!key) return '****'
  return key.slice(0, 4) + '••••••••'
}

export default function Providers() {
  const [providers, setProviders] = useState<Record<string, ProviderProfile> | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/providers')
      .then(r => r.json())
      .then(d => setProviders(d))
      .catch(() => setProviders(null))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="text-xs font-mono text-zinc-500">加载中...</div>

  return (
    <div className="space-y-3">
      <header className="flex items-center justify-between border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Providers</h1>
        <span className="font-mono text-[11px] text-zinc-500">
          {providers ? `${Object.keys(providers).length} profiles` : '0 profiles'}
        </span>
      </header>

      {!providers || Object.keys(providers).length === 0 ? (
        <div className="text-xs font-mono text-zinc-500 border border-zinc-800 bg-zinc-900/40 p-3">
          No providers configured
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
          {Object.entries(providers).map(([name, p]) => (
            <section
              key={name}
              className={`border p-3 ${
                p.active ? 'border-emerald-800 bg-emerald-950/15' : 'border-zinc-800 bg-zinc-900/40'
              }`}
            >
              <div className="flex items-center justify-between pb-1.5 mb-1.5 border-b border-zinc-800/80">
                <span className="font-mono text-xs font-medium text-zinc-100">{name}</span>
                {p.active && (
                  <span className="font-mono text-[10px] px-1.5 py-0.5 rounded-sm border border-emerald-800/70 bg-emerald-950/40 text-emerald-400">
                    active
                  </span>
                )}
              </div>
              <dl className="grid grid-cols-[64px_1fr] gap-y-1 text-xs">
                <dt className="text-zinc-500 font-mono">类型</dt>
                <dd className="text-zinc-300 font-mono">{p.type}</dd>
                <dt className="text-zinc-500 font-mono">模型</dt>
                <dd className="text-zinc-300 font-mono truncate">{p.model}</dd>
                <dt className="text-zinc-500 font-mono">Key</dt>
                <dd className="text-zinc-400 font-mono">{maskKey(p.api_key)}</dd>
              </dl>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}
