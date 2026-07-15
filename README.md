# psr-pl

[![CI](https://github.com/psrs12/psr-pl/actions/workflows/ci.yml/badge.svg)](https://github.com/psrs12/psr-pl/actions/workflows/ci.yml)

Personal loan acquisition platform, rebuilt on Go/DynamoDB/Fargate/EventBridge per this repo's [CLAUDE.md](CLAUDE.md) enterprise Go standards. Corporate-directed rewrite — service boundaries and applicant-facing behavior are preserved from the reference implementation (`psr-personal-loan`, a separate, unmodified repository used only as functional source of truth); the tech stack and internal architecture are deliberately replaced.

See `openspec/changes/rebuild-platform-go/` for the full proposal, design decisions, and specs behind this rewrite.

## Structure

```
application-management-service/   Go service: applicant intake, submission, self-service login
pricing-orchestration-service/    Go: Step Functions state machine + Lambda/Fargate steps, plus a
                                   narrow offer-selection REST API (the one deliberate exception)
workflow-status-service/          Go service: applicant-facing status/next-steps, derived live
                                   from the Step Functions execution (no state of its own)
offer-acceptance-service/         Go service: declarations + e-signature capture
application-management-ui/       React/Vite UI (pre-existing, wired to the services above)
pricing-offers-ui/                <pricing-offer-selector> web component (vanilla JS, not React —
                                   see design.md), loaded by application-management-ui
openspec/                         OpenSpec change proposals and specs
CLAUDE.md                         Architecture standards this codebase follows
```

`offer-acceptance-service`'s peer, `document-service`, and `compliance-orchestration-service` are scoped in the OpenSpec change but not yet built. `invitation-service` is deliberately **not** a standalone service — it's an internal adapter inside `application-management-service` (see `openspec/changes/rebuild-platform-go/design.md` for why) that talks to an external OfferLog system.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) 20+ and npm
- [Docker](https://www.docker.com/) (for DynamoDB Local and Step Functions Local)
- [AWS CLI](https://aws.amazon.com/cli/) (used against local emulators only — no real AWS account needed for local dev)

## Running the backend

Start in this order — each depends on the ones before it:

**1. Step Functions Local + the pricing state machine** (needed before `application-management-service`, since submitting an application starts an execution):
```bash
cd pricing-orchestration-service
make statemachine-up   # Step Functions Local on :8083, registers the state machine
```

**2. `application-management-service`** (DynamoDB Local + the OfferLog mock, both automatic):
```bash
cd application-management-service
make run   # :8081
```
Fixture invitation tokens for the OfferLog mock: `VALID-INVITE-123` (active offer), `USED-INVITE-456` (already used, to test graceful degradation), or any other token starting with `VALID-` (mints a fresh active offer on first use).

**3. `pricing-orchestration-service`'s offer-selection API** (separate from step 1 — that brought up Step Functions Local, this runs the actual API server):
```bash
cd pricing-orchestration-service
make run   # :8082, requires application-management-service running for session validation
```

**4. `workflow-status-service`**:
```bash
cd workflow-status-service
APPLICATION_MANAGEMENT_BASE_URL=http://localhost:8081 \
STEPFUNCTIONS_ENDPOINT_URL=http://localhost:8083 \
AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_DEFAULT_REGION=us-east-1 \
PORT=8086 go run ./cmd/server
```

**5. `offer-acceptance-service`**:
```bash
cd offer-acceptance-service
make run   # :8085, requires application-management-service running
```

Each service's own `make down` (or `make stepfunctions-down` / `make dynamodb-down`) tears down what it started. `make stop` frees a service's own port if a previous `go run` didn't shut down cleanly (see Troubleshooting).

Environment variables each service reads directly are documented in its own `configs/example.env` (where present) and `Makefile`. Common ones: `*_TABLE_NAME`, `*_BASE_URL` (service-to-service calls), `PORT`, `DYNAMODB_ENDPOINT_URL`/`STEPFUNCTIONS_ENDPOINT_URL` (local-dev only — leave unset in real AWS environments), `CORS_ALLOWED_ORIGIN` (defaults to `*`; only matters for local dev since the UI and API share an origin in production).

## Running the UI

```bash
cd application-management-ui
npm install
npm run dev
```

Opens on `http://localhost:3000`. `.env.local` is already configured to point at the backend ports above. Try:
- `http://localhost:3000/apply` — direct application flow
- `http://localhost:3000/apply?token=VALID-INVITE-123` — invitation flow (prefilled, locked name/address)
- `http://localhost:3000/portal/login` — self-service status login (needs an application ID from a prior submission)

**`pricing-offers-ui`** is loaded by `application-management-ui` as a `<script>` tag (a built IIFE, not a dev-server module) — `vite dev` won't serve the right file for this one:
```bash
cd pricing-offers-ui
npm install
npm run build
cd dist && python3 -m http.server 3001
```

Run all of these in separate terminals; all need to be up for the UI to actually work end-to-end.

## Troubleshooting

- **"address already in use"** — a previous `make run` didn't shut down cleanly (`go run` execs a child binary that can outlive a killed parent process). Run that service's `make down`/`make stop`, which kills by port rather than by PID file, then retry.
- **"Failed to fetch" in the browser** — usually a port mismatch between the UI's env vars (`application-management-ui/.env.local`) and the port a service is actually running on.
- **`pricing-offer-selector` doesn't render** — `vite dev` in `pricing-offers-ui` serves ES modules, not the built `pricing-offer-selector.iife.js` filename the UI's `<script>` tag expects. Build it and serve `dist/` statically instead (see above).
