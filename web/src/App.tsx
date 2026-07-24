import { Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom'

import { useAuth } from './auth/AuthContext'
import { ApplicationShell } from './components/layout/ApplicationShell'
import ClusterDetail from './pages/ClusterDetail'
import { Dashboard } from './pages/Dashboard'
import { Login } from './pages/Login'
import PendingAgents from './pages/PendingAgents'
import PolicyDashboard from './pages/PolicyDashboard'

export function App() {
  return (
    <Routes>
      <Route path="login" element={<Login />} />
      <Route element={<RequireAuthentication />}>
        <Route element={<ApplicationShell />}>
          <Route index element={<Dashboard />} />
          <Route path="clusters/:id" element={<ClusterDetail />} />
          <Route path="agents" element={<PendingAgents />} />
          <Route path="policies" element={<PolicyDashboard />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

function RequireAuthentication() {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <main className="grid min-h-dvh place-items-center bg-background text-sm text-muted">
        Checking session…
      </main>
    )
  }
  if (!user) {
    return <Navigate to="/login" state={{ from: location.pathname + location.search }} replace />
  }
  return <Outlet />
}
