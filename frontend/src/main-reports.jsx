import React from 'react'
import { createRoot } from 'react-dom/client'
import Shell from './layout/Shell.jsx'
import ReportsPage from './pages/ReportsPage.jsx'
import { ToastProvider } from './components/Toast.jsx'

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ToastProvider>
      <Shell active="reports">
        <ReportsPage />
      </Shell>
    </ToastProvider>
  </React.StrictMode>
)
