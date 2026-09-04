# Database development

Storage uses Bun query builders with SQLite, PostgreSQL 16+, and MySQL 8.4+.
SQLite remains the default and uses the pure Go modernc driver. The server
backends use Bun pgdriver and go-sql-driver/mysql. Domain structs remain
independent of Bun models and relationship loading.

## Configuration

`db_driver` / `--db-driver` / `MCP_RESEARCH_DB_DRIVER` selects `sqlite`,
`postgres`, or `mysql`. PostgreSQL aliases `pg`, `pgsql`, and `postgresql` are
accepted. Network databases require `db_dsn` / `--db-dsn` /
`MCP_RESEARCH_DB_DSN`. This is a PostgreSQL URL or a Go MySQL driver DSN.
Use TLS settings appropriate to the deployment. Keep credentials in environment
configuration; do not commit them.

For SQLite, `db_dsn` takes precedence over the existing `db` / `--db` /
`MCP_RESEARCH_DB` path. Empty configuration creates an in-memory database.
File databases use WAL and foreign keys. The health endpoint reflects the
effective driver/DSN rather than assuming an empty legacy path means in-memory.

## Queries and transactions

Use `NewSelect`, `NewInsert`, `NewUpdate`, and `NewDelete` in repositories.
Insert maps name columns alongside values; filters bind arguments at the point
where the condition is added. Bun handles quoting and value formatting; it may
inline escaped values rather than send server-side placeholders. Never assemble
SQL from request input or mark request values as `bun.Safe`.

Methods that participate in an existing transaction must use its `Querier`.
Opening a query on the parent DB during that transaction can escape rollback or
deadlock the single-connection SQLite backend. `selectRow` retains existing
positional scanners and the `sql.ErrNoRows` behavior.
Multi-query document snapshots explicitly use `REPEATABLE READ`: PostgreSQL's
default `READ COMMITTED` does not retain a snapshot between statements.

JSON tag expansion, conflict handling and the monotonic read checkpoint live
in `dialect.go`. MySQL uses duplicate-key updates rather than `INSERT IGNORE`,
so invalid foreign keys and other data errors remain errors. `ClientFoundRows`
keeps unchanged updates distinguishable from missing rows.

Short codes use a persistent counter locked inside a transaction. Counters seed
from existing codes and reserve unique numbers for concurrent writers. A failed
create can leave a gap; deleted codes are not recycled automatically.

## Schema upgrades

SQLite keeps migrations `001` through `027` and upgrades with `028`. Its data is
not rebuilt or exported during this change. PostgreSQL and MySQL have fresh
baselines matching that schema and their own counter migration. Switching the
configured driver does not copy existing data between databases.

Add a migration for each supported database whenever the schema changes.
PostgreSQL and SQLite execute each migration and its history record in one
transaction. Server databases serialize startup migrations with advisory locks.
MySQL DDL commits implicitly: `schema_migration_dirty` records an unfinished
migration and blocks subsequent startup. Restore a backup or explicitly repair
and verify the partial schema before resolving that marker/history. Do not
blindly clear the marker and replay non-idempotent DDL.

Text timestamps stay in the existing UTC format. MySQL uses case-sensitive
`utf8mb4_bin` storage and bounded string columns for indexed identifiers.
Case-insensitive searches explicitly apply `LOWER`. JSON values remain text,
preserving existing serializers and upgrade compatibility.

### Production rollout from SQLite

This release does **not** require moving production to PostgreSQL or MySQL.
Keep the existing SQLite path, volume, credentials and configuration. For a
SQLite deployment, leave `db_driver` unset or set it to `sqlite`; do not set a
new `db_dsn` accidentally, since it overrides the existing path. An empty
effective path starts an in-memory database, not your existing deployment.

The upgrade from schema `027` is additive: `028` creates `storage_counters`.
Historical migrations `001`–`027` are unchanged. A file-based regression test
compares all old table contents and definitions before and after two startups,
checks SQLite integrity and foreign keys, exercises legacy short-code numbering,
and verifies the old SQL interface still reads/writes the upgraded schema.
This does not replace a rehearsal with the actual production backup, especially
for older releases or manually modified schemas.

