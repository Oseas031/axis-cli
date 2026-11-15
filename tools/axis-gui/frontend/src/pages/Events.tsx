import { useEffect, useState, useRef, useCallback } from 'react'
import { useWebSocket } from '../hooks/useWebSocket'

interface RawEvent {
  raw: string
  timestamp?: string
}

function parseEvent(line: string): RawEvent {
  try {
    const obj = JSON.parse(line)
    return { raw: line, timestamp: obj.timestamp || obj.created_at }
  } catch {
    return { raw: line }
  }
}

export default function Events() {
  const [events, setEvents] = useState<RawEvent[]>([])
  const [autoScroll, setAutoScroll] = useState(true)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    fetch('/api/events')
      .then(r => r.json())
      .then((data: unknown[]) => setEvents(data.map(item => {
        if (typeof item === 'string') return parseEvent(item)
        const raw = JSON.stringify(item)
        const obj = item as Record<string, string>
        return { raw, timestamp: obj.timestamp || obj.created_at }
      })))
      .catch(() => {})
  }, [])

  const handleWsMessage = useCallback((data: string) => {
    setEvents(prev => [...prev, parseEvent(data)])
  }, [])

  useWebSocket('/ws/events', { onMessage: handleWsMessage })

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [events, autoScroll])

  return (
    <div className="space-y-2 h-full flex flex-col">
      <header className="flex items-center justify-between border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Events</h1>
        <button
          type="button"
          onClick={() => setAutoScroll(!autoScroll)}
          aria-pressed={autoScroll}
          className={`font-mono text-[11px] px-2 py-0.5 rounded-sm border ${
            autoScroll
              ? 'border-emerald-800 bg-emerald-950/30 text-emerald-400'
              : 'border-zinc-700 bg-zinc-900 text-zinc-400'
          }`}
        >
          {autoScroll ? '自动滚动: 开' : '自动滚动: 关'}
        </button>
      </header>

      <div
        ref={containerRef}
        role="log"
        aria-live="off"
        className="flex-1 min-h-0 overflow-auto border border-zinc-800 bg-zinc-900/60 p-2 font-mono text-[11px] leading-relaxed"
      >
        {events.length === 0 ? (
          <div className="text-zinc-600">暂无事件</div>
        ) : (
          events.map((e, i) => (
            <div key={i} className="hover:bg-zinc-800/40 px-1">
              {e.timestamp && <span className="text-emerald-500">{e.timestamp}</span>}
              {e.timestamp && <span className="text-zinc-700"> | </span>}
              <span className="text-zinc-400 break-all">{e.raw}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
