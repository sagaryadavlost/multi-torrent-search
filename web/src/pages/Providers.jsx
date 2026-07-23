import { useState, useEffect } from 'react'
import { Server, CheckCircle, XCircle, AlertCircle, Loader2, RefreshCw, Eye, EyeOff, Edit2 } from 'lucide-react'
import { useProviders } from '../context/ProvidersContext'
import { api } from '../lib/api'
import { useToast } from '../hooks/useToast'

export default function Providers() {
  const { providers, loading, error, refetch, toggleProvider } = useProviders()
  const { toast } = useToast()
  const [editingProvider, setEditingProvider] = useState(null)
  const [refreshing, setRefreshing] = useState(false)

  const handleToggle = async (provider) => {
    const newEnabled = !provider.enabled
    try {
      await api.post('/api/providers/toggle', { name: provider.name, enabled: newEnabled })
      toggleProvider(provider.name, newEnabled)
      toast.success(`${provider.displayName} ${newEnabled ? 'enabled' : 'disabled'}`)
    } catch (err) {
      toast.error(`Failed to ${newEnabled ? 'enable' : 'disable'} ${provider.displayName}`)
    }
  }

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      await refetch()
      await api.post('/api/providers/refresh')
      toast.success('Providers refreshed')
    } catch (err) {
      toast.error('Failed to refresh providers')
    } finally {
      setRefreshing(false)
    }
  }

  const handleTestProvider = async (provider) => {
    try {
      const res = await api.post('/api/providers/test', { name: provider.name })
      if (res.data.success) {
        toast.success(`${provider.displayName} is working`)
      } else {
        toast.error(`${provider.displayName} test failed: ${res.data.error}`)
      }
    } catch (err) {
      toast.error(`Failed to test ${provider.displayName}`)
    }
  }

  const getStatusBadge = (provider) => {
    if (!provider.enabled) {
      return <span className="badge badge-gray">Disabled</span>
    }
    switch (provider.status) {
      case 'online':
        return <span className="badge badge-success flex items-center gap-1"><CheckCircle className="h-3 w-3" /> Online</span>
      case 'limited':
        return <span className="badge badge-warning flex items-center gap-1"><AlertCircle className="h-3 w-3" /> Limited</span>
      case 'offline':
        return <span className="badge badge-danger flex items-center gap-1"><XCircle className="h-3 w-3" /> Offline</span>
      default:
        return <span className="badge badge-gray">Unknown</span>
    }
  }

  const getCategories = (categories) => {
    if (!categories || categories.length === 0) return 'All'
    return categories.map(c => c.charAt(0).toUpperCase() + c.slice(1)).join(', ')
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Providers</h1>
          <button className="btn-secondary" disabled>
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading...
          </button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <ProviderCardSkeleton key={i} />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Providers</h1>
          <p className="text-dark-500 dark:text-dark-400 mt-1">
            Manage torrent search providers and their status
          </p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className="btn-secondary flex items-center gap-2"
        >
          <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          icon={Server}
          value={providers.filter(p => p.enabled).length}
          label="Enabled"
          color="primary"
        />
        <StatCard
          icon={CheckCircle}
          value={providers.filter(p => p.enabled && p.status === 'online').length}
          label="Online"
          color="green"
        />
        <StatCard
          icon={AlertCircle}
          value={providers.filter(p => p.enabled && p.status === 'limited').length}
          label="Limited"
          color="yellow"
        />
        <StatCard
          icon={XCircle}
          value={providers.filter(p => p.enabled && p.status === 'offline').length}
          label="Offline"
          color="red"
        />
      </div>

      {/* Providers Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {providers.map((provider) => (
          <ProviderCard
            key={provider.name}
            provider={provider}
            onToggle={handleToggle}
            onTest={handleTestProvider}
            onEdit={() => setEditingProvider(provider)}
            getStatusBadge={getStatusBadge}
            getCategories={getCategories}
          />
        ))}
      </div>

      {/* Edit Modal */}
      {editingProvider && (
        <ProviderEditModal
          provider={editingProvider}
          onClose={() => setEditingProvider(null)}
          onSave={handleRefresh}
        />
      )}
    </div>
  )
}

