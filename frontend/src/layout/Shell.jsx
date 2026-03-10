import '../styles.css'

const tabs = [
  { id: 'inventory', label: 'Stok Barang', href: '/inventory', icon: '📦' },
  { id: 'stock-in', label: 'Barang Masuk', href: '/stock-in', icon: '⬇️' },
  { id: 'stock-out', label: 'Barang Keluar', href: '/stock-out', icon: '⬆️' },
  { id: 'documents', label: 'Dokumen', href: '/documents', icon: '🧾' },
  { id: 'reports', label: 'Laporan', href: '/reports', icon: '📊' },
  { id: 'master', label: 'Master Data', href: '/master', icon: '🗂️' }
]

export default function Shell({ active, children }) {
  const current = tabs.find(t => t.id === active)
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <img className="brand-logo" src="/assets/logo.png" alt="Tirtamas logo" />
          <div className="brand-title">Tirtamas Inventory</div>
          <div className="brand-sub">Smart Stock Control</div>
        </div>
        <div className="sidebar-title">Menu Utama</div>
        <nav className="side-nav">
          {tabs.map(t => (
            <a key={t.id} className={active === t.id ? 'nav-item active' : 'nav-item'} href={t.href}>
              <span className="nav-icon">{t.icon}</span>
              <span className="nav-label">{t.label}</span>
            </a>
          ))}
        </nav>
      </aside>

      <main className="content">
        <header className="content-header">
          <div>
            <div className="breadcrumb">Beranda / {current?.label}</div>
            <div className="content-title">{current?.label}</div>
          </div>
          <div className="toolbar">
            <span className="chip">Admin Gudang</span>
          </div>
        </header>
        <div className="content-body">{children}</div>
      </main>

      <nav className="bottom-nav">
        {tabs.map(t => (
          <a key={t.id} className={active === t.id ? 'bottom-item active' : 'bottom-item'} href={t.href}>
            <span className="bottom-icon">{t.icon}</span>
            <span className="bottom-label">{t.label}</span>
          </a>
        ))}
      </nav>
    </div>
  )
}
