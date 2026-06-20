# fotoferry

Copy photos from a set of source folders into a destination share, **exactly once**.

fotoferry is built for moving Immich library originals onto a NAS share that
non-technical people then reorganize by hand. Once a photo has been delivered it
must never come back, even after it is moved or deleted at the destination. So
fotoferry keeps its own record of what it has delivered and **never reconciles
against the destination**.

## How it works

For each mapping (`source` -> `dest`) fotoferry walks the source tree and, per file:

1. If the relative path is already recorded with the same size and mtime, skip it
   without reading it.
2. Otherwise hash the file (sha256). If those bytes were already delivered for this
   mapping, just record the new path and move on. This is what makes an Immich
   storage-template change (which moves files to new paths) a no-op instead of a
   re-copy.
3. Otherwise copy it to `dest/<relative-path>` and record it.

Identity is the content hash, scoped per mapping, so the same photo shared by two
people is still delivered to both. The destination is never read back, so a file
the family moves off the share is never re-delivered.

It runs as a one-shot binary on a cron schedule inside the container; cron drives
the repetition.

## Configuration

All configuration is via environment variables.

| Variable        | Default            | Description |
|-----------------|--------------------|-------------|
| `CRON_SCHEDULE` | `0 * * * *`        | When to run, in cron syntax. Read by the container, not the binary. |
| `MAPPING__N__*` | none               | One indexed group per source/dest pair (see below). At least one is required. |
| `STATE_DB`      | `/data/state.db`   | SQLite file recording what has been delivered. |
| `REQUIRE_PATHS` | none               | Comma-separated paths that must exist or the run is skipped (see Safety). |
| `LOG_FORMAT`    | `text`             | `text` or `json`. |
| `DRY_RUN`       | `false`            | Log what would be copied without writing anything. |

### Mappings

Each mapping is an indexed group of three variables:

```
MAPPING__0__ID=alice
MAPPING__0__SOURCE=/library/alice
MAPPING__0__DEST=/dest/alice
MAPPING__1__ID=bob
MAPPING__1__SOURCE=/library/bob
MAPPING__1__DEST=/dest/bob
```

`ID` is the per-mapping key used in the state DB and logs. Indices are read in
order and parsing stops at the first gap.

## Running

```sh
docker compose up --build
```

`docker-compose.yml` bind-mounts the source library read-only, the destination
read-write, and a named volume for the state DB:

```yaml
volumes:
  - /path/to/library:/library:ro
  - /path/to/share:/dest
  - data:/data
```

Keep the state volume on persistent, backed-up storage. If it is lost, fotoferry
has no memory of past deliveries and will re-copy everything currently in the
source.

## Safety

The destination is often a network share that may be unmounted. A bind mount of an
unmounted share looks like an empty directory, so without a guard fotoferry would
copy onto local disk and record those files as delivered.

`REQUIRE_PATHS` prevents this: point it at a sentinel that only exists when the
share is really mounted, e.g. a marker file placed on the NAS:

```
REQUIRE_PATHS=/dest/.fotoferry-online
```

If the marker is missing the run is skipped and cron tries again later.

## Development

```sh
nix develop          # go, golangci-lint, just
just test
just build
just run --help
just docker          # docker compose up --build
```
