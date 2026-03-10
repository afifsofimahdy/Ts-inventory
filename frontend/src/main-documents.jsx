import { createRoot } from 'react-dom/client'
import Shell from './layout/Shell.jsx'
import DocumentsPage from './pages/DocumentsPage.jsx'
import { ToastProvider } from './components/Toast.jsx'

createRoot(document.getElementById('root')).render(
  <ToastProvider>
    <Shell active="documents">
      <DocumentsPage />
    </Shell>
  </ToastProvider>
)
