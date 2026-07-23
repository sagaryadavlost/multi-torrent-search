import { useState, useEffect, useCallback, useMemo } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Search, Filter, Loader2, Magnet, ChevronLeft, ChevronRight, Download, X, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
import { useProviders } from '../context/ProvidersContext'
import { api } from '../lib/api'

export default function SearchResults() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const { filteredProviders, categories } = useProviders()
  
  const [query, setQuery] = useState(searchParams.get('q') || '')
  const [category, setCategory] = useState(searchParams.get('category') || 'all')
  const [provider, setProvider] = useState(searchParams.get('provider') || '')
  const [sort, setSort] = useState(searchParams.get('sort') || 'relevance')
  const [order, setOrder] = useState(searchParams.get('order') || 'desc')
  const [page, setPage] = useState(parseInt(searchParams.get('page')) || 1)
  const [results, setResults] = useState([])
  const [totalResults, setTotalResults] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [showFilters, setShowFilters] = useState(false)

  const PER_PAGE = 20
  const totalPages = Math.ceil(totalResults / PER_PAGE)

  const fetchResults = useCallback(async () => {
    if (!query.trim()) {
      setResults([])
      setTotalResults(0)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const params = new URLSearchParams()
      params.set('q', query)
      params.set('page', page.toString())
      params.set('limit', PER_PAGE.toString())
      if (category !== 'all') params.set('category', category)
      if (provider) params.set('provider', provider)
      params.set('sort', sort)
      params.set('order', order)

      const res = await api.get('/api/search', { params })
      setResults(res.data.torrents || [])
      setTotalResults(res.data.total || 0)
    } catch (err) {
      setError(err.message || 'Search failed')
      console.error('Search error:', err)
    } finally {
      setLoading(false)
    }
  }, [query, category, provider, sort, order, page])

  // Update URL when filters change
  useEffect(() => {
    const params = new URLSearchParams()
    if (query) params.set('q', query)
    if (category !== 'all') params.set('category', category)
    if (provider) params.set('provider', provider)
    if (sort !== 'relevance') params.set('sort', sort)
    if (order !== 'desc') params.set('order', order)
    if (page > 1) params.set('page', page.toString())
    setSearchParams(params, { replace: true })
  }, [query, category, provider, sort, order, page, setSearchParams])

  // Fetch on mount and when query/page changes
  useEffect(() => {
    fetchResults()
  }, [fetchResults])

  const handleSearch = (e) => {
    e.preventDefault()
    setPage(1)
  }

  const clearFilters = () => {
    setCategory('all')
    setProvider('')
    setSort('relevance')
    setOrder('desc')
    setPage(1)
  }

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

  const sortOptions = [
    { value: 'relevance', label: 'Relevance' },
    { value: 'seeders', label: 'Seeders' },
    { value: 'leechers', label: 'Leechers' },
    { value: 'size', label: 'Size' },
    { value: 'date', label: 'Date Added' },
  ]

  const getSortLabel = (value) => sortOptions.find(s => s.value === value)?.label || value

  return (
    <div className="space-y-6">
      {/* Search Header */}
      <div className="card p-6">
        <form onSubmit={handleSearch} className="space-y-4" role="search">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-dark-400" aria-hidden="true" />
              <input
                type="search"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search torrents..."
                className="w-full pl-12 pr-4 py-3 input text-lg"
                autoFocus
              />
            </div>
            <button
              type="submit"
              disabled={!query.trim() || loading}
              className="btn-primary px-6 py-3 whitespace-nowrap"
            >
              {loading ? (
                <>
                  <Loader2 className="h-5 w-5 animate-spin" aria-hidden="true" />
                  Searching...
                </>
              ) : (
                <>
                  <Magnet className="h-5 w-5" aria-hidden="true" />
                  Search
                </>
              )}
            </button>
          </div>

          {/* Filter Toggles */}
          <div className="flex flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={() => setShowFilters(!showFilters)}
              className={`btn-secondary flex items-center gap-2 ${showFilters ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300' : ''}`}
            >
              <Funnel className="h-4 w-4" aria-hidden="true" />
              Filters
              {(category !== 'all' || provider) && (
                <span className="badge badge-primary ml-1">
                  {(category !== 'all' ? 1 : 0) + (provider ? 1 : 0)}
                </span>
              )}
            </button>

            {query && (
              <button
                type="button"
                onClick={() => { setQuery(''); setPage(1); }}
                className="btn-ghost text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 flex items-center gap-1"
              >
                <X className="h-4 w-4" aria-hidden="true" />
                Clear Search
              </button>
            )}

            {(category !== 'all' || provider) && (
              <button
                type="button"
                onClick={clearFilters}
                className="btn-ghost flex items-center gap-1"
              >
                <X className="h-4 w-4" aria-hidden="true" />
                Clear Filters
              </button>
            )}
          </div>

          {/* Collapsible Filters */}
          {showFilters && (
            <div className="animate-slide-down border-t border-dark-200 dark:border-dark-800 pt-4 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <div>
                  <label htmlFor="category-filter" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">Category</label>
                  <select
                    id="category-filter"
                    value={category}
                    onChange={(e) => { setCategory(e.target.value); setPage(1); }}
                    className="select w-full"
                  >
                    {categoryOptions.map((cat) => (
                      <option key={cat.value} value={cat.value}>{cat.label}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label htmlFor="provider-filter" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">Provider</label>
                  <select
                    id="provider-filter"
                    value={provider}
                    onChange={(e) => { setProvider(e.target.value); setPage(1); }}
                    className="select w-full"
                  >
                    <option value="">All Providers</option>
                    {filteredProviders.map((p) => (
                      <option key={p.name} value={p.name}>{p.displayName}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label htmlFor="sort-filter" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">Sort By</label>
                  <select
                    id="sort-filter"
                    value={sort}
                    onChange={(e) => { setSort(e.target.value); setPage(1); }}
                    className="select w-full"
                  >
                    {sortOptions.map((opt) => (
                      <option key={opt.value} value={opt.value}>{opt.label}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label htmlFor="order-filter" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">Order</label>
                  <select
                    id="order-filter"
                    value={order}
                    onChange={(e) => { setOrder(e.target.value); setPage(1); }}
                    className="select w-full"
                  >
                    <option value="desc">
                      {sort === 'date' ? 'Newest First' : 'Descending'}
                    </option>
                    <option value="asc">
                      {sort === 'date' ? 'Oldest First' : 'Ascending'}
                    </option>
                  </select>
                </div>
              </div>
            </div>
          )}
        </form>
      </div>

      {/* Results Info */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 text-sm text-dark-600 dark:text-dark-400">
        <div>
          {query ? (
            <>
              <span className="font-medium">{totalResults.toLocaleString()}</span> results for <span className="font-medium text-primary-600 dark:text-primary-400">"{query}"</span>
              {category !== 'all' && (
                <> in <span className="font-medium text-primary-600 dark:text-primary-400">{categoryOptions.find(c => c.value === category)?.label}</span></>
              )}
              {provider && (
                <> from <span className="font-medium text-primary-600 dark:text-primary-400">{filteredProviders.find(p => p.name === provider)?.displayName}</span></>
              )}
            </>
          ) : (
            'Enter a search query to find torrents'
          )}
        </div>
        {totalResults > 0 && (
          <div className="flex items-center gap-2">
            <span>Page {page} of {totalPages}</span>
          </div>
        )}
      </div>

      {/* Results Grid */}
      {loading ? (
        <TorrentGridSkeleton count={8} />
      ) : results.length === 0 ? (
        <EmptyState query={query} />
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4" role="list" aria-label="Search results">
            {results.map((torrent) => (
              <TorrentCard key={torrent.id} torrent={torrent} formatSize={formatSize} formatDate={formatDate} />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <nav className="flex items-center justify-center gap-2" aria-label="Pagination">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="btn-secondary p-2"
                aria-label="Previous page"
              >
                <ChevronLeft className="h-5 w-5" aria-hidden="true" />
              </button>

              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum
                if (totalPages <= 5) {
                  pageNum = i + 1
                } else if (page <= 3) {
                  pageNum = i + 1
                } else if (page >= totalPages - 2) {
                  pageNum = totalPages - 4 + i
                } else {
                  pageNum = page - 2 + i
                }
                return (
                  <button
                    key={pageNum}
                    onClick={() => setPage(pageNum)}
                    className={`w-10 h-10 rounded-lg text-sm font-medium transition-colors ${
                      page === pageNum
                        ? 'bg-primary-600 text-white'
                        : 'text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800'
                    }`}
                    aria-label={`Page ${pageNum}`}
                    aria-current={page === pageNum ? 'page' : undefined}
                  >
                    {pageNum}
                  </button>
                )
              })}

              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="btn-secondary p-2"
                aria-label="Next page"
              >
                <ChevronRight className="h-5 w-5" aria-hidden="true" />
              </button>
            </nav>
          )}
        </>
      )}
    </div>
  )
}

function TorrentCard({ torrent, formatSize, formatDate }) {
  return (
    <article className="card-hover group" role="listitem">
      <Link
        to={`/torrent/${torrent.provider}/${encodeURIComponent(torrent.detailUrl || torrent.magnet || torrent.id)}`}
        className="block"
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
            <span className="flex items-center gap-1" title="Size">
              <Magnet className="h-3.5 w-3.5" aria-hidden="true" />
              {torrent.size && formatSize(torrent.size)}
            </span>
            <span className="flex items-center gap-1" title="Seeders">
              <ArrowUp className="h-3.5 w-3.5 text-green-500" aria-hidden="true" />
              {torrent.seeders?.toLocaleString()}
            </span>
            <span className="flex items-center gap-1" title="Leechers">
              <ArrowDown className="h-3.5 w-3.5 text-red-500" aria-hidden="true" />
              {torrent.leechers?.toLocaleString()}
            </span>
          </div>
          {torrent.uploaded && (
            <div className="mt-2 text-xs text-dark-400 dark:text-dark-500">
              Added {formatDate(torrent.uploaded)}
            </div>
          )}
        </div>
      </Link>
    </article>
  )
}

function TorrentGridSkeleton({ count }) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4" aria-busy="true" aria-label="Loading results">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="card overflow-hidden">
          <div className="aspect-video skeleton" />
          <div className="p-4 space-y-3">
            <div className="skeleton-text w-3/4" />
            <div className="skeleton-text w-1/2" />
            <div className="skeleton-text w-1/4" />
          </div>
        </div>
      ))}
    </div>
  )
}

function EmptyState({ query }) {
  return (
    <div className="card p-12 text-center">
      <Magnet className="mx-auto h-16 w-16 text-dark-300 dark:text-dark-600 mb-4" aria-hidden="true" />
      <h3 className="text-xl font-semibold text-dark-900 dark:text-dark-100 mb-2">
        {query ? 'No results found' : 'Start searching'}
      </h3>
      <p className="text-dark-500 dark:text-dark-400 max-w-md mx-auto">
        {query
          ? `No torrents found for "${query}". Try different keywords or adjust your filters.`
          : 'Enter a search query above to find torrents from multiple providers.'}
      </p>
      {query && (
        <button
          onClick={() => window.location.reload()}
          className="mt-4 btn-primary"
        >
          Clear Search & Try Again
        </button>
      )}
    </div>
  )
}

// Need to import Link
import { Link } from 'react-router-dom'