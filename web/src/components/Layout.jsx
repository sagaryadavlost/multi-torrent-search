import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { Menu, X, Sun, Moon, Search, Settings, Server, Magnet, Info, ChevronDown } from 'lucide-react'
import { useState, useEffect } from 'react'
import { useTheme } from '../context/ThemeContext'
import { useProviders } from '../context/ProvidersContext'

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { theme, toggleTheme, setTheme, isDark } = useTheme()
  const { providers, filteredProviders, loading: providersLoading } = useProviders()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [selectedProvider, setSelectedProvider] = useState('')

  const handleSearch = (e) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      const params = new URLSearchParams()
      params.set('q', searchQuery)
      if (selectedCategory !== 'all') params.set('category', selectedCategory)
      if (selectedProvider) params.set('provider', selectedProvider)
      navigate(`/search?${params.toString()}`)
    }
  }

  const navItems = [
    { path: '/', label: 'Home', icon: Home },
    { path: '/providers', label: 'Providers', icon: Server },
    { path: '/settings', label: 'Settings', icon: Settings },
  ]

  return (
    <div className="min-h-screen bg-dark-50 dark:bg-dark-950 text-dark-900 dark:text-dark-50 transition-colors duration-200">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white/80 dark:bg-dark-900/80 backdrop-blur-md border-b border-dark-200 dark:border-dark-800">
        <nav className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8" aria-label="Main navigation">
          <div className="flex h-16 items-center justify-between">
            {/* Logo */}
            <Link to="/" className="flex items-center gap-2 text-primary-600 dark:text-primary-400 hover:opacity-80 transition-opacity" aria-label="TorrentSearch Home">
              <Magnet className="h-7 w-7" aria-hidden="true" />
              <span className="font-bold text-xl tracking-tight">TorrentSearch</span>
            </Link>

            {/* Desktop Navigation */}
            <div className="hidden md:flex md:items-center md:gap-6">
              <form onSubmit={handleSearch} className="relative flex-1 max-w-xl" role="search">
                <label htmlFor="search-input" className="sr-only">Search torrents</label>
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-dark-400 dark:text-dark-500" aria-hidden="true" />
                <input
                  id="search-input"
                  type="search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search torrents..."
                  className="w-full pl-10 pr-4 py-2 bg-dark-100 dark:bg-dark-800 border border-dark-200 dark:border-dark-700 rounded-lg text-dark-900 dark:text-dark-100 placeholder-dark-400 dark:placeholder-dark-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all"
                  autoComplete="off"
                />
              </form>

              <div className="flex items-center gap-2">
                {navItems.map((item) => (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={`flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                      location.pathname === item.path
                        ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                        : 'text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800 hover:text-dark-900 dark:hover:text-dark-100'
                    }`}
                  >
                    <item.icon className="h-5 w-5" aria-hidden="true" />
                    {item.label}
                  </Link>
                ))}

                {/* Theme Toggle */}
                <button
                  onClick={toggleTheme}
                  className="p-2 rounded-lg text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800 transition-colors"
                  aria-label={`Switch to ${isDark ? 'light' : 'dark'} mode`}
                >
                  {isDark ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
                </button>

                {/* Mobile Menu Button */}
                <button
                  className="md:hidden p-2 rounded-lg text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800 transition-colors"
                  onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                  aria-expanded={mobileMenuOpen}
                  aria-controls="mobile-menu"
                  aria-label="Toggle menu"
                >
                  {mobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
                </button>
              </div>
            </div>
          </div>
        </nav>

        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div id="mobile-menu" className="md:hidden py-4 border-t border-dark-200 dark:border-dark-800 animate-slide-down">
            <form onSubmit={handleSearch} className="mb-4" role="search">
              <label htmlFor="mobile-search" className="sr-only">Search torrents</label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-dark-400" aria-hidden="true" />
                <input
                  id="mobile-search"
                  type="search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search torrents..."
                  className="w-full pl-10 pr-4 py-2 bg-dark-100 dark:bg-dark-800 border border-dark-200 dark:border-dark-700 rounded-lg text-dark-900 dark:text-dark-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
                />
              </div>
            </form>

            <div className="space-y-1">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setMobileMenuOpen(false)}
                  className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-base font-medium transition-colors ${
                    location.pathname === item.path
                      ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                      : 'text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800'
                  }`}
                >
                  <item.icon className="h-5 w-5" aria-hidden="true" />
                  {item.label}
                </Link>
              ))}

              <div className="pt-2 border-t border-dark-200 dark:border-dark-800">
                <button
                  onClick={toggleTheme}
                  className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-base font-medium text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800 transition-colors"
                >
                  {isDark ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
                  {isDark ? 'Light Mode' : 'Dark Mode'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Search Filters Bar (Desktop) */}
        <div className="hidden md:block px-4 py-3 bg-dark-50 dark:bg-dark-900 border-b border-dark-200 dark:border-dark-800">
          <div className="max-w-7xl mx-auto flex flex-wrap items-center gap-3">
            <label htmlFor="category-filter" className="text-sm font-medium text-dark-600 dark:text-dark-400">Category:</label>
            <select
              id="category-filter"
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="px-3 py-1.5 bg-white dark:bg-dark-800 border border-dark-200 dark:border-dark-700 rounded-lg text-sm text-dark-900 dark:text-dark-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="all">All Categories</option>
              <option value="movies">Movies</option>
              <option value="tv">TV Shows</option>
              <option value="music">Music</option>
              <option value="games">Games</option>
              <option value="software">Software</option>
              <option value="anime">Anime</option>
              <option value="books">Books</option>
              <option value="xxx">XXX</option>
              <option value="other">Other</option>
            </select>

            <label htmlFor="provider-filter" className="text-sm font-medium text-dark-600 dark:text-dark-400">Provider:</label>
            <select
              id="provider-filter"
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value)}
              className="px-3 py-1.5 bg-white dark:bg-dark-800 border border-dark-200 dark:border-dark-700 rounded-lg text-sm text-dark-900 dark:text-dark-100 focus:outline-none focus:ring-2 focus:ring-primary-500 min-w-[180px]"
            >
              <option value="">All Providers</option>
              {filteredProviders.map((p) => (
                <option key={p.name} value={p.name}>
                  {p.displayName} ({p.name})
                </option>
              ))}
            </select>

            <div className="flex-1" />

            <div className="text-sm text-dark-500 dark:text-dark-400">
              {providersLoading ? 'Loading providers...' : `${filteredProviders.length} providers available`}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6" id="main-content">
        <Outlet />
      </main>

      {/* Footer */}
      <footer className="bg-white dark:bg-dark-950 border-t border-dark-200 dark:border-dark-800 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-sm text-dark-500 dark:text-dark-400">
              <Magnet className="h-4 w-4 text-primary-500" aria-hidden="true" />
              <span>TorrentSearch v1.0.0</span>
            </div>
            <div className="flex items-center gap-6 text-sm text-dark-500 dark:text-dark-400">
              <a href="https://github.com/prajwalch/TorrentSearch" target="_blank" rel="noopener noreferrer" className="hover:text-primary-500 transition-colors flex items-center gap-1">
                <Info className="h-4 w-4" aria-hidden="true" />
                About
              </a>
              <a href="https://github.com/prajwalch/TorrentSearch/issues" target="_blank" rel="noopener noreferrer" className="hover:text-primary-500 transition-colors flex items-center gap-1">
                <Info className="h-4 w-4" aria-hidden="true" />
                Report Issue
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

// Home icon component
function Home({ className, ...props }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  )
}