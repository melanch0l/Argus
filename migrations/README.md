# PostgreSQL migrations — Phase 7 boundary

This directory reserves the backend migration boundary; there are no SQL migrations or application tables yet. PostgreSQL 17 development deployment lives in `deploy/`. No SQLite database exists to migrate; this change revises an unimplemented design, not stored event data.

Phase 7 will introduce a pinned pgx v5 driver, checksum-tracked migrations, dedicated migration/application roles and integration tests. Follow `docs/architecture/database-design.md`: native UUIDs, checked unsigned NUMERIC(20,0), exact immutable canonical JSON separate from JSONB, microsecond timestamp projections plus exact nanosecond remainders, per-host locks, unique/FK constraints, bounded transactions, evidence retention and backup/restore. Do not add placeholder repositories or claim database connectivity from the foundation config check.
