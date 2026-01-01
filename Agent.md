# Agent.md

## Project Goals

LLM Studio is a **Next.js + Golang BFF** monolith:

- **Next.js (`website/`)**: UI (based on shadcn `dashboard-01` / `login-03` templates)
- **Golang BFF (`cmd/llm-studio-bff`)**: public Web/API surface and dependency aggregation
- **Dependencies**
  - `identra`: authentication & authorization (called via gRPC in this project)
  - `go-webmods`: app bootstrap (config/logging) and shared clients (e.g. Redis)

## Code Structure (Clean Architecture)

The Golang backend follows Clean Architecture with dependencies pointing inward:

- **`internal/domain`**
  - Pure domain models (no external dependencies)
  - Currently: `Token` / `TokenPair` / `Session` / `User`
- **`internal/application`**
  - Use cases and ports
  - `internal/application/auth`
    - `OAuthGateway`: port for external OAuth/auth service (implemented by infrastructure)
    - `SessionStore`: session persistence port (implemented by infrastructure)
    - Use cases: `StartOAuth` / `CompleteOAuthToSession` / `Logout`
- **`internal/interfaces`**
  - Delivery layer (HTTP handlers / router)
  - `internal/interfaces/httpapi` depends only on the application layer (no direct identra pb usage)
- **`internal/infrastructure`**
  - Infrastructure implementations (gRPC/Redis/config, etc.)
  - `config`: config loading via `go-webmods/app.Config()` + viper
  - `identra`: identra gRPC client implementing `OAuthGateway`
  - `session`: Redis implementation of `SessionStore`
  - `persistence`: GORM (via `go-webmods/gormclient`) implementation of `UserRepository`

## Configuration Loading (go-webmods / viper)

The BFF entrypoint calls:

- `app.Init("llm-studio-bff")`

Layered config loading order (go-webmods internal logic):

- `configs/default.toml`
- `configs/llm-studio-bff/default.toml`
- `configs/{MODE}.toml` (if present)
- `configs/llm-studio-bff/{MODE}.toml` (if present)
- Environment variable overrides: `FOO__BAR` → `foo.bar`

Important note:

- `go-webmods` uses `ReadInConfig()` when loading `configs/llm-studio-bff/default.*`, which **replaces** (does not merge) the earlier `configs/default.*`.
- For this reason, treat `configs/llm-studio-bff/default.toml` as the authoritative config for the BFF, and put overrides like `auth.super_admin_emails` there.

## Auth & Session (GitHub OAuth via identra + Redis Session)

### Goal

**Do not store access/refresh tokens in the browser.** Instead:

1. identra returns `access_token / refresh_token`
2. BFF stores them in Redis (with TTL)
3. BFF only issues a single `session cookie` (HTTP-only)

Additionally:

4. BFF calls identra `GetCurrentUserLoginInfo` during login and stores the returned `LoginInfo`
5. BFF records the **identra uid** in the session payload (`session.user_id`)
6. The identra uid is also used as the **primary key** for the local `User` table (GORM)

### Session payload

Server-side session (stored in Redis) includes:

- `session.user_id`: identra uid (also the local `User.id`)
- `session.token`: the identra `TokenPair` (access/refresh)

The browser only receives `auth.session_cookie_name=sessionID` (HTTP-only).

### User model

LLM Studio maintains a minimal local user table:

- Primary key **is identra uid**
- Record is created on first successful OAuth login (`EnsureExists`)

### Related config (`configs/default.toml`)

- `identra.grpc_addr`: identra gRPC address (default `localhost:50051`)
- `identra.oauth_provider`: default `github`

- `database.driver`: default `sqlite`
- `database.name`: default `data/llm-studio.db`

- `redis.urls`: default `["localhost:6379"]`
- `redis.password`
- `redis.session_prefix`: default `llmstudio:sess:`

- `auth.session_cookie_name`: default `llmstudio_session`
- `auth.cookie_max_age_days`: session TTL (currently used as Redis TTL as well)
- `auth.super_admin_emails`: list of emails that will be granted `super_admin` role on login

### HTTP API (BFF)

- `GET /api/auth/github/login?return_to=/dashboard`
  - BFF asks identra for the OAuth URL and 302 redirects to GitHub
  - `return_to` is stored in a short-lived httpOnly cookie (for redirect back)

- `GET /api/auth/github/callback?code=...&state=...`
  - BFF calls identra `LoginByOAuth`
  - Calls identra `GetCurrentUserLoginInfo` using the access token
  - Stores the returned LoginInfo (snapshot) in the database
  - Uses `login_info.user_id` as the identra uid (local user primary key)
  - Ensures local `User` exists (primary key = identra uid)
  - If `login_info.email` matches `auth.super_admin_emails`, the user is granted `super_admin` role
  - Stores the session payload in Redis at: `redis.session_prefix + sessionID`
  - Sets `auth.session_cookie_name=sessionID`
  - 302 redirects back to `frontend.base_url + return_to` (default `/dashboard`)

- `POST /api/auth/logout`
  - Reads session cookie, deletes Redis session
  - Clears the session cookie

- `GET /api/me`
  - Returns current user info (from session + DB role)

- `GET /api/admin/users`
  - Requires the caller to be `admin` or `super_admin`
  - Returns list of users (id/role + login info snapshot)

- `POST /api/admin/users/{userID}/role`
  - Requires the caller to be `super_admin`
  - Body: `{"role":"admin"}` or `{"role":"user"}`
  - Note: `super_admin` is config-driven and cannot be granted via API

- `GET /api/health`

### Security note (uid extraction)

The current uid extraction parses JWT claims **without signature verification** (used to locate `uid` / `sub`).
If you need stronger guarantees, validate the JWT using identra JWKS before trusting the claims.

## Frontend Integration (Next.js)

### Login button

The GitHub login button in `website/components/login-form.tsx` navigates to:

- `/api/auth/github/login?return_to=/dashboard`

### Proxy `/api` to BFF

`website/next.config.ts` sets a rewrite:

- `/api/:path*` → `${BFF_BASE_URL}/api/:path*` (default `http://localhost:8080`)

## Local Development

### Dependencies

`docker-compose.yml` includes:

- `identra-grpc` (`50051:50051`)
- `redis` (`6379:6379`)

Start:

```bash
docker compose up -d
```

### Run the BFF

```bash
go run ./cmd/llm-studio-bff
```

### Run the frontend

```bash
pnpm -C website dev
```

## Debugging (VSCode)

We provide `.vscode/launch.json`:

- **Debug: llm-studio-bff**
  - Debugs `cmd/llm-studio-bff`
  - `cwd` is set to the repo root so `configs/` can be loaded

- **Debug: next dev (website)**
  - Runs `pnpm dev` under `website/`
  - Use `BFF_BASE_URL` to point to the local BFF

## Notes

- The frontend currently has existing `pnpm lint` issues in the repo (unrelated to the login/proxy changes). We can fix them separately if you want a fully green lint.

