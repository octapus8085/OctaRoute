const samplePolicies = [
  {
    name: 'Allow Internal',
    source: 'us-east',
    destination: 'eu-central',
    action: 'allow'
  },
  {
    name: 'Block Unknown',
    source: 'unknown',
    destination: 'core-services',
    action: 'deny'
  }
]

export default function Policies() {
  return (
    <section>
      <header className="page-header">
        <h1>Policies</h1>
        <p>Routing and security policies applied by the controller.</p>
      </header>
      <table className="table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Source</th>
            <th>Destination</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {samplePolicies.map((policy) => (
            <tr key={policy.name}>
              <td>{policy.name}</td>
              <td>{policy.source}</td>
              <td>{policy.destination}</td>
              <td>{policy.action}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}
