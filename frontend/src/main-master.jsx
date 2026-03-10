import React from 'react'
import { createRoot } from 'react-dom/client'
import Shell from './layout/Shell.jsx'
import MasterDataPage from './pages/MasterDataPage.jsx'
import { ToastProvider } from './components/Toast.jsx'

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ToastProvider>
      <Shell active="master">
        <MasterDataPage />
      </Shell>
    </ToastProvider>
  </React.StrictMode>
)
