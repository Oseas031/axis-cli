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
      <mark className="bg-yellow-600/40 text-yellow-200 rounded px-0.5">{text.slice(idx, idx + query.length)}</mark>
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
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // Cleanup poll timers
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
          // Update the pending message with progress info
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
        resolveTask(taskId, 'failed', (err as any).error || (err as any).hint || `提交失败 (HTTP ${resp.status})`)
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
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-3">
        <h2 className="font-title text-lg font-semibold">Chat</h2>
        <div className="flex items-center gap-2">
          {showSearch ? (
            <div className="flex items-center gap-1">
              <input
                type="text"
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                placeholder="搜索历史..."
                autoFocus
                className="px-2 py-1 text-[11px] rounded border border-zinc-700 bg-zinc-900 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 w-40"
              />
              <button onClick={() => { setShowSearch(false); setSearchQuery('') }} className="text-zinc-500 hover:text-zinc-300"><X size={12} /></button>
            </div>
          ) : (
            <button onClick={() => setShowSearch(true)} className="text-zinc-500 hover:text-zinc-300" title="搜索历史"><Search size={14} /></button>
          )}
          {messages.length > 0 && (
            <button onClick={() => { setMessages([]); saveHistory([]) }} className="text-[10px] text-zinc-600 hover:text-zinc-400">清空</button>
          )}
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto space-y-2 mb-3">
        {hiddenCount > 0 && (
          <button
            onClick={() => setExpanded(true)}
            className="w-full flex items-center justify-center gap-1 py-1.5 text-[11px] text-zinc-500 hover:text-zinc-300 border border-zinc-800 rounded hover:border-zinc-700 transition-colors"
          >
            <ChevronUp size={12} />
            展开 {hiddenCount} 条历史消息
          </button>
        )}
        {expanded && messages.length > RECENT_COUNT && !searchQuery && (
          <button
            onClick={() => setExpanded(false)}
            className="w-full flex items-center justify-center gap-1 py-1 text-[10px] text-zinc-600 hover:text-zinc-400"
          >
            收起历史
          </button>
        )}
        {visibleMessages.length === 0 && !searchQuery && <div className="text-xs text-zinc-600">发送消息开始对话</div>}
        {visibleMessages.length === 0 && searchQuery && <div className="text-xs text-zinc-600">未找到匹配的消息</div>}
        {visibleMessages.map((m, i) => (
          <div key={i} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div className={`max-w-[70%] px-3 py-1.5 rounded text-xs ${
              m.role === 'user'
                ? 'bg-zinc-700 text-zinc-100'
                : m.status === 'pending'
                  ? 'bg-zinc-800 text-zinc-400'
                  : m.status === 'failed'
                    ? 'bg-red-950/30 border border-red-900/50 text-red-300'
                    : 'bg-zinc-800 text-zinc-200'
            }`}>
              {m.status === 'pending' && (
                <span className="inline-block w-3 h-3 border-2 border-zinc-600 border-t-zinc-300 rounded-full animate-spin mr-1.5 align-middle" />
              )}
              {searchQuery ? highlightMatch(m.content, searchQuery) : m.content}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入消息..."
          className="flex-1 px-3 py-1.5 text-xs rounded border border-zinc-700 bg-zinc-900 text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
        />
        <button
          onClick={send}
          className="px-3 py-1.5 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-200 transition-colors"
        >
          <Send size={14} />
        </button>
      </div>
    </div>
  )
}
