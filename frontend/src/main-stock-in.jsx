import React from 'react'
import { createRoot } from 'react-dom/client'
import Shell from './layout/Shell.jsx'
import StockInPage from './pages/StockInPage.jsx'
import { ToastProvider } from './components/Toast.jsx'

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ToastProvider>
      <Shell active="stock-in">
        <StockInPage />
      </Shell>
    </ToastProvider>
  </React.StrictMode>
)
