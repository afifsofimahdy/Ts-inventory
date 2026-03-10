import React from 'react'
import { createRoot } from 'react-dom/client'
import Shell from './layout/Shell.jsx'
import InventoryPage from './pages/InventoryPage.jsx'
import { ToastProvider } from './components/Toast.jsx'

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ToastProvider>
      <Shell active="inventory">
        <InventoryPage />
      </Shell>
    </ToastProvider>
  </React.StrictMode>
)
