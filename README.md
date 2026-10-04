# golang-fido-webauthn

Study backend for learning FIDO2 / WebAuthn (passkeys) in Go, built with the same
Clean Architecture skeleton used in the production `gestao.one` services. It starts
as a standalone Gin + PostgreSQL service and is designed to later fold into `app-cam`
(the auth/RBAC service) and run serverless.

## Goal

Learn the WebAuthn ceremony end to end (registration and login), with the real
engineering conventions: Clean Architecture, consumer-driven interfaces, explicit
pointer vs value semantics, structured logging, table-driven tests, and CI. Each
endpoint is built as a full vertical slice, from HTTP down to persistence.

## Stack

| Concern | Current (local study) | Target (production) |
| --- | --- | --- |
| Language | Go 1.25 | Go 1.25 |
| HTTP | Gin | Gin (via API Gateway) |
| Compute | standalone binary | AWS Lambda (`provided.al2023`, `arm64`) + API Gateway REST |
| Database | PostgreSQL 18 in Docker (no volume, ephemeral) | AWS RDS PostgreSQL |
| ORM | GORM (`gorm.io/driver/postgres`) | same |
| WebAuthn | `github.com/go-webauthn/webauthn` | same |
| IDs | `github.com/google/uuid` (UUIDv7) | same |
| Logging | `log/slog` (JSON) | same, shipped to CloudWatch |
| Tests | `testing` + `testify`, Testcontainers for integration | same |
| CI | GitHub Actions (build, vet, test -race) | same |

The service is written framework-light at the edges so the move to Lambda is a change
of entrypoint (`cmd/lambda/main.go`) and of the ceremony/session store, not a redesign.

## Architecture

Clean Architecture with the dependency arrows pointing inward. The domain depends on
nothing; application depends on domain; infrastructure depends on domain. No framework
tags (gorm, json) leak into the domain entities.

```
cmd/api/                         process entrypoint, dependency wiring
internal/
  fido/
    domain/                      pure entities + repository interfaces + sentinel errors
    application/                 usecases (interface + impl), input structs
    infrastructure/
      http/                      handlers, DTOs, DTO<->domain mappers, error mapping
      repository/                GORM models, model<->domain mappers, DAOs
      webauthn/                  WebAuthn provider (RP config) and ceremony glue
  infrastructure/
    http/                        shared router + middleware (flow-id, logging)
pkg/
  flowid/                        request correlation id in context (framework-neutral)
  writeerror/                    generic HTTP error response writer
  ptr/                           generic Ptr[T] helper
sql/init/                        schema files, applied on container creation
web/                             browser test page (later phase)
```

### Cross-cutting behavior

- Every business request must carry a valid `X-Flow-ID` header (UUID). The middleware
  rejects a missing or malformed value with `400`, propagates the id through
  `context.Context` so every layer logs with the same correlation id, and echoes it
  back on the response. `GET /health` is exempt so infrastructure probes work without
  the header.
- Each request logs `request received` and `request completed` (method, path, status,
  latency) as structured JSON, tied to the flow id.
- Domain errors are mapped to HTTP status in one place per module
  (`infrastructure/http/errors.go`), so the domain never imports HTTP.

## Running locally

```bash
# start ephemeral Postgres (schema in sql/init is applied on container creation)
docker compose up -d

# run the API
go run ./cmd/api

# health check
curl -i localhost:8080/health
```

Quality gates (same as CI):

```bash
go build ./...
go vet ./...
go test -race ./...
```

## Endpoints

Base group: `/fido`. All business routes require a valid `X-Flow-ID` header.

### Implemented

#### `POST /fido/users` — create a WebAuthn user

Creates the user record that a passkey will later be bound to. A WebAuthn credential
needs a stable, opaque, non-PII user handle, so the user is created first and the
ceremony references it afterwards.

Request:

```json
{ "username": "padme", "display_name": "Padme Amidala" }
```

Behavior:

- Validates both fields are present (`400` on missing/malformed body).
- Rejects a duplicate username with `409`.
- Generates a UUIDv7 `hash` that doubles as the WebAuthn user handle (carries no PII).
- Returns `201` with the created user (hash, username, display_name, timestamps).
  The internal numeric id and audit fields are not exposed.

