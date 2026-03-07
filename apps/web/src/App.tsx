import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import StatusPage from './pages/StatusPage'
import AdminLogin from './pages/admin/AdminLogin'
import AdminLayout from './pages/admin/AdminLayout'
import AdminDashboard from './pages/admin/AdminDashboard'
import AdminComponents from './pages/admin/AdminComponents'
import AdminSubComponents from './pages/admin/AdminSubComponents'
import AdminIncidents from './pages/admin/AdminIncidents'
import AdminMaintenance from './pages/admin/AdminMaintenance'
import AdminMonitors from './pages/admin/AdminMonitors'
import AdminSubscribers from './pages/admin/AdminSubscribers'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('admin_token')
  if (!token) return <Navigate to="/admin/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      {/* Public status page */}
      <Route path="/" element={<StatusPage />} />

      {/* Admin auth */}
      <Route path="/admin/login" element={<AdminLogin />} />

      {/* Admin protected routes */}
      <Route
        path="/admin"
        element={
          <ProtectedRoute>
            <AdminLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<AdminDashboard />} />
        <Route path="components" element={<AdminComponents />} />
        <Route path="subcomponents" element={<AdminSubComponents />} />
        <Route path="incidents" element={<AdminIncidents />} />
        <Route path="maintenance" element={<AdminMaintenance />} />
        <Route path="monitors" element={<AdminMonitors />} />
        <Route path="subscribers" element={<AdminSubscribers />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
