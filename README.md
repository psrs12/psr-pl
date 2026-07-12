# psr-pl

Personal loan acquisition platform, rebuilt on Go/DynamoDB/Fargate/EventBridge per this repo's [CLAUDE.md](CLAUDE.md) enterprise Go standards. Corporate-directed rewrite — service boundaries and applicant-facing behavior are preserved from the reference implementation (`psr-personal-loan`, a separate, unmodified repository used only as functional source of truth); the tech stack and internal architecture are deliberately replaced.

See `openspec/changes/rebuild-platform-go/` for the full proposal, design decisions, and specs behind this rewrite.

## Structure

```
application-management-service/   Go service: applicant intake, submission, self-service login
application-management-ui/        React/Vite UI (pre-existing, wired to the service above)
openspec/                         OpenSpec change proposals and specs
CLAUDE.md                         Architecture standards this codebase follows
```

Only `application-management-service` exists so far. The platform's other four services (`pricing-orchestration-service`, `offer-acceptance-service`, `document-service`, `compliance-orchestration-service`) are scoped in the OpenSpec change but not yet built. `invitation-service` is deliberately **not** a standalone service — it's an internal adapter inside `application-management-service` (see `openspec/changes/rebuild-platform-go/design.md` for why) that talks to an external OfferLog system.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) 20+ and npm
- [Docker](https://www.docker.com/) (for DynamoDB Local)
- [AWS CLI](https://aws.amazon.com/cli/) (used against DynamoDB Local only — no real AWS account needed for local dev)

## Running the backend

`application-management-service` needs DynamoDB and the external OfferLog system. For local dev, both are stood up automatically:

```bash
cd application-management-service
make run
```

This single command:
1. Starts DynamoDB Local in Docker (`pl-application` table, TTL enabled for sessions) if not already running.
2. Starts a local OfferLog stand-in (`cmd/offerlog-mock`) if not already running.
3. Runs the service on `:8081`.

Fixture invitation tokens for the OfferLog mock: `VALID-INVITE-123` (active offer), `USED-INVITE-456` (already used, to test graceful degradation), or any other token starting with `VALID-` (mints a fresh active offer on first use).

Stop everything with:

```bash
make down
```

Other targets: `make build`, `make vet`, `make fmt`, `make test`. See the [Makefile](application-management-service/Makefile) for the full list and configurable ports/table name.

Environment variables the service reads directly (see `configs/example.env`): `APPLICATION_TABLE_NAME`, `OFFERLOG_BASE_URL`, `PORT`, `DYNAMODB_ENDPOINT_URL` (local-dev only — leave unset in real AWS environments to use normal endpoint resolution), `CORS_ALLOWED_ORIGIN` (defaults to `*`; only matters for local dev since the UI and API share an origin in production).

## Running the UI

```bash
cd application-management-ui
npm install
npm run dev
```

Opens on `http://localhost:3000`. `.env.local` is already configured to point at the backend on `:8081`, matching `make run`'s default port. Try:
- `http://localhost:3000/apply` — direct application flow
- `http://localhost:3000/apply?token=VALID-INVITE-123` — invitation flow (prefilled, locked name/address)
- `http://localhost:3000/portal/login` — self-service status login (needs an application ID from a prior submission)

Run backend and UI in separate terminals; both need to be up for the UI to actually work end-to-end.

## Troubleshooting

- **"address already in use"** — a previous `make run` didn't shut down cleanly (`go run` execs a child binary that can outlive a killed parent process). Run `make down`, which kills by port rather than by PID file, then retry.
- **"Failed to fetch" in the browser** — usually a port mismatch between the UI's `VITE_APP_MANAGEMENT_API_URL` (`application-management-ui/.env.local`) and the port the service is actually running on (`PORT` in the Makefile, default `8081`).