1. Record the running image digest/binary version and effective DB path. Keep
   the current config and secrets for rollback. Check disk space and permissions.
2. Stop writes and the application. Back up the **whole database volume** before
   replacing the binary/image, including any remaining `research.db-wal` and
   `research.db-shm` files. Do not copy just the main DB file while WAL writes
   are active. Alternatively use SQLite's consistent online backup API/tool.
   Retain a separate, untouched backup and verify it can be restored.
3. Rehearse on an isolated restored copy with no production traffic: run SQLite
   `PRAGMA integrity_check` and `PRAGMA foreign_key_check`, start the candidate,
   then repeat the checks. Verify login/API keys, team visibility, existing
   documents/revisions/read checkpoints, custom skills and export. Create and
   update a test document; confirm revision history and short-code allocation.
4. Deploy the pinned candidate with the **same volume and DB path**. Run one
   SQLite application instance during the upgrade, not old and new writers
   together. Inspect migration/startup logs and `/api/health`; confirm persistent
   mode and that the expected data is present before reopening writes.
5. If validation fails, stop the candidate, preserve the failed volume for
   diagnosis, and restore the complete pre-upgrade backup with the previous
   binary/image and config. Never mix a restored main file with WAL files from
   the failed run. Restoring a backup discards writes made after that backup;
   preserve/export those separately before recovery if writes were reopened.

Creating/merging a PR is not a production deployment. Do not deploy until the
database CI matrix and the production-copy rehearsal have passed.

## Test matrix

The same repository, service, and API tests run against every backend. They
include ownership/share access, documents and revisions, transactions, cascading
deletes, tags, migrations, concurrent code allocation and parameter roundtrips.
The SQLite historical migration tests continue to run in every matrix job.

```bash
go test -race ./cmd/... ./internal/... -count=1

TEST_DATABASE_DRIVER=postgres \
TEST_DATABASE_DSN='postgres://research:research-test@127.0.0.1:5432/research_test?sslmode=disable' \
go test -race ./cmd/... ./internal/... -count=1 -timeout 20m

TEST_DATABASE_DRIVER=mysql \
TEST_DATABASE_DSN='root:research-test@tcp(127.0.0.1:3306)/research_test' \
go test -race ./cmd/... ./internal/... -count=1 -timeout 20m
```

Use dedicated test servers. Credentials need CREATE/DROP DATABASE permission:
each fixture creates a random `research_test_<uuid>` database and removes it on
cleanup. The database named in the DSN is used only as an admin connection and
is never migrated or cleared by the fixture. Tests fail if a selected backend is
unavailable; they do not silently fall back to SQLite. CI runs all three via
`.github/workflows/test.yml`, with temporary database files in memory.

### Coverage

CI uploads `coverage-sqlite`, `coverage-postgres`, and `coverage-mysql` artifacts,
each with a Go profile, per-function summary and an HTML source view. Collect
coverage from service and API tests too: they exercise repository queries that
repository-only tests may not reach.

```bash
go test ./cmd/... ./internal/... -count=1 \
  -coverpkg=./internal/storage,./internal/service,./internal/api/... \
  -coverprofile=/tmp/research-coverage.out
go tool cover -func=/tmp/research-coverage.out
go tool cover -html=/tmp/research-coverage.out -o /tmp/research-coverage.html
```

Go coverage measures executed statements, not SQL correctness or all argument
combinations. A green query line proves that execution reached the call, not
that its filters, joins or rollback behavior are correct. Keep behavioral
assertions on real databases for those contracts. In particular, an interleaved
writer test verifies that document/provenance reads share one snapshot.

## Future vector search

Vector search is a separate follow-up. Bun does not make vector indexes portable.
It should use an explicit store contract carrying research/access filters and
embedding model/version, with implementations selected by database capability.
This change does not install vector extensions, generate embeddings or promise
vector search on MySQL Community. Test any SQLite extension against the current
pure Go driver before choosing it.
