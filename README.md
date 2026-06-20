# limeleaf

Embedded Postgres with pgvector for Go. No Docker, no external database.

## Why

Existing options for embedding Postgres in Go ([fergusstrange/embedded-postgres](https://github.com/fergusstrange/embedded-postgres), [zonkyio/embedded-postgres-binaries](https://github.com/zonkyio/embedded-postgres-binaries)) ship vanilla Postgres binaries with no extension support. Neither supports pgvector, and neither plans to.

Limeleaf compiles Postgres from source with extensions included. pgvector ships by default. Adding another extension means adding an entry to `extensions.yaml` and rebuilding.

## Usage

```go
ctx := context.Background()

pg, err := limeleaf.Start(ctx, limeleaf.Config{
    DataDir: "/tmp/myapp/data",
})
if err != nil {
    log.Fatal(err)
}
defer pg.Stop()

db, err := sql.Open("postgres", pg.ConnectionString())
```

On first run, `Start` downloads the Postgres binary for your platform, runs `initdb`, creates the default database, and enables pgvector. Subsequent calls start the existing cluster.

## Config

| Field | Default | Description |
|---|---|---|
| `DataDir` | (required) | Postgres data directory |
| `Port` | `0` (random) | Port to listen on |
| `Database` | `"limeleaf"` | Database created on first init |
| `Username` | `"limeleaf"` | Superuser name |
| `Settings` | `nil` | Additional postgresql.conf parameters |
| `CacheDir` | `~/.limeleaf/cache/` | Where downloaded binaries are stored |
| `Logger` | discard | Logger for lifecycle events |

## Pre-downloading binaries

To download the Postgres binary separately from starting it (e.g., during install rather than first request):

```go
binDir, err := limeleaf.EnsureBinary(ctx, "")
```

## Platforms

- linux/amd64
- linux/arm64
- darwin/arm64
- darwin/amd64

## Versions

Postgres 16.4, pgvector 0.7.0.

## Building tarballs

```bash
bash build/build.sh
```

Produces a stripped tarball in `dist/` for the current platform. CI builds all four platforms on tag push.

## License

AGPL-3.0
