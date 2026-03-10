import { useEffect, useState } from 'react'
import { useInventoryStore } from '../store/useInventoryStore.js'
import { apiGet } from '../api/client.js'
import Modal from '../components/Modal.jsx'
import { useToast } from '../components/Toast.jsx'
import ConfirmDialog from '../components/ConfirmDialog.jsx'

export default function InventoryPage() {
  const { inventory, loading, error, fetchInventory, adjustStock } = useInventoryStore()
  const [filter, setFilter] = useState({ category: '', customer: '' })
  const [adjust, setAdjust] = useState({ sku: '', new_physical: '', reason: '' })
  const [adjustOpen, setAdjustOpen] = useState(false)
  const [logOpen, setLogOpen] = useState(false)
  const [logSku, setLogSku] = useState('')
  const [logItems, setLogItems] = useState([])
  const [logLoading, setLogLoading] = useState(false)
  const [logError, setLogError] = useState('')
  const [adjustDialogOpen, setAdjustDialogOpen] = useState(false)
  const [quick, setQuick] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const { notify } = useToast()

  useEffect(() => {
    fetchInventory({})
  }, [fetchInventory])

  const onSearch = () => {
    fetchInventory({ customer: filter.customer })
    setPage(1)
  }

  const onAdjust = () => {
    if (!adjust.sku || !adjust.new_physical) {
      notify('error', 'Pilih barang dan isi stok gudang baru.')
      return
    }
    setAdjustDialogOpen(true)
  }

  const confirmAdjust = async () => {
    setAdjustDialogOpen(false)
    try {
      await adjustStock({
        sku: adjust.sku,
        new_physical: Number(adjust.new_physical),
        reason: adjust.reason
      })
      setAdjust({ sku: '', new_physical: '', reason: '' })
      notify('success', 'Penyesuaian stok berhasil disimpan.')
      setAdjustOpen(false)
    } catch (e) {
      notify('error', e.message || 'Gagal menyimpan penyesuaian stok.')
    }
  }

  const filtered = inventory.filter(row => {
    if (!quick) return true
    const q = quick.toLowerCase()
    return (
      row.sku?.toLowerCase().includes(q) ||
      row.name?.toLowerCase().includes(q) ||
      row.customer?.toLowerCase().includes(q) ||
      row.category?.toLowerCase().includes(q)
    )
  })

  const filteredBySelect = filtered.filter(row => {
    if (filter.category && row.category !== filter.category) return false
    if (filter.customer && row.customer !== filter.customer) return false
    return true
  })

  const totalPages = Math.max(1, Math.ceil(filteredBySelect.length / pageSize))
  const safePage = Math.min(page, totalPages)
  const start = (safePage - 1) * pageSize
  const pageItems = filteredBySelect.slice(start, start + pageSize)

  const openAdjust = (sku) => {
    setAdjust({ sku, new_physical: '', reason: '' })
    setAdjustOpen(true)
  }

  const openLogs = async (sku) => {
    setLogSku(sku)
    setLogOpen(true)
    setLogLoading(true)
    setLogError('')
    try {
      const data = await apiGet(`/inventory/adjustments?sku=${encodeURIComponent(sku)}`)
      setLogItems(data)
    } catch (e) {
      setLogError(e.message || 'Gagal memuat log.')
      setLogItems([])
    } finally {
      setLogLoading(false)
    }
  }

  const unique = (arr) => Array.from(new Set(arr.filter(Boolean)))
  const categoryOptions = unique(inventory.map(i => i.category)).sort()
  const customerOptions = unique(inventory.map(i => i.customer)).sort()

  const sourceLabel = (source) => {
    const map = {
      MANUAL: 'Manual',
      STOCK_IN_DONE: 'Masuk',
      STOCK_OUT_ALLOCATED: 'Booking',
      STOCK_OUT_DONE: 'Keluar',
      STOCK_OUT_CANCELLED: 'Batal Booking'
    }
    return map[source] || source || '-'
  }

  const bucketLabel = (bucket) => (bucket === 'AVAILABLE' ? 'Stok Siap' : 'Stok Gudang')

  return (
    <section className="panel">
      <div className="panel-header">Daftar Barang</div>
      <div className="section">
        <div className="section-title">Filter Barang</div>
        <div className="panel-row">
          <select value={filter.category} onChange={e => setFilter({ ...filter, category: e.target.value })}>
            <option value="">Semua Kategori</option>
            {categoryOptions.map(c => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          <select value={filter.customer} onChange={e => setFilter({ ...filter, customer: e.target.value })}>
            <option value="">Semua Customer</option>
            {customerOptions.map(c => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </div>
        <div className="panel-row">
          <input
            placeholder="Cari cepat (SKU/Nama/Customer)..."
            value={quick}
            onChange={e => { setQuick(e.target.value); setPage(1) }}
          />
          <div className="pill">{filteredBySelect.length} / {inventory.length} barang</div>
        </div>
      </div>
      <div className="panel-body">
        {loading && <div className="hint">Memuat data...</div>}
        {error && <div className="error">{error}</div>}
        <div className="action-bar">
          <div className="section-title">Data Barang</div>
          <button className="btn secondary" onClick={() => fetchInventory({})} disabled={loading}>Refresh</button>
        </div>
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>SKU</th>
                <th>Nama Barang</th>
                <th>Kategori</th>
                <th>Customer</th>
                <th>Stok Gudang</th>
                <th>Stok Siap</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {pageItems.map(row => (
                <tr key={row.sku}>
                  <td>{row.sku}</td>
                  <td>{row.name}</td>
                  <td>{row.category || '-'}</td>
                  <td>{row.customer}</td>
                  <td>{row.physical_qty}</td>
                  <td>{row.available_qty}</td>
                  <td>
                    <button className="btn secondary" onClick={() => openAdjust(row.sku)}>Sesuaikan</button>
                    <button className="btn secondary" onClick={() => openLogs(row.sku)}>Riwayat</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="table-footer">
          <div className="rows">
            <span>Tampilkan</span>
            <select value={pageSize} onChange={e => { setPageSize(Number(e.target.value)); setPage(1) }}>
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={20}>20</option>
            </select>
            <span>baris</span>
          </div>
          <div className="pager">
            <button onClick={() => setPage(1)} disabled={safePage === 1}>Awal</button>
            <button onClick={() => setPage(safePage - 1)} disabled={safePage === 1}>Sebelum</button>
            <span className="page-pill">Hal {safePage} / {totalPages}</span>
            <button onClick={() => setPage(safePage + 1)} disabled={safePage === totalPages}>Berikut</button>
            <button onClick={() => setPage(totalPages)} disabled={safePage === totalPages}>Akhir</button>
          </div>
        </div>
      </div>

      <Modal
        open={adjustOpen}
        title="Penyesuaian Stok"
        onClose={() => setAdjustOpen(false)}
        actions={<button className="btn primary" onClick={onAdjust} disabled={loading}>Simpan</button>}
      >
        <div className="panel-row">
          <select value={adjust.sku} onChange={e => setAdjust({ ...adjust, sku: e.target.value })}>
            <option value="">Pilih barang (SKU - Nama)</option>
            {inventory.map(it => (
              <option key={it.sku} value={it.sku}>
                {it.sku} - {it.name}
              </option>
            ))}
          </select>
          <input placeholder="Stok gudang baru" value={adjust.new_physical} onChange={e => setAdjust({ ...adjust, new_physical: e.target.value })} />
          <input placeholder="Catatan (contoh: opname)" value={adjust.reason} onChange={e => setAdjust({ ...adjust, reason: e.target.value })} />
        </div>
      </Modal>

      <ConfirmDialog
        open={adjustDialogOpen}
        title="Simpan Penyesuaian"
        message={`Simpan penyesuaian untuk SKU ${adjust.sku} menjadi ${adjust.new_physical}?`}
        confirmText="Simpan"
        confirmClass="primary"
        onConfirm={confirmAdjust}
        onClose={() => setAdjustDialogOpen(false)}
      />

      <Modal
        open={logOpen}
        title={`Riwayat Stok - ${logSku}`}
        onClose={() => setLogOpen(false)}
        actions={<button className="btn secondary" onClick={() => setLogOpen(false)}>Tutup</button>}
      >
        {logLoading && <div className="hint">Memuat log...</div>}
        {logError && <div className="error">{logError}</div>}
        {!logLoading && !logError && logItems.length === 0 && (
          <div className="hint">Belum ada riwayat stok.</div>
        )}
        {logItems.length > 0 && (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Tanggal</th>
                  <th>Stok</th>
                  <th>Perubahan</th>
                  <th>Sumber</th>
                  <th>Catatan</th>
                </tr>
              </thead>
              <tbody>
                {logItems.map((it, idx) => (
                  <tr key={idx}>
                    <td>{new Date(it.created_at).toLocaleString()}</td>
                    <td>{bucketLabel(it.bucket)}</td>
                    <td className="mono">{it.delta > 0 ? `+${it.delta}` : it.delta}</td>
                    <td>{it.ref_code ? `${sourceLabel(it.source)} (${it.ref_code})` : sourceLabel(it.source)}</td>
                    <td>{it.reason || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Modal>
    </section>
  )
}
