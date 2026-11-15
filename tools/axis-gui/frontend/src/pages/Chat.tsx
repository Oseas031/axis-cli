import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Send, Search, ChevronUp, X } from 'lucide-react'
import { useWebSocket } from '../hooks/useWebSocket'

interface ChatMessage {
  role: 'user' | 'system'
  content: string
  taskId?: string
  status?: 'pending' | 'completed' | 'failed'
  timestamp?: number
}

const STORAGE_KEY = 'axis-chat-history'

function loadHistory(): ChatMessage[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const msgs: ChatMessage[] = JSON.parse(raw)
    // Mark stale pending messages as failed (no active poller after reload)
    return msgs.map(m => m.status === 'pending' ? { ...m, status: 'failed' as const, content: '任务中断（页面已刷新）' } : m)
  } catch { return [] }
}

function saveHistory(msgs: ChatMessage[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(msgs))
}

function highlightMatch(text: string, query: string) {
  if (!query.trim()) return text
  const idx = text.toLowerCase().indexOf(query.toLowerCase())
  if (idx < 0) return text
  return (
    <>
      {text.slice(0, idx)}
      <mark className="bg-amber-900/60 text-amber-200 px-0.5">{text.slice(idx, idx + query.length)}</mark>
      {text.slice(idx + query.length)}
    </>
  )
}

