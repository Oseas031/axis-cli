import { Outlet, NavLink } from 'react-router-dom'
import { LayoutDashboard, ListTodo, Cpu, Radio, MessageSquare, BookOpen } from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/tasks', icon: ListTodo, label: 'Tasks' },
  { to: '/providers', icon: Cpu, label: 'Providers' },
  { to: '/events', icon: Radio, label: 'Events' },
  { to: '/skills', icon: BookOpen, label: 'Skills' },
  { to: '/chat', icon: MessageSquare, label: 'Chat' },
]

export default function App() {
  return (
    <div className="flex h-screen overflow-hidden">
      <aside className="w-48 shrink-0 border-r border-zinc-800 bg-zinc-900 flex flex-col">
        <div className="px-4 py-3 border-b border-zinc-800">
          <h1 className="font-title text-sm font-semibold text-zinc-100 tracking-wide">AXIS GUI</h1>
        </div>
        <nav className="flex-1 py-2 space-y-0.5 px-2">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                  isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/50'
                }`
              }
            >
              <Icon size={14} />
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="flex-1 overflow-auto p-4">
        <Outlet />
      </main>
    </div>
  )
}
