import Modal from './Modal.jsx'

export default function ConfirmDialog({
  open,
  title,
  message,
  confirmText = 'Ya',
  onConfirm,
  onClose,
  confirmClass = 'primary',
  showNote = false,
  noteText = 'Tindakan ini tidak bisa dibatalkan.'
}) {
  return (
    <Modal
      open={open}
      title={title}
      onClose={onClose}
      actions={(
        <button className={`btn ${confirmClass}`} onClick={onConfirm}>
          {confirmText}
        </button>
      )}
    >
      <div className="muted">{message}</div>
      {showNote && <div className="hint">{noteText}</div>}
    </Modal>
  )
}
