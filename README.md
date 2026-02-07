# OctaRoute

OctaRoute provides exit, gateway, and controller daemons with a lightweight control plane UI.

## Repository Layout

- `cmd/` Go binaries for exit, gateway, and controller services.
- `internal/` Shared Go packages.
- `web/` React + Vite control plane UI.
- `scripts/` Install/verify helpers.
- `systemd/` Example unit files.

## Build

```bash
# build all binaries

go build ./cmd/octaroute-exitd

go build ./cmd/octaroute-gatewayd

go build ./cmd/octaroute-controller
```

## Run (local)

Create a JSON config file for each service:

```json
{
  "server": {
    "address": ":8080",
    "bindTailscale": false
  },
  "database": "octaroute.db"
}
```

Run each service with a distinct address:

```bash
./octaroute-controller --config ./controller.json
./octaroute-exitd --config ./exitd.json
./octaroute-gatewayd --config ./gatewayd.json
```

The controller exposes API endpoints:

- `GET /api/nodes`
- `POST /api/nodes`
- `GET /api/policies`
- `POST /api/policies`
- `GET /api/routes`
- `POST /api/routes`

## Web UI

```bash
cd web
npm install
npm run dev
```

The UI is a placeholder dashboard with nodes, zones, and policies pages.

## Install helpers

```bash
./scripts/install_exit.sh
./scripts/install_gateway.sh
```

The scripts build and copy binaries to `/usr/local/bin`. Systemd unit templates live in `systemd/`.

## Verify scaffolding

```bash
./scripts/verify.sh
```
