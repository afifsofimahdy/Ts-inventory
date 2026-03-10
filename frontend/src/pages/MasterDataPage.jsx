import { useEffect, useState } from 'react'
import { apiGet, apiPost } from '../api/client.js'
import Modal from '../components/Modal.jsx'
import { useToast } from '../components/Toast.jsx'
import ConfirmDialog from '../components/ConfirmDialog.jsx'

const tabs = [
  { id: 'customers', label: 'Customer' },
  { id: 'categories', label: 'Kategori' },
  { id: 'products', label: 'Produk' }
]

export default function MasterDataPage() {
  const [tab, setTab] = useState('customers')
  const [customers, setCustomers] = useState([])
  const [categories, setCategories] = useState([])
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [modal, setModal] = useState('')
  const { notify } = useToast()
  const [confirmDialog, setConfirmDialog] = useState({ open: false, type: '', message: '' })

  const [newCustomer, setNewCustomer] = useState('')
  const [newCategory, setNewCategory] = useState('')
  const [newProduct, setNewProduct] = useState({ sku: '', name: '', customer_id: '', category_id: '' })

  const loadAll = async () => {
    setLoading(true)
    try {
      const [c, k, p] = await Promise.all([
        apiGet('/customers'),
        apiGet('/categories'),
        apiGet('/products')
      ])
      setCustomers(c)
      setCategories(k)
      setProducts(p)
      setError('')
    } catch (e) {
      setError(e.message)
      notify('error', e.message || 'Gagal memuat master data.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadAll()
  }, [])

  const addCustomer = () => {
    if (!newCustomer) return
    setConfirmDialog({ open: true, type: 'customer', message: `Tambah customer "${newCustomer}"?` })
  }

  const confirmAddCustomer = async () => {
    setConfirmDialog({ open: false, type: '', message: '' })
    setLoading(true)
    try {
      await apiPost('/customers', { name: newCustomer })
      setNewCustomer('')
      await loadAll()
      notify('success', 'Customer berhasil ditambahkan.')
      setModal('')
    } catch (e) {
      notify('error', e.message || 'Gagal menambah customer.')
    } finally {
      setLoading(false)
    }
  }

  const addCategory = () => {
    if (!newCategory) return
    setConfirmDialog({ open: true, type: 'category', message: `Tambah kategori "${newCategory}"?` })
  }

  const confirmAddCategory = async () => {
    setConfirmDialog({ open: false, type: '', message: '' })
    setLoading(true)
    try {
      await apiPost('/categories', { name: newCategory })
      setNewCategory('')
      await loadAll()
      notify('success', 'Kategori berhasil ditambahkan.')
      setModal('')
    } catch (e) {
      notify('error', e.message || 'Gagal menambah kategori.')
    } finally {
      setLoading(false)
    }
  }

  const addProduct = () => {
    if (!newProduct.sku || !newProduct.name || !newProduct.customer_id || !newProduct.category_id) return
    setConfirmDialog({
      open: true,
      type: 'product',
      message: `Tambah produk "${newProduct.name}" (${newProduct.sku})?`
    })
  }

  const confirmAddProduct = async () => {
    setConfirmDialog({ open: false, type: '', message: '' })
    setLoading(true)
    try {
      await apiPost('/products', {
        sku: newProduct.sku,
        name: newProduct.name,
        customer_id: Number(newProduct.customer_id),
        category_id: Number(newProduct.category_id)
      })
      setNewProduct({ sku: '', name: '', customer_id: '', category_id: '' })
      await loadAll()
      notify('success', 'Produk berhasil ditambahkan.')
      setModal('')
    } catch (e) {
      notify('error', e.message || 'Gagal menambah produk.')
    } finally {
      setLoading(false)
    }
  }

  const closeModal = () => {
    setModal('')
    setNewCustomer('')
    setNewCategory('')
    setNewProduct({ sku: '', name: '', customer_id: '', category_id: '' })
  }

  return (
    <section className="panel">
      <div className="panel-header">Master Data</div>
      <div className="action-bar">
        <div className="section-title">Kategori Data</div>
        <div className="subtabs">
        {tabs.map(t => (
          <button key={t.id} className={tab === t.id ? 'subtab active' : 'subtab'} onClick={() => setTab(t.id)}>
            {t.label}
          </button>
        ))}
        </div>
      </div>

      {error && <div className="error">{error}</div>}
      {loading && <div className="hint">Memuat data...</div>}

      {tab === 'customers' && (
        <>
          <div className="action-bar">
            <div className="section-title">Data Customer</div>
            <button className="btn primary" onClick={() => setModal('customer')}>Tambah Customer</button>
          </div>
          <ul className="list">
            {customers.map(c => (
              <li key={c.id}>{c.name}</li>
            ))}
          </ul>
        </>
      )}

      {tab === 'categories' && (
        <>
          <div className="action-bar">
            <div className="section-title">Data Kategori</div>
            <button className="btn primary" onClick={() => setModal('category')}>Tambah Kategori</button>
          </div>
          <ul className="list">
            {categories.map(c => (
              <li key={c.id}>{c.name}</li>
            ))}
          </ul>
        </>
      )}

      {tab === 'products' && (
        <>
          <div className="action-bar">
            <div className="section-title">Data Produk</div>
            <button className="btn primary" onClick={() => setModal('product')}>Tambah Produk</button>
          </div>
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>SKU</th>
                  <th>Nama Produk</th>
                  <th>Customer</th>
                  <th>Kategori</th>
                </tr>
              </thead>
              <tbody>
                {products.map(p => (
                  <tr key={p.id}>
                    <td>{p.sku}</td>
                    <td>{p.name}</td>
                    <td>{p.customer}</td>
                    <td>{p.category || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      <Modal
        open={modal === 'customer'}
        title="Tambah Customer"
        onClose={closeModal}
        actions={<button onClick={addCustomer} disabled={loading}>Simpan</button>}
      >
        <input placeholder="Nama customer" value={newCustomer} onChange={e => setNewCustomer(e.target.value)} />
      </Modal>

      <Modal
        open={modal === 'category'}
        title="Tambah Kategori"
        onClose={closeModal}
        actions={<button onClick={addCategory} disabled={loading}>Simpan</button>}
      >
        <input placeholder="Nama kategori" value={newCategory} onChange={e => setNewCategory(e.target.value)} />
      </Modal>

      <Modal
        open={modal === 'product'}
        title="Tambah Produk"
        onClose={closeModal}
        actions={<button onClick={addProduct} disabled={loading}>Simpan</button>}
      >
        <div className="panel-row">
          <input placeholder="SKU" value={newProduct.sku} onChange={e => setNewProduct({ ...newProduct, sku: e.target.value })} />
          <input placeholder="Nama produk" value={newProduct.name} onChange={e => setNewProduct({ ...newProduct, name: e.target.value })} />
          <select value={newProduct.customer_id} onChange={e => setNewProduct({ ...newProduct, customer_id: e.target.value })}>
            <option value="">Pilih customer</option>
            {customers.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
          <select value={newProduct.category_id} onChange={e => setNewProduct({ ...newProduct, category_id: e.target.value })}>
            <option value="">Pilih kategori</option>
            {categories.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
      </Modal>

      <ConfirmDialog
        open={confirmDialog.open}
        title="Konfirmasi"
        message={confirmDialog.message}
        confirmText="Simpan"
        confirmClass="primary"
        onConfirm={() => {
          if (confirmDialog.type === 'customer') confirmAddCustomer()
          if (confirmDialog.type === 'category') confirmAddCategory()
          if (confirmDialog.type === 'product') confirmAddProduct()
        }}
        onClose={() => setConfirmDialog({ open: false, type: '', message: '' })}
      />
    </section>
  )
}
