import React from 'react'
import ReactDOM from 'react-dom/client'
import { HashRouter, Routes, Route } from 'react-router-dom'
import App from './App'
import Dashboard from './pages/Dashboard'
import Tasks from './pages/Tasks'
import Providers from './pages/Providers'
import Events from './pages/Events'
import Chat from './pages/Chat'
import Skills from './pages/Skills'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <HashRouter>
      <Routes>
        <Route path="/" element={<App />}>
          <Route index element={<Dashboard />} />
          <Route path="tasks" element={<Tasks />} />
          <Route path="providers" element={<Providers />} />
          <Route path="events" element={<Events />} />
          <Route path="chat" element={<Chat />} />
          <Route path="skills" element={<Skills />} />
        </Route>
      </Routes>
    </HashRouter>
  </React.StrictMode>
)
