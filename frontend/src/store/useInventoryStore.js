import { create } from 'zustand'
import { apiGet, apiPost } from '../api/client.js'

export const useInventoryStore = create((set, get) => ({
  inventory: [],
  loading: false,
  error: null,
  async fetchInventory(filter = {}) {
    set({ loading: true, error: null })
    const qs = new URLSearchParams(filter).toString()
    try {
      const data = await apiGet(`/inventory?${qs}`)
      set({ inventory: data, loading: false })
    } catch (e) {
      set({ error: e.message, loading: false })
    }
  },
  async adjustStock(payload) {
    await apiPost('/inventory/adjust', payload)
    await get().fetchInventory({})
  }
}))
