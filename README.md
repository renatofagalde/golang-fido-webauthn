# golang-fido-webauthn

Study backend for learning FIDO2 / WebAuthn (passkeys) in Go, built with the same
Clean Architecture skeleton used in the production `gestao.one` services. It starts
as a standalone Gin + PostgreSQL service and is designed to later fold into `app-cam`
(the auth/RBAC service) and run serverless.

> Portuguese version below / Versao em portugues mais abaixo.

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

---

# golang-fido-webauthn (PT-BR)

Backend de estudo para aprender FIDO2 / WebAuthn (passkeys) em Go, construido com o
mesmo esqueleto de Clean Architecture usado nos servicos de producao do `gestao.one`.
Comeca como um servico standalone Gin + PostgreSQL e foi desenhado para depois ser
incorporado ao `app-cam` (o servico de auth/RBAC) e rodar serverless.

## Objetivo

Aprender a cerimonia WebAuthn de ponta a ponta (registro e login), com as convencoes
reais de engenharia: Clean Architecture, interfaces consumer-driven, semantica explicita
de ponteiro versus valor, log estruturado, testes table-driven e CI. Cada endpoint e
construido como uma fatia vertical completa, do HTTP ate a persistencia.

## Stack

| Area | Atual (estudo local) | Alvo (producao) |
| --- | --- | --- |
| Linguagem | Go 1.25 | Go 1.25 |
| HTTP | Gin | Gin (via API Gateway) |
| Computacao | binario standalone | AWS Lambda (`provided.al2023`, `arm64`) + API Gateway REST |
| Banco | PostgreSQL 18 em Docker (sem volume, efemero) | AWS RDS PostgreSQL |
| ORM | GORM (`gorm.io/driver/postgres`) | igual |
| WebAuthn | `github.com/go-webauthn/webauthn` | igual |
| IDs | `github.com/google/uuid` (UUIDv7) | igual |
| Log | `log/slog` (JSON) | igual, enviado ao CloudWatch |
| Testes | `testing` + `testify`, Testcontainers na integracao | igual |
| CI | GitHub Actions (build, vet, test -race) | igual |

O servico e escrito com pouca dependencia de framework nas bordas, entao a migracao
para Lambda e uma troca de entrypoint (`cmd/lambda/main.go`) e do store da cerimonia,
nao um redesenho.

## Arquitetura

Clean Architecture com as setas de dependencia apontando para dentro. O dominio nao
depende de nada; a aplicacao depende do dominio; a infraestrutura depende do dominio.
Nenhuma tag de framework (gorm, json) vaza para as entidades de dominio.

```
cmd/api/                         entrypoint do processo, wiring de dependencia
internal/
  fido/
    domain/                      entidades puras + interfaces de repositorio + sentinel errors
    application/                 usecases (interface + impl), structs de input
    infrastructure/
      http/                      handlers, DTOs, mappers DTO<->dominio, mapeamento de erro
      repository/                models GORM, mappers model<->dominio, DAOs
      webauthn/                  provider WebAuthn (config de RP) e cola da cerimonia
  infrastructure/
    http/                        router + middleware compartilhados (flow-id, log)
pkg/
  flowid/                        id de correlacao da request no context (neutro de framework)
  writeerror/                    writer generico de resposta de erro HTTP
  ptr/                           helper generico Ptr[T]
sql/init/                        arquivos de schema, aplicados na criacao do container
web/                             pagina de teste no browser (fase posterior)
```

### Comportamento transversal

- Toda request de negocio precisa carregar um header `X-Flow-ID` valido (UUID). O
  middleware rejeita valor ausente ou malformado com `400`, propaga o id pelo
  `context.Context` para toda camada logar com o mesmo id de correlacao, e devolve ele
  na resposta. O `GET /health` e isento, para sondas de infraestrutura funcionarem sem
  o header.
- Cada request loga `request received` e `request completed` (metodo, path, status,
  latencia) como JSON estruturado, vinculado ao flow id.
- Erros de dominio sao mapeados para status HTTP em um unico lugar por modulo
  (`infrastructure/http/errors.go`), entao o dominio nunca importa HTTP.

## Rodando localmente

```bash
# sobe o Postgres efemero (o schema em sql/init e aplicado na criacao do container)
docker compose up -d

# roda a API
go run ./cmd/api

# health check
curl -i localhost:8080/health
```

Gates de qualidade (os mesmos da CI):

```bash
go build ./...
go vet ./...
go test -race ./...
```

## Endpoints

Grupo base: `/fido`. Toda rota de negocio exige o header `X-Flow-ID` valido.

### Implementado

#### `POST /fido/users` — cria um usuario WebAuthn

Cria o registro de usuario ao qual uma passkey sera vinculada depois. Uma credencial
WebAuthn precisa de um handle estavel, opaco e sem PII, entao o usuario e criado
primeiro e a cerimonia o referencia em seguida.

Request:

```json
{ "username": "padme", "display_name": "Padme Amidala" }
```

Comportamento:

