const sampleNodes = [
  {
    name: 'exit-us-east-1',
    address: '100.64.10.1',
    zone: 'us-east'
  },
  {
    name: 'gateway-eu-central',
    address: '100.64.20.2',
    zone: 'eu-central'
  }
]

export default function Nodes() {
  return (
    <section>
      <header className="page-header">
        <h1>Nodes</h1>
        <p>Registered exit and gateway nodes.</p>
      </header>
      <table className="table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Address</th>
            <th>Zone</th>
          </tr>
        </thead>
        <tbody>
          {sampleNodes.map((node) => (
            <tr key={node.name}>
              <td>{node.name}</td>
              <td>{node.address}</td>
              <td>{node.zone}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}
