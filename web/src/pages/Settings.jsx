import { useState, useEffect } from 'react'
import { 
  Sun, Moon, Monitor, Palette, Bell, Shield, 
  Download, FolderOpen, Trash2, Save, Loader2,
  ChevronRight, ChevronLeft
} from 'lucide-react'
import { useTheme } from '../context/ThemeContext'
import { useToast } from '../hooks/useToast'

export default function Settings() {
  const { theme, setTheme, isDark } = useTheme()
  const { toast } = useToast()
  const [activeTab, setActiveTab] = useState('general')
  const [saving, setSaving] = useState(false)
  const [settings, setSettings] = useState({
    theme: 'system',
    language: 'en',
    resultsPerPage: 20,
    defaultCategory: 'all',
    defaultSort: 'relevance',
    defaultOrder: 'desc',
    showThumbnails: true,
    compactMode: false,
    autoFocusSearch: true,
    saveHistory: true,
    maxHistoryItems: 100,
    enableNotifications: false,
    notificationSound: true,
    safeSearch: false,
    adultContent: false,
    proxyEnabled: false,
    proxyUrl: '',
    downloadPath: '',
    autoOpenMagnet: false,
    confirmBeforeDownload: true,
  })

  useEffect(() => {
    const saved = localStorage.getItem('app-settings')
    if (saved) {
      try {
        setSettings(JSON.parse(saved))
        // Apply theme
        const themeSetting = JSON.parse(saved).theme
        if (themeSetting) setTheme(themeSetting)
      } catch (e) {
        console.error('Failed to parse settings:', e)
      }
    }
    // Set initial theme from context
    setSettings(prev => ({ ...prev, theme: isDark ? 'dark' : 'light' }))
  }, [isDark, setTheme])

  const handleChange = (key, value) => {
    setSettings(prev => ({ ...prev, [key]: value }))
    if (key === 'theme') {
      setTheme(value)
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      localStorage.setItem('app-settings', JSON.stringify(settings))
      toast.success('Settings saved successfully')
    } catch (err) {
      toast.error('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  const handleReset = () => {
    const defaults = {
      theme: 'system',
      language: 'en',
      resultsPerPage: 20,
      defaultCategory: 'all',
      defaultSort: 'relevance',
      defaultOrder: 'desc',
      showThumbnails: true,
      compactMode: false,
      autoFocusSearch: true,
      saveHistory: true,
      maxHistoryItems: 100,
      enableNotifications: false,
      notificationSound: true,
      safeSearch: false,
      adultContent: false,
      proxyEnabled: false,
      proxyUrl: '',
      downloadPath: '',
      autoOpenMagnet: false,
      confirmBeforeDownload: true,
    }
    setSettings(defaults)
    setTheme('system')
    localStorage.removeItem('app-settings')
    toast.success('Settings reset to defaults')
  }

  const tabs = [
    { id: 'general', label: 'General', icon: Sun },
    { id: 'appearance', label: 'Appearance', icon: Palette },
    { id: 'search', label: 'Search', icon: Shield },
    { id: 'downloads', label: 'Downloads', icon: Download },
    { id: 'privacy', label: 'Privacy', icon: Shield },
    { id: 'advanced', label: 'Advanced', icon: ChevronRight },
  ]

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <header>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-dark-500 dark:text-dark-400 mt-1">
          Customize your TorrentSearch experience
        </p>
      </header>

      <div className="card overflow-hidden">
        {/* Tab Navigation */}
        <div className="border-b border-dark-200 dark:border-dark-800 overflow-x-auto">
          <nav className="flex gap-1 px-2" role="tablist" aria-label="Settings categories">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                role="tab"
                aria-selected={activeTab === tab.id}
                aria-controls={`${tab.id}-panel`}
                id={`${tab.id}-tab`}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-4 py-3 text-sm font-medium rounded-t-lg transition-colors whitespace-nowrap ${
                  activeTab === tab.id
                    ? 'bg-white dark:bg-dark-900 text-primary-600 dark:text-primary-400 border-b-2 border-primary-600 dark:border-primary-400 -mb-px'
                    : 'text-dark-500 dark:text-dark-400 hover:text-dark-700 dark:hover:text-dark-300 hover:bg-dark-50 dark:hover:bg-dark-800/50'
                }`}
              >
                <tab.icon className="h-4 w-4" aria-hidden="true" />
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Panels */}
        <div className="p-6">
          {activeTab === 'general' && <GeneralSettings settings={settings} onChange={handleChange} />}
          {activeTab === 'appearance' && <AppearanceSettings settings={settings} onChange={handleChange} />}
          {activeTab === 'search' && <SearchSettings settings={settings} onChange={handleChange} />}
          {activeTab === 'downloads' && <DownloadSettings settings={settings} onChange={handleChange} />}
          {activeTab === 'privacy' && <PrivacySettings settings={settings} onChange={handleChange} />}
          {activeTab === 'advanced' && <AdvancedSettings settings={settings} onChange={handleChange} />}
        </div>

        {/* Save/Reset Bar */}
        <div className="px-6 py-4 border-t border-dark-200 dark:border-dark-800 flex items-center justify-end gap-3 bg-dark-50 dark:bg-dark-900/50">
          <button onClick={handleReset} className="btn-ghost" disabled={saving}>
            Reset to Defaults
          </button>
          <button onClick={handleSave} className="btn-primary" disabled={saving}>
            {saving ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                Save Changes
              </>
            )}
          </button>
        </div>
      </div>

      {/* Info Card */}
      <div className="card p-6 bg-primary-50 dark:bg-primary-900/20 border-primary-200 dark:border-primary-800">
        <div className="flex items-start gap-4">
          <div className="w-10 h-10 bg-primary-100 dark:bg-primary-800 rounded-lg flex items-center justify-center flex-shrink-0">
            <Shield className="h-5 w-5 text-primary-600 dark:text-primary-400" />
          </div>
          <div>
            <h3 className="font-semibold text-primary-900 dark:text-primary-100">Privacy First</h3>
            <p className="text-sm text-primary-700 dark:text-primary-300 mt-1">
              TorrentSearch doesn't track your searches, store personal data, or require accounts. 
              All settings are stored locally in your browser.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

function GeneralSettings({ settings, onChange }) {
  return (
    <div className="space-y-6" role="tabpanel" id="general-panel" aria-labelledby="general-tab">
      <section>
        <h3 className="text-lg font-semibold mb-4">Theme</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {[
            { value: 'light', label: 'Light', icon: Sun, desc: 'Always use light mode' },
            { value: 'dark', label: 'Dark', icon: Moon, desc: 'Always use dark mode' },
            { value: 'system', label: 'System', icon: Monitor, desc: 'Follow system preference' },
          ].map((option) => (
            <label
              key={option.value}
              className={`relative cursor-pointer p-4 rounded-xl border-2 transition-all ${
                settings.theme === option.value
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                  : 'border-dark-200 dark:border-dark-700 hover:border-dark-300 dark:hover:border-dark-600'
              }`}
            >
              <input
                type="radio"
                name="theme"
                value={option.value}
                checked={settings.theme === option.value}
                onChange={(e) => onChange('theme', e.target.value)}
                className="sr-only"
              />
              <div className="flex items-center gap-3">
                <option.icon className="h-6 w-6 text-primary-600 dark:text-primary-400" />
                <div>
                  <div className="font-medium text-dark-900 dark:text-dark-100">{option.label}</div>
                  <div className="text-sm text-dark-500 dark:text-dark-400">{option.desc}</div>
                </div>
              </div>
              {settings.theme === option.value && (
                <div className="absolute top-2 right-2 w-5 h-5 bg-primary-500 rounded-full flex items-center justify-center">
                  <Sun className="h-3 w-3 text-white" />
                </div>
              )}
            </label>
          ))}
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Language</h3>
        <select
          value={settings.language}
          onChange={(e) => onChange('language', e.target.value)}
          className="select max-w-xs"
        >
          <option value="en">English</option>
          <option value="es">Spanish</option>
          <option value="fr">French</option>
          <option value="de">German</option>
          <option value="ru">Russian</option>
          <option value="zh">Chinese</option>
          <option value="ja">Japanese</option>
          <option value="pt">Portuguese</option>
          <option value="it">Italian</option>
          <option value="ko">Korean</option>
        </select>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Behavior</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Auto-focus search input"
            description="Automatically focus the search box when visiting the home page"
            checked={settings.autoFocusSearch}
            onChange={(checked) => onChange('autoFocusSearch', checked)}
          />
          <SettingToggle
            label="Save search history"
            description="Keep a local history of your recent searches"
            checked={settings.saveHistory}
            onChange={(checked) => onChange('saveHistory', checked)}
          />
          <SettingInput
            label="Max history items"
            type="number"
            value={settings.maxHistoryItems}
            onChange={(e) => onChange('maxHistoryItems', parseInt(e.target.value) || 0)}
            min="10"
            max="1000"
            disabled={!settings.saveHistory}
          />
        </div>
      </section>
    </div>
  )
}

function AppearanceSettings({ settings, onChange }) {
  return (
    <div className="space-y-6" role="tabpanel" id="appearance-panel" aria-labelledby="appearance-tab">
      <section>
        <h3 className="text-lg font-semibold mb-4">Layout</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Show thumbnails"
            description="Display torrent thumbnails in search results"
            checked={settings.showThumbnails}
            onChange={(checked) => onChange('showThumbnails', checked)}
          />
          <SettingToggle
            label="Compact mode"
            description="Reduce spacing for more results per page"
            checked={settings.compactMode}
            onChange={(checked) => onChange('compactMode', checked)}
          />
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Results Per Page</h3>
        <select
          value={settings.resultsPerPage}
          onChange={(e) => onChange('resultsPerPage', parseInt(e.target.value))}
          className="select max-w-xs"
        >
          <option value={10}>10</option>
          <option value={20}>20</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </section>
    </div>
  )
}

function SearchSettings({ settings, onChange }) {
  return (
    <div className="space-y-6" role="tabpanel" id="search-panel" aria-labelledby="search-tab">
      <section>
        <h3 className="text-lg font-semibold mb-4">Default Search Options</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Default Category</label>
            <select
              value={settings.defaultCategory}
              onChange={(e) => onChange('defaultCategory', e.target.value)}
              className="select w-full"
            >
              <option value="all">All Categories</option>
              <option value="movies">Movies</option>
              <option value="tv">TV Shows</option>
              <option value="music">Music</option>
              <option value="games">Games</option>
              <option value="software">Software</option>
              <option value="anime">Anime</option>
              <option value="books">Books</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Default Sort</label>
            <select
              value={settings.defaultSort}
              onChange={(e) => onChange('defaultSort', e.target.value)}
              className="select w-full"
            >
              <option value="relevance">Relevance</option>
              <option value="seeders">Seeders</option>
              <option value="leechers">Leechers</option>
              <option value="size">Size</option>
              <option value="date">Date Added</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Default Order</label>
            <select
              value={settings.defaultOrder}
              onChange={(e) => onChange('defaultOrder', e.target.value)}
              className="select w-full"
            >
              <option value="desc">Descending</option>
              <option value="asc">Ascending</option>
            </select>
          </div>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Content Filters</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Safe search"
            description="Filter out potentially inappropriate content"
            checked={settings.safeSearch}
            onChange={(checked) => onChange('safeSearch', checked)}
          />
          <SettingToggle
            label="Show adult content"
            description="Include XXX category in search results"
            checked={settings.adultContent}
            onChange={(checked) => onChange('adultContent', checked)}
            disabled={settings.safeSearch}
          />
        </div>
      </section>
    </div>
  )
}

function DownloadSettings({ settings, onChange }) {
  return (
    <div className="space-y-6" role="tabpanel" id="downloads-panel" aria-labelledby="downloads-tab">
      <section>
        <h3 className="text-lg font-semibold mb-4">Download Behavior</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Auto-open magnet links"
            description="Automatically open magnet links in your torrent client"
            checked={settings.autoOpenMagnet}
            onChange={(checked) => onChange('autoOpenMagnet', checked)}
          />
          <SettingToggle
            label="Confirm before download"
            description="Show confirmation dialog before starting downloads"
            checked={settings.confirmBeforeDownload}
            onChange={(checked) => onChange('confirmBeforeDownload', checked)}
          />
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Download Path</h3>
        <SettingInput
          label="Default download folder"
          type="text"
          value={settings.downloadPath}
          onChange={(e) => onChange('downloadPath', e.target.value)}
          placeholder="Leave empty to use browser default"
        />
        <p className="text-sm text-dark-500 dark:text-dark-400">
          Note: Browsers may restrict access to file system paths. This is used as a hint for your torrent client.
        </p>
      </section>
    </div>
  )
}

function PrivacySettings({ settings, onChange }) {
  return (
    <div className="space-y-6" role="tabpanel" id="privacy-panel" aria-labelledby="privacy-tab">
      <section>
        <h3 className="text-lg font-semibold mb-4">Notifications</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Enable notifications"
            description="Show browser notifications for search completion and downloads"
            checked={settings.enableNotifications}
            onChange={async (checked) => {
              onChange('enableNotifications', checked)
              if (checked && Notification.permission === 'default') {
                await Notification.requestPermission()
              }
            }}
          />
          <SettingToggle
            label="Notification sound"
            description="Play a sound when notifications appear"
            checked={settings.notificationSound}
            onChange={(checked) => onChange('notificationSound', checked)}
            disabled={!settings.enableNotifications}
          />
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Data Management</h3>
        <div className="space-y-3">
          <button className="btn-secondary w-full justify-start" onClick={() => {
            localStorage.removeItem('search-history')
            toast.success('Search history cleared')
          }}>
            <Trash2 className="h-4 w-4" />
            Clear Search History
          </button>
          <button className="btn-secondary w-full justify-start" onClick={() => {
            localStorage.removeItem('app-settings')
            window.location.reload()
          }}>
            <Trash2 className="h-4 w-4" />
            Reset All Settings
          </button>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Proxy / VPN</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Use proxy"
            description="Route searches through a proxy server"
            checked={settings.proxyEnabled}
            onChange={(checked) => onChange('proxyEnabled', checked)}
          />
          <SettingInput
            label="Proxy URL"
            type="text"
            value={settings.proxyUrl}
            onChange={(e) => onChange('proxyUrl', e.target.value)}
            placeholder="http://proxy.example.com:8080"
            disabled={!settings.proxyEnabled}
          />
        </div>
      </section>
    </div>
  )
}

function AdvancedSettings({ settings, onChange }) {
  return (
    <div className="space-y-6" role="tabpanel" id="advanced-panel" aria-labelledby="advanced-tab">
      <section>
        <h3 className="text-lg font-semibold mb-4">Developer Options</h3>
        <div className="space-y-3">
          <SettingToggle
            label="Debug mode"
            description="Enable verbose logging in console"
            checked={false}
            onChange={() => {}}
          />
          <SettingToggle
            label="Show API response times"
            description="Display search latency in results"
            checked={false}
            onChange={() => {}}
          />
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">Cache Management</h3>
        <div className="space-y-3">
          <button className="btn-secondary w-full justify-start" onClick={() => {
            caches.keys().then(names => {
              names.forEach(name => caches.delete(name))
            })
            toast.success('Cache cleared')
          }}>
            <Trash2 className="h-4 w-4" />
            Clear Cache
          </button>
          <p className="text-sm text-dark-500 dark:text-dark-400">
            Clears cached provider data and search results. May temporarily slow down searches.
          </p>
        </div>
      </section>

      <section>
        <h3 className="text-lg font-semibold mb-4">About</h3>
        <div className="space-y-2 text-sm text-dark-600 dark:text-dark-400">
          <div className="flex justify-between">
            <span>Version</span>
            <span className="font-mono">1.0.0</span>
          </div>
          <div className="flex justify-between">
            <span>Build</span>
            <span className="font-mono">web</span>
          </div>
          <div className="flex justify-between">
            <span>Repository</span>
            <a href="https://github.com/prajwalch/TorrentSearch" target="_blank" rel="noopener noreferrer" className="text-primary-600 dark:text-primary-400 hover:underline">
              GitHub
            </a>
          </div>
        </div>
      </section>
    </div>
  )
}

function SettingToggle({ label, description, checked, onChange, disabled }) {
  return (
    <label className={`flex items-center justify-between p-4 rounded-lg border transition-colors ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:bg-dark-50 dark:hover:bg-dark-800/50'}`}>
      <div>
        <div className="font-medium text-dark-900 dark:text-dark-100">{label}</div>
        <div className="text-sm text-dark-500 dark:text-dark-400">{description}</div>
      </div>
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        disabled={disabled}
        className="w-5 h-5 text-primary-600 border-dark-300 rounded focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:focus:ring-offset-dark-900"
        aria-label={label}
      />
    </label>
  )
}

function SettingInput({ label, type, value, onChange, placeholder, disabled, min, max }) {
  return (
    <div className="space-y-1">
      <label className="block text-sm font-medium text-dark-700 dark:text-dark-300">{label}</label>
      <input
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        disabled={disabled}
        min={min}
        max={max}
        className={`input ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
      />
    </div>
  )
}

import { toast } from 'react-hot-toast'