import { NavLink, Route, Routes } from 'react-router-dom'
import Dashboard from './pages/Dashboard.jsx'
import Nodes from './pages/Nodes.jsx'
import Zones from './pages/Zones.jsx'
import Policies from './pages/Policies.jsx'

const routes = [
  { path: '/', label: 'Dashboard', element: <Dashboard /> },
  { path: '/nodes', label: 'Nodes', element: <Nodes /> },
  { path: '/zones', label: 'Zones', element: <Zones /> },
  { path: '/policies', label: 'Policies', element: <Policies /> }
]

export default function App() {
  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">OctaRoute</div>
        <nav>
          {routes.map((route) => (
            <NavLink
              key={route.path}
              to={route.path}
              className={({ isActive }) =>
                isActive ? 'nav-link active' : 'nav-link'
              }
              end={route.path === '/'}
            >
              {route.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Routes>
          {routes.map((route) => (
            <Route key={route.path} path={route.path} element={route.element} />
          ))}
        </Routes>
      </main>
    </div>
  )
}
