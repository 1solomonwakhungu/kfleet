import { Navigate, Route, Routes } from 'react-router-dom'

import { ApplicationShell } from './components/layout/ApplicationShell'
import ClusterDetail from './pages/ClusterDetail'
import { Dashboard } from './pages/Dashboard'
import PendingAgents from './pages/PendingAgents'

export function App() {
  return (
    <Routes>
      <Route element={<ApplicationShell />}>
        <Route index element={<Dashboard />} />
        <Route path="clusters/:id" element={<ClusterDetail />} />
        <Route path="agents" element={<PendingAgents />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
