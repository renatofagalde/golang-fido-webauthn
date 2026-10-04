
# golang-fido-webauthn

FIDO2 / WebAuthn (passkeys) backend in Go.

> Portuguese version below / Versao em portugues mais abaixo.

## Stack

- Go 1.25
- Gin
- PostgreSQL (local: Docker, no volume / target: AWS RDS)
- GORM (`gorm.io/driver/postgres`)
- `github.com/go-webauthn/webauthn`
- `github.com/google/uuid` (UUIDv7)
- `log/slog`
- Target runtime: AWS Lambda (`provided.al2023`, `arm64`) + API Gateway REST

## Run

```bash
docker compose up -d
go run ./cmd/api
curl -i localhost:8080/health
```

## Endpoints

Base group: `/fido`. Business routes require a valid `X-Flow-ID` header (UUID). `GET /health` is exempt.

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Liveness. Returns `200 {"status":"ok"}`. |
| POST | `/fido/users` | Create a WebAuthn user (handle for future passkeys). `409` on duplicate username. Returns `201`. |
| POST | `/fido/register/begin` | Start passkey registration for an existing user. Generates a challenge, stores the session, returns the creation options + `ceremony_id`. |
| POST | `/fido/register/finish` | Finish registration. Consumes the session (single-use, expires on read), verifies the attestation, stores the credential. `201` on success, `400` if invalid/expired. |
| POST | `/fido/login/begin` | Start passkey login (public route). Generates an assertion challenge, stores the session, returns the request options + `ceremony_id`. |
| POST | `/fido/login/finish` | Finish login. Consumes the session, verifies the signature, updates the sign counter (clone detection). `200` with the verified user, `400` if invalid/expired. |
| GET | `/fido/credentials` | List the user's registered credentials. Never returns key material. |
| DELETE | `/fido/credentials/{id}` | Revoke a credential (e.g. lost device). `204` on success, `404` if not found. |

`finish` requires a real authenticator (Touch ID, YubiKey, phone, or the browser virtual authenticator). A plain HTTP client (Postman) can call `begin` but cannot complete `finish`.

---

# golang-fido-webauthn (PT-BR)

Backend FIDO2 / WebAuthn (passkeys) em Go.

## Stack

- Go 1.25
- Gin
- PostgreSQL (local: Docker, sem volume / alvo: AWS RDS)
- GORM (`gorm.io/driver/postgres`)
- `github.com/go-webauthn/webauthn`
- `github.com/google/uuid` (UUIDv7)
- `log/slog`
- Runtime alvo: AWS Lambda (`provided.al2023`, `arm64`) + API Gateway REST

## Rodar

```bash
docker compose up -d
go run ./cmd/api
curl -i localhost:8080/health
```

## Endpoints

Grupo base: `/fido`. Rotas de negocio exigem o header `X-Flow-ID` valido (UUID). O `GET /health` e isento.

| Metodo | Path | Descricao |
| --- | --- | --- |
| GET | `/health` | Liveness. Retorna `200 {"status":"ok"}`. |
| POST | `/fido/users` | Cria um usuario WebAuthn (handle para passkeys futuras). `409` em username duplicado. Retorna `201`. |
| POST | `/fido/register/begin` | Inicia o registro de passkey para um usuario existente. Gera um challenge, guarda a sessao, retorna as opcoes de criacao + `ceremony_id`. |
| POST | `/fido/register/finish` | Finaliza o registro. Consome a sessao (uso unico, expira na leitura), verifica a attestation, guarda a credencial. `201` em sucesso, `400` se invalido/expirado. |
| POST | `/fido/login/begin` | Inicia o login por passkey (rota publica). Gera um challenge de assertion, guarda a sessao, retorna as opcoes de request + `ceremony_id`. |
| POST | `/fido/login/finish` | Finaliza o login. Consome a sessao, verifica a assinatura, atualiza o sign counter (deteccao de clone). `200` com o usuario verificado, `400` se invalido/expirado. |
| GET | `/fido/credentials` | Lista as credenciais registradas do usuario. Nunca retorna material de chave. |
| DELETE | `/fido/credentials/{id}` | Revoga uma credencial (ex.: device perdido). `204` em sucesso, `404` se nao encontrada. |

O `finish` exige um autenticador real (Touch ID, YubiKey, celular ou o autenticador virtual do browser). Um cliente HTTP puro (Postman) chama o `begin`, mas nao completa o `finish`.
