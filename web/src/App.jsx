import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ProvidersProvider } from './context/ProvidersContext'
import { ThemeProvider } from './context/ThemeContext'
import Layout from './components/Layout'
import Home from './pages/Home'
import SearchResults from './pages/SearchResults'
import TorrentDetails from './pages/TorrentDetails'
import Providers from './pages/Providers'
import Settings from './pages/Settings'
import NotFound from './pages/NotFound'

function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <ProvidersProvider>
          <Layout>
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/search" element={<SearchResults />} />
              <Route path="/torrent/:provider/:detailUrl(*)" element={<TorrentDetails />} />
              <Route path="/providers" element={<Providers />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<NotFound />} />
            </Routes>
          </Layout>
        </ProvidersProvider>
      </ThemeProvider>
    </BrowserRouter>
  )
}

export default App