function ProviderCard({ provider, onToggle, onTest, onEdit, getStatusBadge, getCategories }) {
  return (
    <div className="card-hover p-6">
      <div className="flex items-start justify-between gap-4 mb-4">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 bg-dark-100 dark:bg-dark-800 rounded-xl flex items-center justify-center">
            <Server className="h-6 w-6 text-primary-600 dark:text-primary-400" />
          </div>
          <div>
            <h3 className="font-semibold text-dark-900 dark:text-dark-100">{provider.displayName}</h3>
            <p className="text-sm text-dark-500 dark:text-dark-400">{provider.name}</p>
          </div>
        </div>
        {getStatusBadge(provider)}
      </div>

      <div className="space-y-3 mb-4">
        <div className="flex items-center gap-2 text-sm text-dark-500 dark:text-dark-400">
          <span className="font-medium text-dark-900 dark:text-dark-100">Categories:</span>
          <span>{getCategories(provider.categories)}</span>
        </div>
        {provider.description && (
          <p className="text-sm text-dark-600 dark:text-dark-400 line-clamp-2">{provider.description}</p>
        )}
        {provider.url && (
          <a
            href={provider.url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-primary-600 dark:text-primary-400 hover:underline flex items-center gap-1"
          >
            <Eye className="h-3.5 w-3.5" />
            Visit Site
          </a>
        )}
      </div>

      <div className="flex flex-wrap gap-2 pt-4 border-t border-dark-200 dark:border-dark-800">
        <button
          onClick={() => onToggle(provider)}
          className={`btn-sm flex-1 min-w-[100px] ${provider.enabled ? 'btn-secondary' : 'btn-primary'}`}
        >
          {provider.enabled ? 'Disable' : 'Enable'}
        </button>
        <button
          onClick={() => onTest(provider)}
          className="btn-ghost text-sm flex-1 min-w-[100px]"
          disabled={!provider.enabled}
        >
          Test
        </button>
        <button
          onClick={() => onEdit(provider)}
          className="btn-ghost text-sm p-2"
          title="Edit"
        >
          <Edit2 className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}

function ProviderEditModal({ provider, onClose, onSave }) {
  const [formData, setFormData] = useState({
    displayName: provider.displayName,
    description: provider.description || '',
    categories: provider.categories?.join(', ') || '',
    enabled: provider.enabled,
    priority: provider.priority || 0,
  })
  const [saving, setSaving] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.put(`/api/providers/${provider.name}`, {
        ...formData,
        categories: formData.categories.split(',').map(c => c.trim()).filter(Boolean),
      })
      onSave()
      onClose()
    } catch (err) {
      console.error('Failed to save provider:', err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="bg-white dark:bg-dark-900 rounded-xl shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6 border-b border-dark-200 dark:border-dark-800 flex items-center justify-between">
          <h2 className="text-xl font-semibold">Edit Provider</h2>
          <button onClick={onClose} className="p-1 text-dark-400 hover:text-dark-600 dark:hover:text-dark-300">
            <XCircle className="h-5 w-5" />
          </button>
        </div>
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Display Name</label>
            <input
              type="text"
              value={formData.displayName}
              onChange={(e) => setFormData({ ...formData, displayName: e.target.value })}
              className="input"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="input"
              rows={3}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Categories (comma separated)</label>
            <input
              type="text"
              value={formData.categories}
              onChange={(e) => setFormData({ ...formData, categories: e.target.value })}
              className="input"
              placeholder="movies, tv, music, games"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Priority</label>
            <input
              type="number"
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 0 })}
              className="input"
              min="0"
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="enabled"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              className="w-4 h-4 text-primary-600 border-dark-300 rounded focus:ring-primary-500"
            />
            <label htmlFor="enabled" className="text-sm">Enabled</label>
          </div>
        </div>
        <div className="p-6 border-t border-dark-200 dark:border-dark-800 flex justify-end gap-3">
          <button onClick={onClose} className="btn-secondary" disabled={saving}>
            Cancel
          </button>
          <button onClick={handleSave} className="btn-primary" disabled={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>
    </div>
  )
}

function StatCard({ icon: Icon, value, label, color }) {
  const colors = {
    primary: 'bg-primary-500/10 text-primary-600 dark:text-primary-400',
    green: 'bg-green-500/10 text-green-600 dark:text-green-400',
    yellow: 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400',
    red: 'bg-red-500/10 text-red-600 dark:text-red-400',
  }

  return (
    <div className={`card p-6 ${colors[color]} dark:bg-opacity-20`}>
      <div className="flex items-center justify-between mb-2">
        <Icon className="h-6 w-6" />
      </div>
      <div className="text-3xl font-bold">{value}</div>
      <div className="text-sm opacity-75">{label}</div>
    </div>
  )
}

function ProviderCardSkeleton() {
  return (
    <div className="card p-6 animate-pulse">
      <div className="flex items-start justify-between gap-4 mb-4">
        <div className="flex items-center gap-3">
          <div className="skeleton w-12 h-12 rounded-xl" />
          <div>
            <div className="skeleton-text w-32" />
            <div className="skeleton-text w-24 mt-1" />
          </div>
        </div>
        <div className="skeleton badge" style={{ width: '80px' }} />
      </div>
      <div className="space-y-3 mb-4">
        <div className="skeleton-text w-3/4" />
        <div className="skeleton-text w-1/2" />
      </div>
      <div className="flex gap-2 pt-4 border-t border-dark-200 dark:border-dark-800">
        <div className="skeleton btn-secondary flex-1 min-w-[100px]" style={{ height: '36px' }} />
        <div className="skeleton btn-ghost flex-1 min-w-[100px]" style={{ height: '36px' }} />
      </div>
    </div>
  )
}
