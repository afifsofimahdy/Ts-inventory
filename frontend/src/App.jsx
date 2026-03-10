import { useEffect, useState } from 'react'
import InventoryPage from './pages/InventoryPage.jsx'
import StockInPage from './pages/StockInPage.jsx'
import StockOutPage from './pages/StockOutPage.jsx'
import ReportsPage from './pages/ReportsPage.jsx'
import MasterDataPage from './pages/MasterDataPage.jsx'
import { useInventoryStore } from './store/useInventoryStore.js'
import { ToastProvider } from './components/Toast.jsx'
import './styles.css'

const tabs = [
  { id: 'inventory', label: 'Inventory' },
  { id: 'stock-in', label: 'Stock In' },
  { id: 'stock-out', label: 'Stock Out' },
  { id: 'reports', label: 'Reports' },
  { id: 'master', label: 'Master Data' }
]

export default function App() {
  const [tab, setTab] = useState('inventory')
  const fetchInventory = useInventoryStore(s => s.fetchInventory)

  useEffect(() => {
    fetchInventory({})
  }, [fetchInventory])

  return (
    <ToastProvider>
      <div className="app-shell">
        <aside className="sidebar">
          <div className="brand">
            <img className="brand-logo" src="/assets/logo.png" alt="Tirtamas logo" />
            <div className="brand-title">Tirtamas Inventory</div>
            <div className="brand-sub">Stock Control</div>
          </div>
          <nav className="side-nav">
            {tabs.map(t => (
              <button
                key={t.id}
                className={tab === t.id ? 'nav-item active' : 'nav-item'}
                onClick={() => setTab(t.id)}
              >
                <span className="nav-label">{t.label}</span>
              </button>
            ))}
          </nav>
        </aside>

        <main className="content">
          <header className="content-header">
            <div className="content-title">{tabs.find(t => t.id === tab)?.label}</div>
            <div className="content-subtitle">Tirtamas Inventory</div>
          </header>
          <div className="content-body">
            {tab === 'inventory' && <InventoryPage />}
            {tab === 'stock-in' && <StockInPage />}
            {tab === 'stock-out' && <StockOutPage />}
            {tab === 'reports' && <ReportsPage />}
            {tab === 'master' && <MasterDataPage />}
          </div>
        </main>

        <nav className="bottom-nav">
          {tabs.map(t => (
            <button
              key={t.id}
              className={tab === t.id ? 'bottom-item active' : 'bottom-item'}
              onClick={() => setTab(t.id)}
            >
              <span className="bottom-label">{t.label}</span>
            </button>
          ))}
        </nav>
      </div>
    </ToastProvider>
  )
}