#### `GET /health` — liveness

Returns `200 {"status":"ok"}`. Exempt from the flow-id requirement.

### Planned — WebAuthn ceremony

WebAuthn is a two-step ceremony. The `begin` step issues a cryptographic challenge that
the `finish` step must validate, so the challenge (`SessionData`) is persisted between
the two calls (in `fido_ceremony`), single-use, and expired on read, never trusted past
its deadline.

A real authenticator is required to complete `finish` (Touch ID, YubiKey, phone, or the
browser's virtual authenticator). A plain HTTP client like Postman can call `begin` but
cannot complete `finish`, because it cannot produce the signature.

#### `POST /fido/register/begin` — start passkey registration

For an already authenticated/existing user who is adding a passkey to their account.

Behavior:

- Loads the user by hash; `404` if unknown.
- Builds the RP-scoped WebAuthn options and generates a challenge.
- Persists the `SessionData` tied to a `ceremony_id` with a short TTL.
- Returns `200` with the `PublicKeyCredentialCreationOptions` (JSON the browser passes
  to `navigator.credentials.create()`) plus the `ceremony_id` the client returns on finish.

#### `POST /fido/register/finish` — complete passkey registration

Behavior:

- Consumes the ceremony session atomically; if it is missing or expired, `400`
  (the client must restart from `begin`). The challenge is single-use.
- Verifies the attestation produced by the authenticator against the stored challenge.
- On success, stores the new credential (credential id, public key, sign count, AAGUID,
  transports, backup flags) bound to the user, and returns `201`.
- On verification failure, `400`. The challenge is already burned, so a retry needs a
  fresh `begin`.

#### `POST /fido/login/begin` — start passkey login

Public route (this is how a user authenticates, so no prior session exists). In
production this lives outside the authenticated group, next to token issuance.

Behavior:

- Generates an assertion challenge (for a known user, or discoverable for passkeys).
- Persists the `SessionData` under a `ceremony_id` with a short TTL.
- Returns `200` with the `PublicKeyCredentialRequestOptions` for
  `navigator.credentials.get()` plus the `ceremony_id`.

#### `POST /fido/login/finish` — complete passkey login

Behavior:

- Consumes the ceremony session atomically; missing or expired is `400`.
- Looks up the stored public key by credential id and verifies the signature against
  the stored challenge.
- Validates and updates the signature counter (clone detection): a counter that goes
  backwards flags a cloned authenticator.
- On success, returns `200` with the verified user identity. Token/session issuance is
  out of scope for this study service and belongs to `app-cam` when integrated.

#### `GET /fido/credentials` — list the user's credentials

Behavior:

- Returns the credentials registered for the user (credential id, friendly name,
  transports, created/last-used timestamps, backup state). Never returns the private
  key material (there is none server-side) nor raw internal ids.

#### `DELETE /fido/credentials/{id}` — revoke a credential

Behavior:

- Soft-deletes/revokes a single credential owned by the user, for example when a device
  is lost. `404` if the credential does not exist or does not belong to the user.
  Returns `204` on success.

## Relying Party (study scope)

Fixed to localhost for local development:

- RP ID: `localhost`
- RP origin: `http://localhost:8080`

`localhost` counts as a secure context, so the browser ceremony works without HTTPS.
Native app support (iOS `AuthenticationServices`, Android Credential Manager) and
per-tenant RP configuration are deferred to the production phase.

## Roadmap

1. [x] `POST /fido/users` vertical slice with flow-id middleware, error mapping, tests, CI
2. [ ] PostgreSQL DAO for `fido_user` (replace in-memory scaffold) with Testcontainers
3. [ ] `fido_ceremony` + `fido_credential` schema and ceremony session store
4. [ ] `register/begin` + `register/finish`
5. [ ] `login/begin` + `login/finish`
6. [ ] `GET /fido/credentials` + `DELETE /fido/credentials/{id}`
7. [ ] Browser test page (`web/`) to exercise a real passkey
8. [ ] E2E tests with a virtual authenticator
9. [ ] Fold into `app-cam`: Lambda entrypoint, RDS, per-tenant RP, token integration
