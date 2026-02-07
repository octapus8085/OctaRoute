export default function Dashboard() {
  return (
    <section>
      <header className="page-header">
        <h1>Dashboard</h1>
        <p>Overview of OctaRoute control plane activity.</p>
      </header>
      <div className="cards">
        <div className="card">
          <div className="badge">Nodes</div>
          <h2>0 Active</h2>
          <p>Connect exit and gateway nodes to begin routing.</p>
        </div>
        <div className="card">
          <div className="badge">Policies</div>
          <h2>0 Draft</h2>
          <p>Create security policies to enforce routing rules.</p>
        </div>
        <div className="card">
          <div className="badge">Zones</div>
          <h2>0 Zones</h2>
          <p>Group nodes into zones to simplify policy management.</p>
        </div>
      </div>
    </section>
  )
}
