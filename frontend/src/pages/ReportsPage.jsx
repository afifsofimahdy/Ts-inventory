import { useEffect, useState } from 'react'
import { apiGet } from '../api/client.js'
import { useInventoryStore } from '../store/useInventoryStore.js'
import { useToast } from '../components/Toast.jsx'
import Modal from '../components/Modal.jsx'

export default function ReportsPage() {
  const [stockIns, setStockIns] = useState([])
  const [stockOuts, setStockOuts] = useState([])
  const [loading, setLoading] = useState(false)
  const [quick, setQuick] = useState('')
  const [pageIn, setPageIn] = useState(1)
  const [pageOut, setPageOut] = useState(1)
  const [tab, setTab] = useState('in')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [sjOpen, setSjOpen] = useState(false)
  const [sjData, setSjData] = useState(null)
  const [sjType, setSjType] = useState('in')
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
      const qs = new URLSearchParams()
      if (dateFrom) qs.set('date_from', dateFrom)
      if (dateTo) qs.set('date_to', dateTo)
      const ins = await apiGet(`/reports/stock-ins?${qs.toString()}`)
      const outs = await apiGet(`/reports/stock-outs?${qs.toString()}`)
      setStockIns(ins)
      setStockOuts(outs)
    } catch (e) {
      notify('error', e.message || 'Gagal memuat laporan.')
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

  const openSuratJalan = (tx) => {
    setSjData(tx)
    setSjType(tab)
    setSjOpen(true)
  }

  const printSj = () => {
    if (!sjData) return
    const title = sjType === 'out' ? 'BUKTI KELUAR BARANG' : 'BUKTI MASUK BARANG'
    const html = `
      <!doctype html>
      <html>
        <head>
          <meta charset="utf-8" />
          <title>${title}</title>
          <style>
            @page { size: A4; margin: 18mm; }
            * { box-sizing: border-box; }
            body { font-family: Arial, sans-serif; color: #0f172a; }
            .wrap { border: 1px solid #e2e8f0; padding: 16px; }
            .header { display: flex; justify-content: space-between; gap: 16px; border-bottom: 1px dashed #cbd5e1; padding-bottom: 10px; margin-bottom: 16px; }
            .logo { width: 64px; height: 64px; object-fit: contain; }
            .title { font-size: 20px; font-weight: 800; letter-spacing: 0.8px; margin-top: 6px; }
            .sub { color: #64748b; font-size: 12px; }
            .meta { text-align: right; font-size: 12px; }
            .table { width: 100%; border-collapse: collapse; margin-top: 10px; }
            .table th, .table td { border: 1px solid #e2e8f0; padding: 8px; font-size: 12px; }
            .table th { background: #f1f5f9; text-transform: uppercase; letter-spacing: 0.4px; }
            .sign { display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px; margin-top: 28px; }
            .sign-label { font-size: 12px; color: #64748b; margin-bottom: 36px; }
            .sign-line { border-bottom: 1px solid #e2e8f0; height: 24px; }
          </style>
        </head>
        <body>
          <div class="wrap">
            <div class="header">
              <div>
                <img class="logo" src="/assets/logo.png" alt="Tirtamas logo" />
                <div class="title">${title}</div>
                <div class="sub">Gudang Tirtamas</div>
              </div>
              <div class="meta">
                <div>Nomor: ${sjData.code}</div>
                <div>Tanggal: ${sjData.created_at ? new Date(sjData.created_at).toLocaleDateString() : '-'}</div>
                <div>Status: ${statusLabel(sjData.status)}</div>
              </div>
            </div>
            <table class="table">
              <thead>
                <tr>
                  <th style="width:48px;">No</th>
                  <th>Barang</th>
                  <th style="width:80px;">Qty</th>
                </tr>
              </thead>
              <tbody>
                ${sjData.items?.map((it, idx) => `
                  <tr>
                    <td>${idx + 1}</td>
                    <td>${itemLabel(it.sku)}</td>
                    <td>${it.qty}</td>
                  </tr>
                `).join('') || ''}
              </tbody>
            </table>
            <div class="sign">
              <div>
                <div class="sign-label">Diserahkan</div>
                <div class="sign-line"></div>
              </div>
              <div>
                <div class="sign-label">Diterima</div>
                <div class="sign-line"></div>
              </div>
              <div>
                <div class="sign-label">Mengetahui</div>
                <div class="sign-line"></div>
              </div>
            </div>
          </div>
        </body>
      </html>
    `
    const win = window.open('', '_blank')
    if (!win) return
    win.document.open()
    win.document.write(html)
    win.document.close()
    win.focus()
    win.print()
  }

  return (
    <section className="panel">
      <div className="panel-header">Laporan</div>
      <div className="action-bar">
        <div className="section-title">Ringkasan</div>
        <div className="panel-row">
          <input
            placeholder="Cari cepat transaksi/produk..."
            value={quick}
            onChange={e => { setQuick(e.target.value); setPageIn(1); setPageOut(1) }}
          />
          <input type="date" value={dateFrom} onChange={e => setDateFrom(e.target.value)} />
          <input type="date" value={dateTo} onChange={e => setDateTo(e.target.value)} />
          <button className="btn secondary" onClick={load}>Terapkan</button>
        </div>
      </div>
      {loading && <div className="hint">Memuat laporan...</div>}

      <div className="panel-divider" />
      <div className="action-bar report-tabs">
        <div className="chip-group">
          <button className={`btn ${tab === 'in' ? 'primary' : 'secondary'}`} onClick={() => setTab('in')}>Barang Masuk</button>
          <button className={`btn ${tab === 'out' ? 'primary' : 'secondary'}`} onClick={() => setTab('out')}>Barang Keluar</button>
        </div>
      </div>

      {tab === 'in' && (() => {
        const filtered = filterTx(stockIns)
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
                  <div className="panel-row">
                    <button className="btn secondary" onClick={() => openSuratJalan(tx)}>Lihat Dokumen</button>
                  </div>
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
        const filtered = filterTx(stockOuts)
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
                  <div className="panel-row">
                    <button className="btn secondary" onClick={() => openSuratJalan(tx)}>Lihat Dokumen</button>
                  </div>
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

      <Modal
        open={sjOpen}
        title={sjType === 'out' ? 'Bukti Keluar Barang' : 'Bukti Masuk Barang'}
        onClose={() => setSjOpen(false)}
        actions={(
          <>
            <button className="btn secondary" onClick={() => setSjOpen(false)}>Tutup</button>
            <button className="btn primary" onClick={printSj}>Cetak PDF</button>
          </>
        )}
      >
        {!sjData && <div className="hint">Data surat jalan tidak tersedia.</div>}
        {sjData && (
          <div className="sj print-area">
            <div className="sj-header">
              <div>
                <img className="sj-logo" src="/assets/logo.png" alt="Tirtamas logo" />
                <div className="sj-title">{sjType === 'out' ? 'BUKTI KELUAR BARANG' : 'BUKTI MASUK BARANG'}</div>
                <div className="sj-sub">Gudang Tirtamas</div>
              </div>
              <div className="sj-meta">
                <div>Nomor: <span className="mono">{sjData.code}</span></div>
                <div>Tanggal: {sjData.created_at ? new Date(sjData.created_at).toLocaleDateString() : '-'}</div>
                <div>Status: {statusLabel(sjData.status)}</div>
              </div>
            </div>

            <div className="sj-table">
              <div className="sj-row sj-head">
                <div>No</div>
                <div>Barang</div>
                <div>Qty</div>
              </div>
              {sjData.items?.map((it, idx) => (
                <div key={idx} className="sj-row">
                  <div>{idx + 1}</div>
                  <div>{itemLabel(it.sku)}</div>
                  <div>{it.qty}</div>
                </div>
              ))}
            </div>

            <div className="sj-sign">
              <div>
                <div className="sj-sign-label">Diserahkan</div>
                <div className="sj-sign-line" />
              </div>
              <div>
                <div className="sj-sign-label">Diterima</div>
                <div className="sj-sign-line" />
              </div>
              <div>
                <div className="sj-sign-label">Mengetahui</div>
                <div className="sj-sign-line" />
              </div>
            </div>
          </div>
        )}
      </Modal>
    </section>
  )
}
