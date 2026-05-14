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

  if (loading) return <div className="text-xs text-zinc-500">加载中...</div>

  if (!providers || Object.keys(providers).length === 0) {
    return (
      <div className="space-y-3">
        <h2 className="font-title text-lg font-semibold">Providers</h2>
        <div className="text-sm text-zinc-500">No providers configured</div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <h2 className="font-title text-lg font-semibold">Providers</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {Object.entries(providers).map(([name, p]) => (
          <div key={name} className={`rounded border p-3 ${p.active ? 'border-emerald-800 bg-emerald-950/20' : 'border-zinc-800 bg-zinc-900'}`}>
            <div className="flex items-center justify-between mb-1">
              <span className="text-sm font-medium text-zinc-200">{name}</span>
              {p.active && <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-900/50 text-emerald-400">active</span>}
            </div>
            <div className="text-xs text-zinc-500 space-y-0.5">
              <div>类型: <span className="text-zinc-400">{p.type}</span></div>
              <div>模型: <span className="text-zinc-400 font-mono">{p.model}</span></div>
              <div>API Key: <span className="text-zinc-400 font-mono">{maskKey(p.api_key)}</span></div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
