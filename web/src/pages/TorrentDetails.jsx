import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { 
  Magnet, Download, Copy, ExternalLink, Share2, 
  Loader2, AlertCircle, Check, FileText, 
  ArrowUp, ArrowDown, Calendar, Hash, Tag,
  ChevronDown, ChevronUp
} from 'lucide-react'
import { api } from '../lib/api'
import { useToast } from '../hooks/useToast'

export default function TorrentDetails() {
  const { provider, detailUrl } = useParams()
  const { toast } = useToast()
  const [torrent, setTorrent] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [copied, setCopied] = useState(false)
  const [showFiles, setShowFiles] = useState(false)

  const fetchDetails = useCallback(async () => {
    try {
      setLoading(true)
      const res = await api.get('/api/torrent/details', {
        params: { provider, url: detailUrl }
      })
      setTorrent(res.data.torrent)
    } catch (err) {
      setError(err.message || 'Failed to load torrent details')
      console.error('Details error:', err)
    } finally {
      setLoading(false)
    }
  }, [provider, detailUrl])

  useEffect(() => {
    fetchDetails()
  }, [fetchDetails])

  const copyToClipboard = async (text, label) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      toast.success(`${label} copied to clipboard`)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      toast.error(`Failed to copy ${label}`)
    }
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
    return `${size.toFixed(2)} ${units[unitIndex]}`
  }

  const formatDate = (dateStr) => {
    if (!dateStr) return 'Unknown'
    return new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const formatNumber = (num) => {
    if (!num) return 'Unknown'
    return num.toLocaleString()
  }

  if (loading) {
    return <DetailsSkeleton />
  }

  if (error || !torrent) {
    return (
      <div className="card p-12 text-center">
        <AlertCircle className="mx-auto h-16 w-16 text-red-500 mb-4" aria-hidden="true" />
        <h2 className="text-2xl font-bold text-dark-900 dark:text-dark-100 mb-2">Unable to Load Details</h2>
        <p className="text-dark-500 dark:text-dark-400 mb-6 max-w-md mx-auto">
          {error || 'The torrent details could not be loaded. The provider may be unavailable or the torrent may have been removed.'}
        </p>
        <Link to="/search" className="btn-primary inline-flex">
          <Magnet className="h-4 w-4" aria-hidden="true" />
          Search Again
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="card p-6">
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-3 flex-wrap">
              <span className="badge badge-primary">{torrent.category}</span>
              <span className="badge badge-success">{torrent.provider}</span>
              {torrent.verified && <span className="badge badge-success flex items-center gap-1"><Check className="h-3 w-3" /> Verified</span>}
            </div>
            <h1 className="text-2xl md:text-3xl font-bold text-dark-900 dark:text-dark-100 mb-2 line-clamp-3">
              {torrent.name}
            </h1>
            <div className="flex flex-wrap items-center gap-4 text-sm text-dark-500 dark:text-dark-400">
              <span className="flex items-center gap-1">
                <Hash className="h-4 w-4" aria-hidden="true" />
                {torrent.size && formatSize(torrent.size)}
              </span>
              <span className="flex items-center gap-1">
                <ArrowUp className="h-4 w-4 text-green-500" aria-hidden="true" />
                {formatNumber(torrent.seeders)} Seeders
              </span>
              <span className="flex items-center gap-1">
                <ArrowDown className="h-4 w-4 text-red-500" aria-hidden="true" />
                {formatNumber(torrent.leechers)} Leechers
              </span>
              {torrent.uploaded && (
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" aria-hidden="true" />
                  {formatDate(torrent.uploaded)}
                </span>
              )}
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {torrent.magnet && (
              <button
                onClick={() => copyToClipboard(torrent.magnet, 'Magnet link')}
                className="btn-primary flex items-center gap-2"
              >
                <Magnet className="h-4 w-4" aria-hidden="true" />
                {copied ? <Check className="h-4 w-4" /> : 'Copy Magnet'}
              </button>
            )}
            {torrent.torrentFile && (
              <a
                href={torrent.torrentFile}
                target="_blank"
                rel="noopener noreferrer"
                className="btn-secondary flex items-center gap-2"
              >
                <Download className="h-4 w-4" aria-hidden="true" />
                Download .torrent
              </a>
            )}
            {torrent.detailUrl && (
              <a
                href={torrent.detailUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="btn-ghost flex items-center gap-2"
              >
                <ExternalLink className="h-4 w-4" aria-hidden="true" />
                View on Site
              </a>
            )}
          </div>
        </div>
      </div>

      <div className="grid lg:grid-cols-3 gap-6">
        {/* Main Info */}
        <div className="lg:col-span-2 space-y-6">
          {/* Description */}
          {torrent.description && (
            <section className="card">
              <div className="p-6 border-b border-dark-200 dark:border-dark-800">
                <h2 className="text-lg font-semibold">Description</h2>
              </div>
              <div className="p-6 prose dark:prose-invert max-w-none">
                <div 
                  className="whitespace-pre-wrap text-dark-700 dark:text-dark-300"
                  dangerouslySetInnerHTML={{ __html: torrent.description }}
                />
              </div>
            </section>
          )}

          {/* Files */}
          {torrent.files && torrent.files.length > 0 && (
            <section className="card">
              <div className="p-6 border-b border-dark-200 dark:border-dark-800 flex items-center justify-between">
                <h2 className="text-lg font-semibold">Files ({torrent.files.length})</h2>
                <button
                  onClick={() => setShowFiles(!showFiles)}
                  className="btn-ghost text-sm flex items-center gap-1"
                >
                  {showFiles ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                  {showFiles ? 'Collapse' : 'Expand'}
                </button>
              </div>
              {showFiles && (
                <div className="p-6">
                  <div className="overflow-x-auto">
                    <table className="table w-full">
                      <thead>
                        <tr>
                          <th>Name</th>
                          <th className="text-right">Size</th>
                        </tr>
                      </thead>
                      <tbody>
                        {torrent.files.map((file, index) => (
                          <tr key={index}>
                            <td className="font-mono text-sm">{file.name}</td>
                            <td className="text-right text-dark-500 dark:text-dark-400">
                              {file.size && formatSize(file.size)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </section>
          )}

          {/* Technical Details */}
          <section className="card">
            <div className="p-6 border-b border-dark-200 dark:border-dark-800">
              <h2 className="text-lg font-semibold">Technical Details</h2>
            </div>
            <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
              <DetailRow label="Info Hash" value={torrent.infoHash} copyable />
              <DetailRow label="Provider" value={torrent.provider} />
              <DetailRow label="Category" value={torrent.category} />
              <DetailRow label="Size" value={torrent.size ? formatSize(torrent.size) : 'Unknown'} />
              <DetailRow label="Seeders" value={formatNumber(torrent.seeders)} />
              <DetailRow label="Leechers" value={formatNumber(torrent.leechers)} />
              <DetailRow label="Uploaded" value={torrent.uploaded ? formatDate(torrent.uploaded) : 'Unknown'} />
              <DetailRow label="Verified" value={torrent.verified ? 'Yes' : 'No'} />
              {torrent.uploader && (
                <DetailRow label="Uploader" value={torrent.uploader} />
              )}
              {torrent.imdbId && (
                <DetailRow label="IMDb ID" value={torrent.imdbId} />
              )}
              {torrent.tmdbId && (
                <DetailRow label="TMDb ID" value={torrent.tmdbId} />
              )}
            </div>
          </section>
        </div>

        {/* Sidebar */}
        <aside className="space-y-6">
          {/* Thumbnail */}
          {torrent.thumbnail && (
            <div className="card overflow-hidden">
              <img
                src={torrent.thumbnail}
                alt={torrent.name}
                className="w-full aspect-video object-cover"
              />
            </div>
          )}

          {/* Quick Actions */}
          <div className="card p-6 space-y-3">
            <h3 className="font-semibold">Quick Actions</h3>
            <div className="space-y-2">
              {torrent.magnet && (
                <button
                  onClick={() => copyToClipboard(torrent.magnet, 'Magnet link')}
                  className="btn-secondary w-full justify-start"
                >
                  <Magnet className="h-4 w-4" aria-hidden="true" />
                  Copy Magnet Link
                </button>
              )}
              {torrent.torrentFile && (
                <a
                  href={torrent.torrentFile}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn-secondary w-full justify-start"
                >
                  <Download className="h-4 w-4" aria-hidden="true" />
                  Download .torrent File
                </a>
              )}
              {torrent.detailUrl && (
                <a
                  href={torrent.detailUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn-ghost w-full justify-start"
                >
                  <ExternalLink className="h-4 w-4" aria-hidden="true" />
                  Open in Browser
                </a>
              )}
              <button
                onClick={() => {
                  if (navigator.share && torrent.magnet) {
                    navigator.share({
                      title: torrent.name,
                      text: torrent.magnet,
                      url: window.location.href
                    })
                  } else if (torrent.magnet) {
                    copyToClipboard(torrent.magnet, 'Magnet link')
                  }
                }}
                className="btn-ghost w-full justify-start"
              >
                <Share2 className="h-4 w-4" aria-hidden="true" />
                Share
              </button>
            </div>
          </div>

          {/* Related */}
          {torrent.related && torrent.related.length > 0 && (
            <div className="card p-6">
              <h3 className="font-semibold mb-4">Related Torrents</h3>
              <div className="space-y-3">
                {torrent.related.slice(0, 5).map((related, index) => (
                  <Link
                    key={index}
                    to={`/torrent/${related.provider}/${encodeURIComponent(related.detailUrl || related.magnet || related.id)}`}
                    className="block p-3 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-800 transition-colors"
                  >
                    <p className="text-sm font-medium text-dark-900 dark:text-dark-100 line-clamp-1">
                      {related.name}
                    </p>
                    <p className="text-xs text-dark-500 dark:text-dark-400 flex items-center gap-1 mt-1">
                      <ArrowUp className="h-3 w-3 text-green-500" aria-hidden="true" />
                      {formatNumber(related.seeders)}
                      <ArrowDown className="h-3 w-3 text-red-500 ml-2" aria-hidden="true" />
                      {formatNumber(related.leechers)}
                    </p>
                  </Link>
                ))}
              </div>
            </div>
          )}
        </aside>
      </div>
    </div>
  )
}

function DetailRow({ label, value, copyable }) {
  const [copied, setCopied] = useState(false)
  
  const handleCopy = async () => {
    if (!copyable || !value) return
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Copy failed:', err)
    }
  }

  return (
    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 p-3 bg-dark-50 dark:bg-dark-800/50 rounded-lg">
      <span className="text-sm text-dark-500 dark:text-dark-400">{label}</span>
      <div className="flex items-center gap-2">
        <span className="text-sm font-mono text-dark-900 dark:text-dark-100 break-all max-w-[200px] sm:max-w-none">
          {value || '—'}
        </span>
        {copyable && value && (
          <button
            onClick={handleCopy}
            className="p-1.5 text-dark-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
            aria-label={`Copy ${label}`}
          >
            {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
          </button>
        )}
      </div>
    </div>
  )
}

function DetailsSkeleton() {
  return (
    <div className="space-y-6">
      <div className="card p-6">
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div className="flex-1">
            <div className="flex gap-2 mb-3">
              <div className="skeleton badge" style={{ width: '80px' }} />
              <div className="skeleton badge" style={{ width: '80px' }} />
            </div>
            <div className="skeleton-title mb-2" style={{ width: '60%' }} />
            <div className="skeleton-text" style={{ width: '40%' }} />
            <div className="skeleton-text" style={{ width: '30%' }} />
          </div>
          <div className="flex gap-2">
            <div className="skeleton btn-primary" style={{ width: '140px', height: '40px' }} />
            <div className="skeleton btn-secondary" style={{ width: '140px', height: '40px' }} />
          </div>
        </div>
      </div>
      <div className="grid lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="card p-6">
            <div className="skeleton-title mb-4" />
            <div className="space-y-3">
              <div className="skeleton-text" />
              <div className="skeleton-text" />
              <div className="skeleton-text" />
              <div className="skeleton-text" />
              <div className="skeleton-text" />
            </div>
          </div>
          <div className="card p-6">
            <div className="skeleton-title mb-4" />
            <div className="grid grid-cols-2 gap-4">
              <div className="skeleton-text" /><div className="skeleton-text" />
              <div className="skeleton-text" /><div className="skeleton-text" />
              <div className="skeleton-text" /><div className="skeleton-text" />
              <div className="skeleton-text" /><div className="skeleton-text" />
            </div>
          </div>
        </div>
        <aside className="space-y-6">
          <div className="card aspect-video skeleton" />
          <div className="card p-6">
            <div className="skeleton-title mb-4" />
            <div className="space-y-2">
              <div className="skeleton btn-secondary" style={{ width: '100%', height: '40px' }} />
              <div className="skeleton btn-secondary" style={{ width: '100%', height: '40px' }} />
              <div className="skeleton btn-ghost" style={{ width: '100%', height: '40px' }} />
              <div className="skeleton btn-ghost" style={{ width: '100%', height: '40px' }} />
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}