- Valida que ambos os campos estao presentes (`400` em body ausente/malformado).
- Rejeita username duplicado com `409`.
- Gera um `hash` UUIDv7 que tambem serve de handle WebAuthn (sem PII).
- Retorna `201` com o usuario criado (hash, username, display_name, timestamps). O id
  numerico interno e os campos de auditoria nao sao expostos.

#### `GET /health` — liveness

Retorna `200 {"status":"ok"}`. Isento do header de flow-id.

### Planejado — cerimonia WebAuthn

WebAuthn e uma cerimonia de dois passos. O passo `begin` emite um challenge
criptografico que o passo `finish` precisa validar, entao o challenge (`SessionData`) e
persistido entre as duas chamadas (na `fido_ceremony`), de uso unico, e expirado na
leitura, nunca confiado apos o prazo.

Um autenticador real e necessario para completar o `finish` (Touch ID, YubiKey, celular
ou o autenticador virtual do browser). Um cliente HTTP puro como o Postman consegue
chamar o `begin`, mas nao completa o `finish`, porque nao produz a assinatura.

#### `POST /fido/register/begin` — inicia o registro de passkey

Para um usuario ja existente/autenticado que esta adicionando uma passkey a conta.

Comportamento:

- Carrega o usuario pelo hash; `404` se desconhecido.
- Monta as opcoes WebAuthn escopadas pelo RP e gera um challenge.
- Persiste o `SessionData` vinculado a um `ceremony_id` com TTL curto.
- Retorna `200` com o `PublicKeyCredentialCreationOptions` (JSON que o browser passa ao
  `navigator.credentials.create()`) e o `ceremony_id` que o cliente devolve no finish.

#### `POST /fido/register/finish` — completa o registro de passkey

Comportamento:

- Consome a sessao da cerimonia atomicamente; se ausente ou expirada, `400` (o cliente
  reinicia do `begin`). O challenge e de uso unico.
- Verifica a attestation produzida pelo autenticador contra o challenge armazenado.
- Em sucesso, armazena a nova credencial (credential id, chave publica, sign count,
  AAGUID, transports, flags de backup) vinculada ao usuario, e retorna `201`.
- Em falha de verificacao, `400`. O challenge ja foi queimado, entao um retry exige um
  novo `begin`.

#### `POST /fido/login/begin` — inicia o login por passkey

Rota publica (e assim que o usuario se autentica, entao nao ha sessao previa). Em
producao fica fora do grupo autenticado, ao lado da emissao de token.

Comportamento:

- Gera um challenge de assertion (para um usuario conhecido, ou discoverable para
  passkeys).
- Persiste o `SessionData` sob um `ceremony_id` com TTL curto.
- Retorna `200` com o `PublicKeyCredentialRequestOptions` para o
  `navigator.credentials.get()` e o `ceremony_id`.

#### `POST /fido/login/finish` — completa o login por passkey

Comportamento:

- Consome a sessao da cerimonia atomicamente; ausente ou expirada e `400`.
- Busca a chave publica armazenada pelo credential id e verifica a assinatura contra o
  challenge armazenado.
- Valida e atualiza o contador de assinatura (deteccao de clone): um contador que anda
  para tras sinaliza um autenticador clonado.
- Em sucesso, retorna `200` com a identidade do usuario verificada. Emissao de
  token/sessao esta fora do escopo deste servico de estudo e pertence ao `app-cam` na
  integracao.

#### `GET /fido/credentials` — lista as credenciais do usuario

Comportamento:

- Retorna as credenciais registradas para o usuario (credential id, nome amigavel,
  transports, timestamps de criacao/ultimo uso, estado de backup). Nunca retorna o
  material da chave privada (nao existe no servidor) nem ids internos crus.

#### `DELETE /fido/credentials/{id}` — revoga uma credencial

Comportamento:

- Faz soft-delete/revoga uma unica credencial do usuario, por exemplo quando um device
  e perdido. `404` se a credencial nao existe ou nao pertence ao usuario. Retorna `204`
  em sucesso.

## Relying Party (escopo de estudo)

Fixo em localhost para desenvolvimento local:

- RP ID: `localhost`
- RP origin: `http://localhost:8080`

`localhost` conta como secure context, entao a cerimonia no browser funciona sem HTTPS.
Suporte a app nativo (iOS `AuthenticationServices`, Android Credential Manager) e
configuracao de RP por tenant ficam para a fase de producao.

## Roadmap

1. [x] Fatia vertical do `POST /fido/users` com middleware de flow-id, mapeamento de erro, testes e CI
2. [ ] DAO PostgreSQL do `fido_user` (troca o andaime em memoria) com Testcontainers
3. [ ] Schema `fido_ceremony` + `fido_credential` e store da sessao da cerimonia
4. [ ] `register/begin` + `register/finish`
5. [ ] `login/begin` + `login/finish`
6. [ ] `GET /fido/credentials` + `DELETE /fido/credentials/{id}`
7. [ ] Pagina de teste no browser (`web/`) para exercitar uma passkey real
8. [ ] Testes E2E com autenticador virtual
9. [ ] Incorporar ao `app-cam`: entrypoint Lambda, RDS, RP por tenant, integracao de token
