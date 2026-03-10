import { useEffect, useState } from 'react'
import { apiGet } from '../api/client.js'
import { useInventoryStore } from '../store/useInventoryStore.js'
import { useToast } from '../components/Toast.jsx'

export default function DocumentsPage() {
  const [stockIns, setStockIns] = useState([])
  const [stockOuts, setStockOuts] = useState([])
  const [loading, setLoading] = useState(false)
  const [quick, setQuick] = useState('')
  const [pageIn, setPageIn] = useState(1)
  const [pageOut, setPageOut] = useState(1)
  const [tab, setTab] = useState('in')
  const [statusFilter, setStatusFilter] = useState('')
  const pageSize = 5
  const { inventory, fetchInventory } = useInventoryStore()
  const { notify } = useToast()

  const statusLabel = (status) => {
    const map = {
      DRAFT: 'Baru',
      ALLOCATED: 'Dibooking',
      CREATED: 'Baru',
      IN_PROGRESS: 'Diproses',
      DONE: 'Selesai',
      CANCELLED: 'Batal'
    }
    return map[status] || status
  }

  const itemLabel = (sku) => {
    const it = inventory.find(i => i.sku === sku)
    return it ? `${it.sku} - ${it.name}` : sku
  }

  const load = async () => {
    setLoading(true)
    try {
      const ins = await apiGet('/stock-ins?status=CREATED,IN_PROGRESS,DONE,CANCELLED')
      const outs = await apiGet('/stock-outs?status=DRAFT,ALLOCATED,IN_PROGRESS,DONE,CANCELLED')
      setStockIns(ins)
      setStockOuts(outs)
    } catch (e) {
      notify('error', e.message || 'Gagal memuat dokumen.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchInventory({})
    load()
    const type = new URLSearchParams(window.location.search).get('type')
    if (type === 'out') setTab('out')
    if (type === 'in') setTab('in')
    document.body.classList.add('documents-page')
    return () => document.body.classList.remove('documents-page')
  }, [])

  const filterTx = (list) => {
    if (!quick) return list
    const q = quick.toLowerCase()
    return list.filter(tx =>
      tx.code?.toLowerCase().includes(q) ||
      tx.status?.toLowerCase().includes(q) ||
      tx.items?.some(i => itemLabel(i.sku).toLowerCase().includes(q))
    )
  }

  const pageSlice = (list, page) => {
    const totalPages = Math.max(1, Math.ceil(list.length / pageSize))
    const safePage = Math.min(page, totalPages)
    const start = (safePage - 1) * pageSize
    return { items: list.slice(start, start + pageSize), totalPages, safePage }
  }

  const statusOptions = tab === 'in'
    ? ['IN_PROGRESS', 'CREATED', 'CANCELLED']
    : ['IN_PROGRESS', 'ALLOCATED', 'DRAFT', 'CANCELLED']

  const applyStatus = (list) => {
    if (!statusFilter || !statusOptions.includes(statusFilter)) return list
    return list.filter(tx => tx.status === statusFilter)
  }

  useEffect(() => {
    if (!statusOptions.includes(statusFilter)) {
      setStatusFilter('')
    }
  }, [tab])

  return (
    <section className="panel">
      <div className="panel-header">Dokumen</div>
      <div className="action-bar">
        <div className="section-title">Ringkasan</div>
        <div className="panel-row">
          <input
            placeholder="Cari cepat transaksi/produk..."
            value={quick}
            onChange={e => { setQuick(e.target.value); setPageIn(1); setPageOut(1) }}
          />
          <button className="btn secondary" onClick={load}>Refresh</button>
        </div>
      </div>
      {loading && <div className="hint">Memuat dokumen...</div>}

      <div className="panel-divider" />
      <div className="action-bar report-tabs">
        <div className="chip-group report-type-chips">
          <button className={`btn ${tab === 'in' ? 'primary' : 'secondary'}`} onClick={() => setTab('in')}>Barang Masuk</button>
          <button className={`btn ${tab === 'out' ? 'primary' : 'secondary'}`} onClick={() => setTab('out')}>Barang Keluar</button>
        </div>
        <div className="chip-group chip-floating">
          {statusOptions.map(s => (
            <button
              key={s}
              className={`btn ${statusFilter === s ? 'primary' : 'secondary'}`}
              onClick={() => { setStatusFilter(s); setPageIn(1); setPageOut(1) }}
            >
              {statusLabel(s)}
            </button>
          ))}
        </div>
      </div>

      {tab === 'in' && (() => {
        const filtered = applyStatus(filterTx(stockIns))
        const { items, totalPages, safePage } = pageSlice(filtered, pageIn)
        return (
          <>
            <ul className="list">
              {items.map(tx => (
                <li key={tx.code}>
                  <div className="row-between">
                    <div className="mono">{tx.code}</div>
                    <span className={`badge ${tx.status?.toLowerCase()}`}>{statusLabel(tx.status)}</span>
                  </div>
                  <div className="muted">Barang: {tx.items?.map(i => `(${itemLabel(i.sku)} x ${i.qty})`).join(', ')}</div>
                </li>
              ))}
            </ul>
            <div className="table-footer">
              <div className="pager">
                <button onClick={() => setPageIn(1)} disabled={safePage === 1}>Awal</button>
                <button onClick={() => setPageIn(safePage - 1)} disabled={safePage === 1}>Sebelum</button>
                <span className="page-pill">Hal {safePage} / {totalPages}</span>
                <button onClick={() => setPageIn(safePage + 1)} disabled={safePage === totalPages}>Berikut</button>
                <button onClick={() => setPageIn(totalPages)} disabled={safePage === totalPages}>Akhir</button>
              </div>
            </div>
          </>
        )
      })()}

      {tab === 'out' && (() => {
        const filtered = applyStatus(filterTx(stockOuts))
        const { items, totalPages, safePage } = pageSlice(filtered, pageOut)
        return (
          <>
            <ul className="list">
              {items.map(tx => (
                <li key={tx.code}>
                  <div className="row-between">
                    <div className="mono">{tx.code}</div>
                    <span className={`badge ${tx.status?.toLowerCase()}`}>{statusLabel(tx.status)}</span>
                  </div>
                  <div className="muted">Barang: {tx.items?.map(i => `(${itemLabel(i.sku)} x ${i.qty})`).join(', ')}</div>
                </li>
              ))}
            </ul>
            <div className="table-footer">
              <div className="pager">
                <button onClick={() => setPageOut(1)} disabled={safePage === 1}>Awal</button>
                <button onClick={() => setPageOut(safePage - 1)} disabled={safePage === 1}>Sebelum</button>
                <span className="page-pill">Hal {safePage} / {totalPages}</span>
                <button onClick={() => setPageOut(safePage + 1)} disabled={safePage === totalPages}>Berikut</button>
                <button onClick={() => setPageOut(totalPages)} disabled={safePage === totalPages}>Akhir</button>
              </div>
            </div>
          </>
        )
      })()}
    </section>
  )
}
