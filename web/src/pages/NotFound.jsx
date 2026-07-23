import { Link } from 'react-router-dom'
import { Home, Search, ArrowLeft } from 'lucide-react'

export default function NotFound() {
  return (
    <div className="min-h-[60vh] flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <div className="w-24 h-24 mx-auto mb-6 bg-dark-100 dark:bg-dark-800 rounded-full flex items-center justify-center">
          <Search className="h-12 w-12 text-dark-400 dark:text-dark-500" />
        </div>
        
        <h1 className="text-4xl font-bold text-dark-900 dark:text-dark-100 mb-3">404</h1>
        <h2 className="text-xl text-dark-600 dark:text-dark-400 mb-4">Page Not Found</h2>
        <p className="text-dark-500 dark:text-dark-400 mb-8">
          The page you're looking for doesn't exist or has been moved.
        </p>
        
        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <Link to="/" className="btn-primary">
            <Home className="h-4 w-4 mr-2" />
            Go Home
          </Link>
          <Link to="/search" className="btn-secondary">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Search Torrents
          </Link>
        </div>
      </div>
    </div>
  )
}