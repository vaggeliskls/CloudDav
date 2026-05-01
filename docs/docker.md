# Docker

The published image is `ghcr.io/vaggeliskls/cloud-webdav-server:latest` — distroless `scratch`, non-root, ~10 MB. The default container reads configuration from environment variables.

## Quick start (docker compose)

The repository ships a [docker-compose.yml](https://github.com/vaggeliskls/cloud-webdav-server/blob/main/docker-compose.yml) that runs the published image with a local-filesystem backend on `./webdav-data`.

```sh
cp .env.example .env
docker compose up -d
```

The server is on `http://localhost:8080`. Stop with `docker compose down`.

## Plain docker run

Local filesystem:

```sh
docker run -d --name webdav \
  -p 8080:8080 \
  -e STORAGE_TYPE=local \
  -e LOCAL_DATA_PATH=/data \
  -e BASIC_AUTH_ENABLED=true \
  -e BASIC_USERS="alice:alice123" \
  -e FOLDER_PERMISSIONS="/files:*:rw" \
  -v $(pwd)/webdav-data:/data \
  ghcr.io/vaggeliskls/cloud-webdav-server:latest
```

S3 backend:

```sh
docker run -d --name webdav \
  -p 8080:8080 \
  -e STORAGE_TYPE=s3 \
  -e S3_BUCKET=my-webdav-bucket \
  -e S3_REGION=us-east-1 \
  -e AWS_ACCESS_KEY_ID=... \
  -e AWS_SECRET_ACCESS_KEY=... \
  -e BASIC_USERS="alice:alice123" \
  ghcr.io/vaggeliskls/cloud-webdav-server:latest
```

For GCS and Azure, see the per-cloud pages — they cover credential mounting in detail.

## Healthcheck

The image bundles a `--healthcheck` subcommand that hits `/_health` and exits 0/1. Use it in a `HEALTHCHECK` directive or compose stanza:

```yaml
healthcheck:
  test: ["CMD", "/webdav-server", "--healthcheck"]
  interval: 30s
  timeout: 5s
  retries: 3
  start_period: 5s
```

## Building from source

```sh
docker build -t cloud-webdav-server:dev .
```

The Dockerfile is a multi-stage build: a Go builder layer, then a `scratch`-based final image with the static binary, CA bundle, and a non-root user.

## Image tags

| Tag        | Meaning                                     |
|------------|---------------------------------------------|
| `latest`   | Most recent main-branch build               |
| `vX.Y.Z`   | Tagged release                              |
| `sha-...`  | Commit-pinned image (immutable)             |

Pin to a SHA or version tag in production — never `latest`.

## Logs

The server emits structured JSON to stdout. Set `LOG_LEVEL=debug` for verbose output (default `info`). Pipe through `jq` or your log aggregator of choice — every request is logged with method, path, status, duration, and authenticated user.

```json
{"time":"2026-05-01T18:46:54Z","level":"INFO","msg":"request","method":"PUT","path":"/files/README.md","status":201,"duration":"95ms","user":"alice"}
```
