# CloudDav — WebDAV Server for Cloud Storage

A lightweight WebDAV server written in Go. Mount cloud storage (GCS, S3) or a local directory as a WebDAV drive.

## Features

- **Storage backends** — Local filesystem, Amazon S3 / MinIO, Google Cloud Storage
- **Path-based permissions** — Per-folder access rules with user lists, wildcards, and exclusions
- **Authentication** — HTTP Basic, LDAP / Active Directory, OpenID Connect (Bearer token)
- **Read-only mode** — Lock folders to `ro` per-folder or per-user
- **Auto-create folders** — Directories are created at startup from the permission config
- **CORS, health check, browser block** — Optional HTTP middleware

## Quick Start

```sh
cp .env.example .env   # edit as needed
docker compose up
```

The server listens on `http://localhost:8080`.

## Configuration

All configuration is via environment variables (or `.env`).

| Variable | Default | Description |
|---|---|---|
| `STORAGE_TYPE` | `local` | `local` · `s3` · `gcs` |
| `LOCAL_DATA_PATH` | `./webdav-data` | Root dir for local storage |
| `FOLDER_PERMISSIONS` | `/files:*:rw` | Comma-separated permission rules |
| `BASIC_AUTH_ENABLED` | `true` | Enable HTTP Basic auth |
| `BASIC_USERS` | — | `"alice:pass1 bob:pass2"` |
| `AUTO_CREATE_FOLDERS` | `true` | Create folders at startup |
| `HEALTH_CHECK_ENABLED` | `false` | Expose `/_health` endpoint |
| `SERVER_PORT` | `8080` | Listening port |

See `.env` for the full list including S3, GCS, LDAP, and OIDC options.

### Folder Permissions

```
FOLDER_PERMISSIONS=/public:public:ro,/files:*:rw,/alice:alice:rw
```

Format: `/path:users:mode`

| Field | Values |
|---|---|
| `users` | `public` (no auth) · `*` (any authenticated) · `alice bob` (specific users) · `* !charlie` (exclude) |
| `mode` | `ro` (read-only) · `rw` (read-write) |

Longest prefix wins, so `/private` takes precedence over `/`.

### Storage Backends

**Local**
```env
STORAGE_TYPE=local
LOCAL_DATA_PATH=/data
```

**Amazon S3 / MinIO**
```env
STORAGE_TYPE=s3
S3_BUCKET=my-bucket
S3_REGION=us-east-1
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
```

**Google Cloud Storage**
```env
STORAGE_TYPE=gcs
GCS_BUCKET=my-bucket
GOOGLE_APPLICATION_CREDENTIALS=/run/secrets/sa.json
```

## Mounting

**macOS / Linux (Finder / davfs2)**
```
http://localhost:8080/files/
```

**curl**
```sh
# Upload
curl -T myfile.txt http://localhost:8080/files/myfile.txt -u alice:alice123
# Download
curl http://localhost:8080/files/myfile.txt -u alice:alice123 -O
```

## Development

```sh
go build ./...
go test ./...
```
