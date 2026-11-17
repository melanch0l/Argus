# Frontend Phase 2 foundation

Static React/TypeScript/Vite scaffold only. The page explicitly says monitoring
is not implemented. No dashboard, API requests, simulated monitoring records,
proxy, backend client, credentials or kernel functionality are included.
Phase 8 owns the actual dashboard and service integration; see
[`src/services/README.md`](src/services/README.md).

## Reproduce

Use Node **22.18.0**, npm **10.9.3** (`.nvmrc`, `packageManager` and engine checks).
Run from `web/`:

```sh
npm ci
npm run build
npm test
npm audit
```

`npm test` first builds production output, compiles tests with TypeScript to
`.test-dist/`, then uses Node's built-in test runner. Tests cover static React
rendering, the pure API path helper (including unsafe segments and exact uint64
strings), and emitted HTML/JavaScript/CSS build assets. They are not browser
end-to-end, API, monitoring, accessibility-audit or Linux compatibility tests.

For optional interactive local work: `npm run dev` or `npm run preview` binds to
`127.0.0.1`. These commands remain running; neither is a production server.
Vite's built-in TSX handling is sufficient; no React plugin or test framework is
needed for this static foundation. Fast Refresh is not configured.

## Local validation record — 2026-11-02

On macOS arm64, Node 22.18.0 / npm 10.9.3:

- `npm install --package-lock-only --ignore-scripts`: lock generated successfully.
- `npm ci && npm test && npm audit --audit-level=low`: passed; production build
  and typecheck succeeded, all **5 tests passed**, **0 audit vulnerabilities**.
- `npm ls --depth=0`: all exact direct versions below resolved successfully.
- `npm audit --json`: zero vulnerabilities across all severities, including dev
  dependencies. This is the result at execution time only.

The first smoke-test run exposed an incorrect compiled-test relative path to
`dist`; it was corrected before the successful clean-install rerun above.
No browser end-to-end or target Linux validation was performed.

## Dependency manifest and update policy

Exact direct versions verified against npm registry metadata on **2026-11-02**:

| Package | Version | Purpose |
|---|---|---|
| `react`, `react-dom` | 19.3.1 | Static React rendering |
| `vite` | 8.3.5 | Development tooling and production bundling |
| `typescript` | 7.0.2 | Application/test type checking and test compilation |
| `@types/react`, `@types/react-dom` | 19.3.0 | Matching React typings |
| `@types/node` | 22.20.5 | Node 22 built-in test/filesystem typings |

Vite declares Node `^20.19.0 || >=22.12.0`; Node 22.18.0 meets this requirement.
React DOM declares React `^19.3.1`, and its typings require React typings
`^19.3.0`. TypeScript declares Node `>=16.20.0`. Full installation, typecheck,
build and tests are the local compatibility evidence, not metadata alone.

`package-lock.json` pins the resolved dependency graph and integrity hashes;
commit it with `package.json`. Use `npm ci`, not unreviewed dependency upgrades.
Before updating pins, inspect npm metadata/engines/peer dependencies and security
advisories, regenerate the lock with the documented npm version, then rerun clean
install, build, tests and `npm audit`. Avoid blind `npm audit fix --force`.
Audit only covers advisories known to npm at execution time, not a security
certification. Node 22.18.0 is the requested reproducibility baseline, not a claim
that it remains the latest secure Node patch; review a patched Node 22 baseline
and matching CI pin as a separate maintenance change.

CI definitions and action provenance are documented in
[`../.github/workflows/README.md`](../.github/workflows/README.md). CI performs no
deployment and cannot establish compatibility with the monitored Ubuntu kernel.
