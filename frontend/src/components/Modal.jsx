export default function Modal({ open, title, onClose, children, actions }) {
  if (!open) return null
  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-title">{title}</div>
        <div className="modal-body">{children}</div>
        <div className="modal-actions">
          <button className="ghost" onClick={onClose}>Batal</button>
          {actions}
        </div>
      </div>
    </div>
  )
}
