import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react'
import { api } from '../lib/api'

const ProvidersContext = createContext(null)

export function ProvidersProvider({ children }) {
  const [providers, setProviders] = useState([])
  const [categories, setCategories] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [enabledOnly, setEnabledOnly] = useState(true)

  const fetchProviders = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const [providersRes, categoriesRes] = await Promise.all([
        api.get('/api/providers'),
        api.get('/api/categories'),
      ])
      setProviders(providersRes.data.providers || [])
      setCategories(categoriesRes.data.categories || [])
    } catch (err) {
      setError(err.message)
      console.error('Failed to fetch providers:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchProviders()
  }, [fetchProviders])

  const filteredProviders = useMemo(() => {
    let filtered = providers
    if (enabledOnly) {
      filtered = filtered.filter(p => p.enabled)
    }
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(p => p.categories.includes(selectedCategory))
    }
    return filtered
  }, [providers, selectedCategory, enabledOnly])

  const toggleProvider = useCallback(async (providerName, enabled) => {
    // This would call an API to enable/disable provider
    setProviders(prev => prev.map(p => 
      p.name === providerName ? { ...p, enabled } : p
    ))
  }, [])

  const value = {
    providers,
    filteredProviders,
    categories,
    loading,
    error,
    selectedCategory,
    setSelectedCategory,
    enabledOnly,
    setEnabledOnly,
    toggleProvider,
    refetch: fetchProviders,
  }

  return (
    <ProvidersContext.Provider value={value}>
      {children}
    </ProvidersContext.Provider>
  )
}

export function useProviders() {
  const context = useContext(ProvidersContext)
  if (!context) {
    throw new Error('useProviders must be used within a ProvidersProvider')
  }
  return context
}