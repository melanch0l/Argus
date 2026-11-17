# Phase 2 configuration foundation

These JSON examples contain no credentials and do not start services. The Go
language minimum is 1.25.0; the project toolchain is pinned to 1.27.3. Only the
standard library is used by the configuration and command foundations. No environment overrides or DSNs are
supported. Agent/backend configuration schemas are separate; `version: 1` is
required. Files are limited to 65,536 bytes, UTF-8, one JSON object. Unknown or
case-aliased fields, duplicate keys (including escaped equivalents), nulls,
incorrect types, trailing values, and unsupported versions fail with fixed
errors that never include file paths, keys, values, or decoder error details.

```sh
go run ./cmd/agent --version
go run ./cmd/backend --version
go run ./cmd/agent --check-config --config configs/agent.example.json
go run ./cmd/backend --check-config --config configs/backend.example.json
```

Exit codes: 0 for help, version, or valid settings; 2 for invalid flags/config;
1 for default `not_implemented` startup. `go run` itself wraps nonzero program
exit codes. No arguments starts **nothing** and reports `not_implemented`, not
readiness. `--config` requires `--check-config`; `--version` cannot be combined
with config/check. Checks open only the selected config, never token/password
files, sockets, state directories, database, or network. They verify no actual
credentials, role privileges, certificates, environment, or connectivity.
Default startup creates only a stderr component logger, not a long-running
service or monitoring goroutine. The Go runtime may have its own goroutines.

## Settings and provisional bounds

- Logging: `level` debug/info/warn/error (default info); `format` json/text
  (default json). No configuration or credentials are logged. Sensitive-key
  filtering is defense in depth, not a guarantee for free-form messages.
- Agent `backend_url`: HTTPS; HTTP is allowed only with literal loopback IPs,
  never `localhost` or DNS names. Exact base path `/api/v1`, no userinfo, query,
  fragment, or encoded path aliases. Explicit ports: 1..65535.
- `token_file`, agent `state_dir` and `docker_socket`, PostgreSQL `password_file`:
  clean absolute Unix paths; no URI syntax, root-only path, traversal, NUL,
  backslash, or control characters. No file permissions/existence checks yet.
- Backend `listen`: literal loopback IP and port 1..65535; default
  `127.0.0.1:8080`. Public-facing TLS/auth/proxy deployment is deferred.
- PostgreSQL host: simple DNS hostname, literal IP, or clean absolute Unix socket
  **directory**. Port defaults to 5432, range 1..65535. Database and application
  user: 1..63 ASCII letters/digits/underscore, not starting with a digit.
  Known bootstrap/admin role names are rejected; this is not a live privilege
  audit and cannot prove that another role is least privileged.
- PostgreSQL TLS: default and remote requirement `verify-full`. Local
  `sslmode: "disable"` requires **both** `allow_local_insecure: true` and a literal
  loopback IP or Unix socket directory. DNS `localhost` is not an exception.
  Unix sockets require this explicit local setting because TLS peer verification
  is not available there. Local insecure mode is an operator trust decision,
  not a production remote default. No certificates/passwords are read yet.
- Pool max defaults to 16, bounded 1..`total_connection_budget`; total budget
  defaults to 30, bounded 1..30. This is a configuration ceiling, not actual
  cross-process admission control. Deployment must budget all app instances and
  separate migration/admin connections against server capacity in Phase 7.
- Request timeout: default 30 seconds, 1..120. PostgreSQL connection timeout:
  default 5, 1..30; statement timeout: default 30, 1..120. All are integer seconds.
  These are validated provisional settings, not timers enforced by a service.

Missing deployment-specific fields are invalid: agent URL/token/state/socket;
backend token and PostgreSQL host/database/user/password file. Defaults apply
only to omitted optional settings, never explicit null or zero. PostgreSQL 17
server settings, actual pgx pool, migrations, bounded transactions, auth and
runtime compatibility remain later-phase work. No synthetic events or event
schema are introduced here.
