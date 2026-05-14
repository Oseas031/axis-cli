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
    <div className="space-y-3 h-full flex flex-col">
      <div className="flex items-center justify-between">
        <h2 className="font-title text-lg font-semibold">Events</h2>
        <button
          onClick={() => setAutoScroll(!autoScroll)}
          className={`text-xs px-2 py-0.5 rounded border ${
            autoScroll ? 'border-emerald-700 text-emerald-400' : 'border-zinc-700 text-zinc-400'
          }`}
        >
          {autoScroll ? '自动滚动: 开' : '自动滚动: 关'}
        </button>
      </div>

      <div
        ref={containerRef}
        className="flex-1 min-h-0 overflow-auto rounded border border-zinc-800 bg-zinc-900 p-2 font-mono text-[11px] leading-relaxed"
      >
        {events.length === 0 ? (
          <div className="text-zinc-600">暂无事件</div>
        ) : (
          events.map((e, i) => (
            <div key={i} className="hover:bg-zinc-800/30 px-1 rounded">
              {e.timestamp && <span className="text-emerald-500">{e.timestamp}</span>}
              {e.timestamp && <span className="text-zinc-700"> | </span>}
              <span className="text-zinc-400">{e.raw}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
