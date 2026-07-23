import { Link, useSearchParams } from 'react-router-dom'
import { useState, useEffect, useCallback } from 'react'
import { Search, Filter, Loader2, Magnet, TrendingUp, Star, Download, Clock, AlertCircle } from 'lucide-react'
import { useProviders } from '../context/ProvidersContext'
import { api } from '../lib/api'

export default function Home() {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [trending, setTrending] = useState([])
  const [recent, setRecent] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [searchParams] = useSearchParams()
  const { filteredProviders, categories } = useProviders()

  // Initialize search from URL params
  useEffect(() => {
    const q = searchParams.get('q')
    const cat = searchParams.get('category')
    if (q) setSearchQuery(q)
    if (cat) setSelectedCategory(cat)
  }, [searchParams])

  const handleSearch = useCallback(async (e) => {
    e.preventDefault()
    if (!searchQuery.trim()) return
    
    const params = new URLSearchParams()
    params.set('q', searchQuery)
    if (selectedCategory !== 'all') params.set('category', selectedCategory)
    window.location.href = `/search?${params.toString()}`
  }, [searchQuery, selectedCategory])

  const fetchTrending = useCallback(async () => {
    try {
      const res = await api.get('/api/trending', { params: { limit: 10 } })
      setTrending(res.data.torrents || [])
    } catch (err) {
      console.error('Failed to fetch trending:', err)
    }
  }, [])

  const fetchRecent = useCallback(async () => {
    try {
      const res = await api.get('/api/recent', { params: { limit: 10 } })
      setRecent(res.data.torrents || [])
    } catch (err) {
      console.error('Failed to fetch recent:', err)
    }
  }, [])

  useEffect(() => {
    fetchTrending()
    fetchRecent()
  }, [fetchTrending, fetchRecent])

  const formatSize = (bytes) => {
    if (!bytes) return 'Unknown'
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let size = bytes
    let unitIndex = 0
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    return `${size.toFixed(1)} ${units[unitIndex]}`
  }

  const formatDate = (dateStr) => {
    if (!dateStr) return 'Unknown'
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  const categoryOptions = [
    { value: 'all', label: 'All Categories' },
    { value: 'movies', label: 'Movies' },
    { value: 'tv', label: 'TV Shows' },
    { value: 'music', label: 'Music' },
    { value: 'games', label: 'Games' },
    { value: 'software', label: 'Software' },
    { value: 'anime', label: 'Anime' },
    { value: 'books', label: 'Books' },
    { value: 'xxx', label: 'XXX' },
    { value: 'other', label: 'Other' },
  ]

  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <section className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary-600 via-primary-700 to-primary-900 p-8 md:p-12 text-white">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-primary-500/20 via-transparent to-transparent" />
        <div className="relative max-w-3xl">
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight mb-4">
            Find Any Torrent <span className="text-yellow-300">Instantly</span>
          </h1>
          <p className="text-lg md:text-xl text-primary-100 mb-8 max-w-2xl">
            Search across multiple torrent providers simultaneously. Fast, private, and always up-to-date.
          </p>
          
          {/* Search Form */}
          <form onSubmit={handleSearch} className="relative" role="search">
            <div className="flex flex-col sm:flex-row gap-3">
              <div className="relative flex-1">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-primary-300" aria-hidden="true" />
                <input
                  type="search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search for movies, TV shows, music, games, software..."
                  className="w-full pl-12 pr-4 py-4 bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl text-white placeholder-primary-200 focus:outline-none focus:ring-2 focus:ring-yellow-400 focus:border-transparent text-lg"
                  autoFocus
                  autoComplete="off"
                />
              </div>
              <button
                type="submit"
                disabled={!searchQuery.trim() || loading}
                className="px-8 py-4 bg-yellow-400 hover:bg-yellow-300 text-primary-900 font-bold rounded-xl transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
              >
                <Magnet className="h-5 w-5" aria-hidden="true" />
                Search
              </button>
            </div>
            
            {/* Category Filter */}
            <div className="mt-4 flex flex-wrap items-center gap-3">
              <Filter className="h-5 w-5 text-primary-300" aria-hidden="true" />
              <select
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
                className="px-4 py-2 bg-white/10 backdrop-blur-sm border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-yellow-400 focus:border-transparent cursor-pointer"
              >
                {categoryOptions.map((cat) => (
                  <option key={cat.value} value={cat.value}>{cat.label}</option>
                ))}
              </select>
              <span className="text-sm text-primary-200 hidden sm:inline">
                {filteredProviders.length} providers available
              </span>
            </div>
          </form>
        </div>
        
        {/* Decorative elements */}
        <div className="absolute top-0 right-0 w-1/2 h-full opacity-10">
          <Magnet className="h-full w-full" aria-hidden="true" />
        </div>
      </section>

      {/* Stats */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-4" aria-label="Statistics">
        <StatCard
          icon={Magnet}
          value={filteredProviders.length}
          label="Active Providers"
          color="primary"
        />
        <StatCard
          icon={TrendingUp}
          value={trending.length}
          label="Trending Now"
          color="yellow"
        />
        <StatCard
          icon={Clock}
          value={recent.length}
          label="Recent Added"
          color="green"
        />
        <StatCard
          icon={Star}
          value="99.9%"
          label="Uptime"
          color="purple"
        />
      </section>

      {/* Trending */}
      {trending.length > 0 && (
        <section aria-labelledby="trending-heading">
          <div className="flex items-center justify-between mb-6">
            <h2 id="trending-heading" className="text-2xl font-bold flex items-center gap-2">
              <TrendingUp className="h-6 w-6 text-yellow-500" aria-hidden="true" />
              Trending Now
            </h2>
            <Link
              to="/search"
              query={{ sort: 'seeders', order: 'desc' }}
              className="text-primary-600 dark:text-primary-400 hover:underline text-sm font-medium flex items-center gap-1"
            >
              View All
              <Search className="h-4 w-4" aria-hidden="true" />
            </Link>
          </div>
          <TorrentGrid torrents={trending} />
        </section>
      )}

      {/* Recent */}
      {recent.length > 0 && (
        <section aria-labelledby="recent-heading">
          <div className="flex items-center justify-between mb-6">
            <h2 id="recent-heading" className="text-2xl font-bold flex items-center gap-2">
              <Clock className="h-6 w-6 text-green-500" aria-hidden="true" />
              Recently Added
            </h2>
            <Link
              to="/search"
              query={{ sort: 'date', order: 'desc' }}
              className="text-primary-600 dark:text-primary-400 hover:underline text-sm font-medium flex items-center gap-1"
            >
              View All
              <Search className="h-4 w-4" aria-hidden="true" />
            </Link>
          </div>
          <TorrentGrid torrents={recent} />
        </section>
      )}

      {/* Features */}
      <section className="pt-8" aria-labelledby="features-heading">
        <h2 id="features-heading" className="text-2xl font-bold text-center mb-12">Why TorrentSearch?</h2>
        <div className="grid md:grid-cols-3 gap-6">
          <FeatureCard
            icon={Magnet}
            title="Multi-Source Search"
            description="Search across 10+ torrent providers simultaneously for the best results"
          />
          <FeatureCard
            icon={Download}
            title="Direct Downloads"
            description="Get magnet links and torrent files instantly without redirects"
          />
          <FeatureCard
            icon={AlertCircle}
            title="Safe & Private"
            description="No tracking, no ads, no login required. Your searches stay private"
          />
        </div>
      </section>

      {/* Providers List */}
      <section aria-labelledby="providers-heading">
        <h2 id="providers-heading" className="text-2xl font-bold mb-6">Available Providers</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {filteredProviders.map((provider) => (
            <ProviderCard key={provider.name} provider={provider} />
          ))}
        </div>
      </section>
    </div>
  )
}

function StatCard({ icon: Icon, value, label, color }) {
  const colors = {
    primary: 'bg-primary-500/10 text-primary-600 dark:text-primary-400',
    yellow: 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400',
    green: 'bg-green-500/10 text-green-600 dark:text-green-400',
    purple: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
  }

  return (
    <div className={`rounded-xl p-6 ${colors[color]} dark:bg-opacity-20`}>
      <div className="flex items-center justify-between mb-2">
        <Icon className="h-6 w-6" aria-hidden="true" />
      </div>
      <div className="text-3xl font-bold">{value}</div>
      <div className="text-sm opacity-75">{label}</div>
    </div>
  )
}

function FeatureCard({ icon: Icon, title, description }) {
  return (
    <div className="card p-6 text-center hover:shadow-lg transition-shadow">
      <div className="mx-auto mb-4 w-12 h-12 bg-primary-100 dark:bg-primary-900/30 rounded-xl flex items-center justify-center">
        <Icon className="h-6 w-6 text-primary-600 dark:text-primary-400" aria-hidden="true" />
      </div>
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      <p className="text-dark-600 dark:text-dark-400">{description}</p>
    </div>
  )
}

function ProviderCard({ provider }) {
  const statusColors = {
    online: 'bg-green-500',
    offline: 'bg-red-500',
    limited: 'bg-yellow-500',
  }

  return (
    <Link
      to="/providers"
      className="card-hover p-4 text-center group"
      aria-label={`View ${provider.displayName} provider`}
    >
      <div className="relative mb-3">
        <div className="mx-auto w-16 h-16 bg-dark-100 dark:bg-dark-800 rounded-xl flex items-center justify-center group-hover:scale-105 transition-transform">
          <Magnet className="h-8 w-8 text-primary-600 dark:text-primary-400" aria-hidden="true" />
        </div>
        <span className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white dark:border-dark-900 ${statusColors[provider.status] || statusColors.offline}`} />
      </div>
      <h3 className="font-medium text-sm truncate">{provider.displayName}</h3>
      <p className="text-xs text-dark-500 dark:text-dark-400 truncate">{provider.description}</p>
    </Link>
  )
}

function TorrentGrid({ torrents }) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      {torrents.map((torrent) => (
        <TorrentCard key={torrent.id} torrent={torrent} />
      ))}
    </div>
  )
}

function TorrentCard({ torrent }) {
  return (
    <Link
      to={`/torrent/${torrent.provider}/${encodeURIComponent(torrent.detailUrl || torrent.magnet || torrent.id)}`}
      className="card-hover group"
      Name="card-hover group block"
    >
      <div className="aspect-video bg-dark-100 dark:bg-dark-800 relative overflow-hidden">
        {torrent.thumbnail ? (
          <img
            src={torrent.thumbnail}
            alt=""
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <Magnet className="h-12 w-12 text-dark-300 dark:text-dark-600" aria-hidden="true" />
          </div>
        )}
        <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
        <div className="absolute bottom-0 left-0 right-0 p-3 transform translate-y-full group-hover:translate-y-0 transition-transform">
          <div className="flex items-center justify-between gap-2">
            <span className="badge badge-primary">{torrent.category}</span>
            <span className="badge badge-success">{torrent.provider}</span>
          </div>
        </div>
      </div>
      <div className="p-4">
        <h3 className="font-medium text-dark-900 dark:text-dark-100 line-clamp-2 mb-2 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
          {torrent.name}
        </h3>
        <div className="flex flex-wrap items-center gap-3 text-sm text-dark-500 dark:text-dark-400">
          <span className="flex items-center gap-1">
            <Magnet className="h-3.5 w-3.5" aria-hidden="true" />
            {torrent.size && formatSize(torrent.size)}
          </span>
          <span className="flex items-center gap-1">
            <Download className="h-3.5 w-3.5" aria-hidden="true" />
            {torrent.seeders?.toLocaleString()} S
          </span>
          <span className="flex items-center gap-1">
            <Download className="h-3.5 w-3.5" aria-hidden="true" />
            {torrent.leechers?.toLocaleString()} L
          </span>
        </div>
        {torrent.uploaded && (
          <div className="mt-2 text-xs text-dark-400 dark:text-dark-500">
            Added {formatDate(torrent.uploaded)}
          </div>
        )}
      </div>
    </Link>
  )
}