import { Link, Route, Routes } from 'react-router-dom'

import { Dashboard } from './pages/Dashboard'

function DetailPlaceholder() {
  return (
    <main className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <Link to="/" className="inline-flex min-h-11 items-center whitespace-nowrap text-sm font-semibold text-accent">
        ← Back to clusters
      </Link>
      <p className="mt-8 text-muted">Cluster detail is being prepared.</p>
    </main>
  )
}

export function App() {
  return (
    <Routes>
      <Route path="/" element={<Dashboard />} />
      <Route path="/clusters/:id" element={<DetailPlaceholder />} />
    </Routes>
  )
}
