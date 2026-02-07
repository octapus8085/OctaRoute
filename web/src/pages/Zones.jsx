const sampleZones = [
  {
    name: 'us-east',
    description: 'Primary exits for US East.'
  },
  {
    name: 'eu-central',
    description: 'Regional gateways for Europe.'
  }
]

export default function Zones() {
  return (
    <section>
      <header className="page-header">
        <h1>Zones</h1>
        <p>Logical groupings for routing policy decisions.</p>
      </header>
      <div className="cards">
        {sampleZones.map((zone) => (
          <div className="card" key={zone.name}>
            <h3>{zone.name}</h3>
            <p>{zone.description}</p>
          </div>
        ))}
      </div>
    </section>
  )
}
