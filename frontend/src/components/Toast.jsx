import { createContext, useCallback, useContext, useMemo, useState } from 'react'

const ToastContext = createContext(null)

const genId = () => `${Date.now()}-${Math.random().toString(16).slice(2)}`

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])

  const remove = useCallback((id) => {
    setToasts((current) => current.filter(t => t.id !== id))
  }, [])

  const notify = useCallback((type, text, options = {}) => {
    const id = genId()
    const ttl = options.ttl ?? (type === 'error' ? 6000 : 3500)
    setToasts((current) => [...current, { id, type, text }])
    if (ttl > 0) {
      setTimeout(() => remove(id), ttl)
    }
  }, [remove])

  const value = useMemo(() => ({ notify }), [notify])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="toast-container" aria-live="polite" aria-atomic="true">
        {toasts.map(t => (
          <div key={t.id} className={`toast ${t.type}`}>
            <div className="toast-text">{t.text}</div>
            <button className="toast-close" onClick={() => remove(t.id)} aria-label="Tutup notifikasi">×</button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used inside ToastProvider')
  return ctx
}
