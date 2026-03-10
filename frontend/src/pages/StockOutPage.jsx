import { useEffect, useState } from 'react'
import { apiDelete, apiGet, apiPost } from '../api/client.js'
import { useInventoryStore } from '../store/useInventoryStore.js'
import Modal from '../components/Modal.jsx'
import { useToast } from '../components/Toast.jsx'
import ConfirmDialog from '../components/ConfirmDialog.jsx'

export default function StockOutPage() {
  const [code, setCode] = useState('')
  const [items, setItems] = useState([])
  const [line, setLine] = useState({ sku: '', qty: '' })
  const [allocateCode, setAllocateCode] = useState('')
  const [statusUpdate, setStatusUpdate] = useState({ code: '', status: 'IN_PROGRESS' })
  const [lookupCode, setLookupCode] = useState('')
  const [lookupData, setLookupData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [pending, setPending] = useState([])
  const [pendingLoading, setPendingLoading] = useState(false)
  const [pendingError, setPendingError] = useState('')
  const [finalDialog, setFinalDialog] = useState({ open: false, status: '' })
  const [deleteDialog, setDeleteDialog] = useState({ open: false, code: '' })
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [allocateDialogOpen, setAllocateDialogOpen] = useState(false)
  const { inventory, fetchInventory } = useInventoryStore()
  const { notify } = useToast()

  const allowedNext = (from) => {
    const map = {
      DRAFT: ['ALLOCATED', 'CANCELLED'],
      ALLOCATED: ['IN_PROGRESS', 'CANCELLED'],
      IN_PROGRESS: ['DONE', 'CANCELLED'],
      DONE: [],
      CANCELLED: []
    }
    return map[from] || []
  }

  const statusLabel = (status) => {
    const map = {
      DRAFT: 'Baru',
      ALLOCATED: 'Dibooking',
      IN_PROGRESS: 'Diproses',
      DONE: 'Selesai',
      CANCELLED: 'Batal'
    }
    return map[status] || status
  }

  const statusInfo = (status) => {
    const map = {
      DRAFT: 'Dokumen baru. Lanjutkan dengan booking stok.',
      ALLOCATED: 'Stok sudah dibooking. Lanjutkan ke proses.',
      IN_PROGRESS: 'Proses berjalan. Jika dibatalkan, stok siap kembali.',
      DONE: 'Selesai: barang sudah keluar, stok gudang berkurang.',
      CANCELLED: 'Dibatalkan: stok siap kembali.'
    }
    return map[status] || ''
  }

  const nextAction = (status) => {
    const map = {
      DRAFT: 'Aksi berikutnya: Booking stok.',
      ALLOCATED: 'Aksi berikutnya: Mulai proses.',
      IN_PROGRESS: 'Aksi berikutnya: Selesaikan atau Batalkan.',
      DONE: 'Tidak ada aksi lanjutan.',
      CANCELLED: 'Tidak ada aksi lanjutan.'
    }
    return map[status] || ''
  }

  const loadPending = async () => {
    setPendingLoading(true)
    try {
      const data = await apiGet('/stock-outs?status=DRAFT,ALLOCATED,IN_PROGRESS')
      setPending(data)
      setPendingError('')
    } catch (e) {
      setPendingError(e.message || 'Gagal memuat dokumen.')
    } finally {
      setPendingLoading(false)
    }
  }

  const pickPending = async (docCode) => {
    setLookupCode(docCode)
    await onLookup(docCode)
  }

  const addLine = () => {
    const qty = Number(line.qty)
    if (!line.sku || !qty || qty <= 0) {
      notify('error', 'Pilih barang dan isi jumlah minimal 1.')
      return
    }
    if (items.some(it => it.sku === line.sku)) {
      notify('error', 'Barang sudah ada di daftar. Hapus dulu lalu pilih ulang dengan jumlah yang benar.')
      return
    }
    setItems([...items, { sku: line.sku, qty }])
    setLine({ sku: '', qty: '' })
  }

  const removeLine = (sku) => {
    setItems(items.filter(it => it.sku !== sku))
  }

  const nameBySku = (sku) => {
    const found = inventory.find(it => it.sku === sku)
    return found?.name || ''
  }

  const itemsSummary = () => items
    .map(it => `${it.sku}${nameBySku(it.sku) ? ` - ${nameBySku(it.sku)}` : ''} x ${it.qty}`)
    .join(', ')

  const submit = () => {
    if (items.length === 0) {
      notify('error', 'Item wajib diisi.')
      return
    }
    setCreateDialogOpen(true)
  }

  const confirmCreate = async () => {
    setCreateDialogOpen(false)
    setLoading(true)
    try {
      const res = await apiPost('/stock-outs', { items })
      notify('success', `Berhasil dibuat. Nomor Dokumen: ${res.code}`)
      setCode('')
      setItems([])
      setCreateOpen(false)
      await loadPending()
    } catch (e) {
      notify('error', e.message || 'Gagal membuat barang keluar.')
    } finally {
      setLoading(false)
    }
  }

  const allocate = () => {
    if (!allocateCode) {
      notify('error', 'Nomor dokumen wajib diisi.')
      return
    }
    setAllocateDialogOpen(true)
  }

  const confirmAllocate = async () => {
    setAllocateDialogOpen(false)
    setLoading(true)
    try {
      await apiPost(`/stock-outs/${allocateCode}/allocate`, {})
      notify('success', 'Booking stok berhasil')
      if (lookupData?.code === allocateCode) {
        await onLookup()
      }
      await loadPending()
    } catch (e) {
      notify('error', e.message || 'Gagal booking stok.')
    } finally {
      setLoading(false)
    }
  }

  const updateStatus = () => {
    if (!statusUpdate.code) {
      notify('error', 'Nomor dokumen wajib diisi.')
      return
    }
    setStatusDialogOpen(true)
  }

  const confirmStatus = async () => {
    setStatusDialogOpen(false)
    setLoading(true)
    try {
      await apiPost(`/stock-outs/${statusUpdate.code}/status`, { status: statusUpdate.status })
      if (statusUpdate.status === 'DONE' || statusUpdate.status === 'CANCELLED') {
        setFinalDialog({ open: true, status: statusUpdate.status })
        return
      }
      notify('success', 'Status berhasil diperbarui')
      if (lookupData?.code === statusUpdate.code) {
        await onLookup()
      }
      await loadPending()
    } catch (e) {
      notify('error', e.message || 'Gagal update status.')
    } finally {
      setLoading(false)
    }
  }

  const onLookup = async (overrideCode) => {
    const codeToUse = overrideCode || lookupCode || ''
    if (!codeToUse) {
      notify('error', 'Nomor dokumen wajib diisi.')
      return
    }
    setLoading(true)
    try {
      const data = await apiGet(`/stock-outs/code/${codeToUse}`)
      setLookupData(data)
      setAllocateCode(data.code)
      const next = allowedNext(data.status)
      setStatusUpdate({ code: data.code, status: next[0] || data.status })
    } catch (e) {
      setLookupData(null)
      notify('error', e.message || 'Dokumen tidak ditemukan.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchInventory({})
    loadPending()
  }, [])

  const deleteDoc = async (docCode) => {
    try {
      await apiDelete(`/stock-outs/${docCode}`)
      await loadPending()
      notify('success', `Dokumen ${docCode} dihapus.`)
    } catch (e) {
      notify('error', e.message || 'Gagal menghapus.')
    }
  }

  const openDelete = (docCode) => {
    setDeleteDialog({ open: true, code: docCode })
  }

  const confirmDelete = async () => {
    const codeToDelete = deleteDialog.code
    setDeleteDialog({ open: false, code: '' })
    if (!codeToDelete) return
    await deleteDoc(codeToDelete)
  }

  const closeCreate = () => {
    setCreateOpen(false)
    setItems([])
    setLine({ sku: '', qty: '' })
  }

  const isTerminal = lookupData && (lookupData.status === 'DONE' || lookupData.status === 'CANCELLED')

  return (
    <section className="panel">
      <div className="panel-header">Barang Keluar</div>
      <div className="hint">Tahap 1: booking stok. Tahap 2: proses barang keluar sampai selesai. Jika dibatalkan saat proses berjalan, stok siap otomatis kembali.</div>
      <div className="action-bar">
        <div className="section-title">Dokumen</div>
        <div className="panel-row">
          <button className="btn primary" onClick={() => setCreateOpen(true)}>Buat Barang Keluar</button>
          <button className="btn secondary" onClick={loadPending}>Muat Dokumen Pending</button>
        </div>
      </div>

      {pendingLoading && <div className="hint">Memuat dokumen...</div>}
      {pendingError && <div className="error">{pendingError}</div>}
      {pending.length > 0 && (
        <div className="card">
          <div className="card-title">Dokumen Pending</div>
          <ul className="list">
            {pending.map(tx => (
              <li key={tx.code} className="row-between">
                <div>
                  <div className="mono">{tx.code}</div>
                  <div className="muted">Status: <span className={`badge ${tx.status?.toLowerCase()}`}>{statusLabel(tx.status)}</span></div>
                  <div className="muted">{nextAction(tx.status)}</div>
                </div>
                <div className="panel-row">
                  <button className="btn secondary" onClick={() => pickPending(tx.code)}>Lanjutkan</button>
                  <button className="btn danger" onClick={() => openDelete(tx.code)}>Hapus</button>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
      {!pendingLoading && !pendingError && pending.length === 0 && (
        <div className="hint">Belum ada dokumen yang perlu diproses.</div>
      )}

      <div className="panel-divider" />

      <div className="panel-header">Pilih Dokumen</div>
      <div className="hint">Pilih dokumen dari daftar pending untuk diproses.</div>
      {lookupData && (
        <div className="card">
          <div className="card-title">Dokumen {lookupData.code}</div>
          <div className="muted">Status: <span className={`badge ${lookupData.status?.toLowerCase()}`}>{statusLabel(lookupData.status)}</span></div>
          <div className="muted">{statusInfo(lookupData.status)}</div>
          <div className="muted">{nextAction(lookupData.status)}</div>
          <div className="muted">Barang: {lookupData.items?.map(it => `${it.sku} x ${it.qty}`).join(', ')}</div>
        </div>
      )}
      <div className="panel-header">Tahap 1: Booking Stok</div>
      <div className="panel-row">
        <input placeholder="Nomor Dokumen" value={allocateCode} disabled />
        <button className="btn primary" onClick={allocate} disabled={loading || (lookupData && lookupData.status !== 'DRAFT')}>
          {loading ? 'Memproses...' : 'Booking Stok'}
        </button>
      </div>
      {lookupData && lookupData.status !== 'DRAFT' && (
        <div className="hint">Booking hanya tersedia untuk dokumen yang baru dibuat.</div>
      )}

      <div className="panel-divider" />

      <div className="panel-header">Tahap 2: Proses</div>
      <div className="panel-row">
        <input placeholder="Nomor Dokumen" value={statusUpdate.code} disabled />
        <select
          value={statusUpdate.status}
          onChange={e => setStatusUpdate({ ...statusUpdate, status: e.target.value })}
          disabled={!lookupData || isTerminal}
        >
          {(lookupData ? allowedNext(lookupData.status) : []).map(opt => (
            <option key={opt} value={opt}>{opt}</option>
          ))}
        </select>
        <button className="btn primary" onClick={updateStatus} disabled={loading || !lookupData || isTerminal}>
          {loading ? 'Memproses...' : 'Ubah Status'}
        </button>
      </div>
      {isTerminal && <div className="hint">Dokumen sudah final. Tidak ada perubahan status.</div>}
      {!lookupData && <div className="hint">Pilih dokumen dulu agar opsi status sesuai alur two-phase.</div>}

      <Modal
        open={createOpen}
        title="Buat Barang Keluar"
        onClose={closeCreate}
        actions={<button className="btn primary" onClick={submit} disabled={loading}>Simpan</button>}
      >
        <div className="panel-row">
          <input placeholder="Nomor Dokumen (otomatis)" value={code} disabled />
        </div>
        <div className="panel-row">
          <select value={line.sku} onChange={e => setLine({ ...line, sku: e.target.value })}>
            <option value="">Pilih barang (SKU - Nama)</option>
            {inventory.map(it => (
              <option key={it.sku} value={it.sku}>
                {it.sku} - {it.name}
              </option>
            ))}
          </select>
          <input placeholder="Jumlah keluar" value={line.qty} onChange={e => setLine({ ...line, qty: e.target.value })} />
          <button className="btn secondary" onClick={addLine} disabled={loading}>Tambah</button>
        </div>
        <ul className="list">
          {items.map((it) => (
            <li key={it.sku} className="row-between">
              <span>Barang {it.sku}{nameBySku(it.sku) ? ` - ${nameBySku(it.sku)}` : ''} x {it.qty}</span>
              <button className="btn danger" onClick={() => removeLine(it.sku)} disabled={loading}>Hapus</button>
            </li>
          ))}
        </ul>
        {inventory.length === 0 && (
          <div className="hint">Daftar barang belum terisi. Buka menu Master Data lalu tambahkan produk.</div>
        )}
      </Modal>

      <ConfirmDialog
        open={finalDialog.open}
        title="Status Tersimpan"
        message={`Status dokumen sudah menjadi ${statusLabel(finalDialog.status)}. Halaman akan diperbarui.`}
        confirmText="Perbarui Halaman"
        confirmClass="primary"
        onConfirm={() => window.location.reload()}
        onClose={() => setFinalDialog({ open: false, status: '' })}
      />

      <ConfirmDialog
        open={createDialogOpen}
        title="Simpan Dokumen"
        message={`Simpan dokumen barang keluar dengan item: ${itemsSummary() || '-' }?`}
        confirmText="Simpan"
        confirmClass="primary"
        onConfirm={confirmCreate}
        onClose={() => setCreateDialogOpen(false)}
      />

      <ConfirmDialog
        open={allocateDialogOpen}
        title="Booking Stok"
        message={`Booking stok untuk dokumen ${allocateCode}?`}
        confirmText="Booking"
        confirmClass="primary"
        onConfirm={confirmAllocate}
        onClose={() => setAllocateDialogOpen(false)}
      />

      <ConfirmDialog
        open={statusDialogOpen}
        title="Ubah Status"
        message={`Ubah status dokumen ${statusUpdate.code} ke ${statusLabel(statusUpdate.status)}?`}
        confirmText="Update"
        confirmClass="primary"
        onConfirm={confirmStatus}
        onClose={() => setStatusDialogOpen(false)}
      />

      <ConfirmDialog
        open={deleteDialog.open}
        title="Hapus Dokumen"
        message={`Hapus dokumen ${deleteDialog.code}?`}
        confirmText="Hapus"
        confirmClass="danger"
        showNote
        onConfirm={confirmDelete}
        onClose={() => setDeleteDialog({ open: false, code: '' })}
      />
    </section>
  )
}