export default function Chat() {
  const [messages, setMessages] = useState<ChatMessage[]>(loadHistory)
  const [input, setInput] = useState('')
  const [pendingTaskIds, setPendingTaskIds] = useState<Set<string>>(new Set())
  const [expanded, setExpanded] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [showSearch, setShowSearch] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)
  const pollTimers = useRef<Map<string, ReturnType<typeof setInterval>>>(new Map())

  const RECENT_COUNT = 20

  const visibleMessages = useMemo(() => {
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      return messages.filter(m => m.content.toLowerCase().includes(q))
    }
    if (expanded || messages.length <= RECENT_COUNT) return messages
    return messages.slice(-RECENT_COUNT)
  }, [messages, expanded, searchQuery])

  const hiddenCount = searchQuery ? 0 : (expanded ? 0 : Math.max(0, messages.length - RECENT_COUNT))

  useEffect(() => {
    saveHistory(messages)
    bottomRef.current?.scrollIntoView({ behavior: 'auto' })
  }, [messages])

  useEffect(() => {
    return () => {
      pollTimers.current.forEach(t => clearInterval(t))
    }
  }, [])

  const resolveTask = useCallback((taskId: string, status: 'completed' | 'failed', result?: string) => {
    setPendingTaskIds(prev => { const n = new Set(prev); n.delete(taskId); return n })
    const timer = pollTimers.current.get(taskId)
    if (timer) { clearInterval(timer); pollTimers.current.delete(taskId) }

    setMessages(prev => {
      const updated = [...prev]
      const idx = updated.findIndex(m => m.taskId === taskId && m.role === 'system')
      if (idx >= 0) {
        updated[idx] = { ...updated[idx], status, content: result || (status === 'completed' ? '任务完成' : '任务失败') }
      } else {
        updated.push({ role: 'system', content: result || (status === 'completed' ? '任务完成' : '任务失败'), taskId, status })
      }
      return updated
    })
  }, [])

  const startPolling = useCallback((taskId: string) => {
    let errorCount = 0
    const maxErrors = 5
    const timeout = setTimeout(() => resolveTask(taskId, 'failed', '任务超时（60s 未完成）'), 60000)
    const timer = setInterval(() => {
      fetch(`/api/tasks/${taskId}/status`)
        .then(r => {
          if (!r.ok) throw new Error(`HTTP ${r.status}`)
          return r.json()
        })
        .then(d => {
          errorCount = 0
          if (d.status === 'completed' || d.status === 'failed') {
            clearTimeout(timeout)
            const text = d.output?.text || d.output?.result || d.error || d.message || (d.output ? JSON.stringify(d.output) : undefined)
            resolveTask(taskId, d.status, text)
          }
        })
        .catch(() => {
          errorCount++
          if (errorCount >= maxErrors) {
            clearTimeout(timeout)
            resolveTask(taskId, 'failed', '任务状态查询失败（服务不可达）')
          }
        })
    }, 2000)
    pollTimers.current.set(taskId, timer)
  }, [resolveTask])

  const handleWsMessage = useCallback((data: string) => {
    try {
      const event = JSON.parse(data)
      if (event.task_id && pendingTaskIds.has(event.task_id)) {
        if (event.status === 'completed' || event.status === 'failed') {
          resolveTask(event.task_id, event.status, event.message || event.result)
        } else if (event.event_type === 'execution_started' || event.event_type === 'tool_executed') {
          setMessages(prev => prev.map(m =>
            m.taskId === event.task_id && m.status === 'pending'
              ? { ...m, content: event.message || '处理中...' }
              : m
          ))
        }
      }
    } catch { /* ignore */ }
  }, [pendingTaskIds, resolveTask])

  useWebSocket('/ws/events', { onMessage: handleWsMessage })

  const send = async () => {
    const text = input.trim()
    if (!text) return
    setInput('')

    const taskId = `chat-${Date.now()}`
    const userMsg: ChatMessage = { role: 'user', content: text, timestamp: Date.now() }
    const sysMsg: ChatMessage = { role: 'system', content: '处理中...', taskId, status: 'pending', timestamp: Date.now() }
    setMessages(prev => [...prev, userMsg, sysMsg])
    setPendingTaskIds(prev => new Set(prev).add(taskId))

    try {
      const resp = await fetch('/api/tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ task: { task_id: taskId, contract_id: 'default', input: { message: text }, status: 'pending', metadata: { executor_type: 'agent' } } })
      })
      if (!resp.ok) {
        const err = await resp.json().catch(() => ({}))
        resolveTask(taskId, 'failed', (err as { error?: string; hint?: string }).error || (err as { hint?: string }).hint || `提交失败 (HTTP ${resp.status})`)
        return
      }
      startPolling(taskId)
    } catch {
      resolveTask(taskId, 'failed', '发送失败，无法连接服务')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send() }
  }

  return (
    <div className="flex flex-col h-full space-y-2">
      <header className="flex items-center justify-between border-b border-zinc-800 pb-2">
        <h1 className="font-mono text-sm font-semibold text-zinc-100 uppercase tracking-wider">Chat</h1>
        <div className="flex items-center gap-2">
          {showSearch ? (
            <div className="flex items-center gap-1">
              <label htmlFor="chat-search" className="sr-only">搜索历史</label>
              <input
                id="chat-search"
                type="text"
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                placeholder="搜索历史..."
                autoFocus
                className="px-2 py-0.5 text-[11px] font-mono rounded-sm border border-zinc-700 bg-zinc-900 text-zinc-200 placeholder-zinc-600 w-40"
              />
              <button
                type="button"
                aria-label="关闭搜索"
                onClick={() => { setShowSearch(false); setSearchQuery('') }}
                className="text-zinc-500 hover:text-zinc-300 p-0.5"
              >
                <X size={12} aria-hidden="true" />
              </button>
            </div>
          ) : (
            <button
              type="button"
              aria-label="搜索历史"
              onClick={() => setShowSearch(true)}
              className="text-zinc-400 hover:text-zinc-200 p-0.5"
              title="搜索历史"
            >
              <Search size={14} aria-hidden="true" />
            </button>
          )}
          {messages.length > 0 && (
            <button
              type="button"
              onClick={() => { setMessages([]); saveHistory([]) }}
              className="font-mono text-[10px] text-zinc-500 hover:text-zinc-300 border border-zinc-800 px-1.5 py-0.5 rounded-sm"
            >
              清空
            </button>
          )}
        </div>
      </header>

      <div
        role="log"
        aria-live="polite"
        className="flex-1 min-h-0 overflow-auto border border-zinc-800 bg-zinc-900/40 divide-y divide-zinc-800/60"
      >
        {hiddenCount > 0 && (
          <button
            type="button"
            onClick={() => setExpanded(true)}
            className="w-full flex items-center justify-center gap-1 py-1.5 font-mono text-[11px] text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40"
          >
            <ChevronUp size={12} aria-hidden="true" />
            展开 {hiddenCount} 条历史消息
          </button>
        )}
        {expanded && messages.length > RECENT_COUNT && !searchQuery && (
          <button
            type="button"
            onClick={() => setExpanded(false)}
            className="w-full flex items-center justify-center gap-1 py-1 font-mono text-[10px] text-zinc-500 hover:text-zinc-300"
          >
            收起历史
          </button>
        )}
        {visibleMessages.length === 0 && !searchQuery && (
          <div className="p-3 font-mono text-xs text-zinc-600">输入指令以提交至 Axis 控制面</div>
        )}
        {visibleMessages.length === 0 && searchQuery && (
          <div className="p-3 font-mono text-xs text-zinc-600">未找到匹配的消息</div>
        )}
        {visibleMessages.map((m, i) => (
          <div
            key={i}
            className={`px-3 py-2 text-xs ${
              m.role === 'user'
                ? 'bg-zinc-900/80 border-l-2 border-zinc-500'
                : m.status === 'failed'
                  ? 'bg-red-950/20 border-l-2 border-red-600 text-red-300'
                  : m.status === 'pending'
                    ? 'border-l-2 border-amber-500 text-zinc-400'
                    : 'border-l-2 border-emerald-600 text-zinc-200'
            }`}
          >
            <div className="flex items-center justify-between font-mono text-[10px] text-zinc-500 mb-1">
              <span>{m.role === 'user' ? '> USER' : `AXIS${m.taskId ? ` · ${m.taskId}` : ''}`}</span>
              {m.status && (
                <span className={
                  m.status === 'completed' ? 'text-emerald-400' :
                  m.status === 'failed' ? 'text-red-400' : 'text-amber-400'
                }>
                  {m.status}
                </span>
              )}
            </div>
            <div className="whitespace-pre-wrap break-words leading-relaxed">
              {searchQuery ? highlightMatch(m.content, searchQuery) : m.content}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div className="flex gap-2">
        <label htmlFor="chat-input" className="sr-only">输入指令</label>
        <input
          id="chat-input"
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入消息..."
          className="flex-1 px-2.5 py-1.5 text-xs font-mono rounded-sm border border-zinc-700 bg-zinc-900 text-zinc-200 placeholder-zinc-600 focus:border-zinc-500"
        />
        <button
          type="button"
          aria-label="发送消息"
          onClick={send}
          className="px-3 py-1.5 rounded-sm border border-zinc-700 bg-zinc-800 hover:bg-zinc-700 text-zinc-100"
        >
          <Send size={13} aria-hidden="true" />
        </button>
      </div>
    </div>
  )
}
