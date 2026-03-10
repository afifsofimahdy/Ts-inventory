import React from 'react'
import { createRoot } from 'react-dom/client'
import Shell from './layout/Shell.jsx'
import StockOutPage from './pages/StockOutPage.jsx'
import { ToastProvider } from './components/Toast.jsx'

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ToastProvider>
      <Shell active="stock-out">
        <StockOutPage />
      </Shell>
    </ToastProvider>
  </React.StrictMode>
)
