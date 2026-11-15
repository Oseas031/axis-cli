import { Outlet, NavLink } from 'react-router-dom'
import { LayoutDashboard, ListTodo, Cpu, Radio, MessageSquare, BookOpen, Inbox as InboxIcon } from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/tasks', icon: ListTodo, label: 'Tasks' },
  { to: '/providers', icon: Cpu, label: 'Providers' },
  { to: '/events', icon: Radio, label: 'Events' },
  { to: '/skills', icon: BookOpen, label: 'Skills' },
  { to: '/inbox', icon: InboxIcon, label: 'Inbox' },
  { to: '/chat', icon: MessageSquare, label: 'Chat' },
]

export default function App() {
  return (
    <div className="flex h-screen overflow-hidden bg-zinc-950 text-zinc-200">
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      <aside className="w-44 shrink-0 border-r border-zinc-800 bg-zinc-900/70 flex flex-col">
        <div className="px-3 py-2.5 border-b border-zinc-800 flex items-center justify-between">
          <span className="font-mono text-xs font-semibold text-zinc-100 tracking-wider uppercase">
            AXIS // OBS
          </span>
          <span className="text-[10px] font-mono text-zinc-500">CLI</span>
        </div>
        <nav aria-label="Primary" className="flex-1 py-1.5 px-1.5 space-y-0.5 overflow-y-auto">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-2 px-2.5 py-1.5 rounded-sm text-xs font-mono border-l-2 ${
                  isActive
                    ? 'border-emerald-500 bg-zinc-800/80 text-zinc-100'
                    : 'border-transparent text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40'
                }`
              }
            >
              <Icon size={13} aria-hidden="true" />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="px-3 py-2 border-t border-zinc-800 text-[10px] font-mono text-zinc-600">
          read-only observatory
        </div>
      </aside>
      <main id="main-content" tabIndex={-1} className="flex-1 min-w-0 overflow-auto p-4">
        <Outlet />
      </main>
    </div>
  )
}
