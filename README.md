# Cloud Webdav Server — WebDAV Server for Cloud Storage

[![CI](https://github.com/vaggeliskls/cloud-webdav-server/actions/workflows/ci.yml/badge.svg)](https://github.com/vaggeliskls/cloud-webdav-server/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/vaggeliskls/cloud-webdav-server)](https://goreportcard.com/report/github.com/vaggeliskls/cloud-webdav-server)

A lightweight, production-ready WebDAV server written in Go.
Mount **Amazon S3**, **Google Cloud Storage**, or a **local directory** as a WebDAV drive with per-folder access control and multiple authentication methods.

> **Inspired by** [vaggeliskls/webdav-server](https://github.com/vaggeliskls/webdav-server) — if you don't need cloud storage, check out that project for a simpler Docker-based WebDAV server with Basic, LDAP, and OAuth/OIDC support.

---

## Features

| Feature | Description |
|---|---|
| ☁️ **Storage backends** | Local filesystem, Amazon S3 / MinIO, Google Cloud Storage |
| 🔒 **Path-based permissions** | Per-folder access rules with user lists, wildcards, and exclusions |
| 🔑 **Authentication** | HTTP Basic, LDAP / Active Directory, OpenID Connect (Bearer token) |
| 📖 **Read-only mode** | Lock folders to `ro` per-folder or per-user |
| 📁 **Auto-create folders** | Directories created at startup from the permission config |
| 🌐 **CORS** | Configurable cross-origin support for web clients |
| 🩺 **Health check** | Optional `/_health` endpoint for load-balancer probes |
| 🚫 **Browser block** | Prevents accidental access from browsers (optional) |
| 🐳 **Minimal Docker image** | Distroless `scratch` image, non-root user, ~10 MB |

---

## Quick Start

```sh
cp .env.example .env   # edit as needed
docker compose up
```

The server listens on `http://localhost:8080`.

Try it immediately with curl:

> **Note:** For the curl examples below to work, the `files/` folder (mapped to `LOCAL_DATA_PATH` on the host) must be readable **and** writable by the container user. Run this on the host before starting the container:

```sh
# 📤 Upload a file
curl -T README.md http://localhost:8080/files/README.md -u alice:alice123

# 📥 Download a file
curl http://localhost:8080/files/README.md -u alice:alice123 -O

# 📂 List directory (PROPFIND)
curl -X PROPFIND http://localhost:8080/files/ -u alice:alice123
```

---

## Configuration

All configuration is via environment variables (or a `.env` file).

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `STORAGE_TYPE` | `local` | `local` · `s3` · `gcs` |
| `LOCAL_DATA_PATH` | `/data` | Root dir for local storage |
| `FOLDER_PERMISSIONS` | `/files:*:rw` | Comma-separated permission rules (see below) |
| `BASIC_AUTH_ENABLED` | `true` | Enable HTTP Basic auth |
| `BASIC_USERS` | — | Space-separated `"alice:pass1 bob:pass2"` |
| `AUTO_CREATE_FOLDERS` | `true` | Create configured folders at startup |
| `BROWSER_BLOCK_ENABLED` | `false` | Return 403 to browser User-Agents |
| `CORS_ENABLED` | `false` | Enable CORS headers |
| `CORS_ALLOWED_ORIGINS` | `*` | Allowed origins |
| `SERVER_PORT` | `8080` | Listening port |

> See [.env.example](.env.example) for the full list including S3, GCS, LDAP, and OIDC options.

### Folder Permissions

```
FOLDER_PERMISSIONS=/public:public:ro,/files:*:rw,/alice:alice:rw
```

Format: `/path:users:mode`

| Field | Values |
|---|---|
| `path` | Any URL prefix (e.g. `/files`, `/team/docs`) |
| `users` | `public` (no auth) · `*` (any authenticated) · `alice bob` (specific) · `* !charlie` (exclude) |
| `mode` | `ro` (read-only) · `rw` (read-write) |

**Longest prefix wins** — `/private/secret` takes precedence over `/private`.

#### Examples

```sh
# Public read-only share + private rw for alice and bob, charlie excluded from /shared
FOLDER_PERMISSIONS=/public:public:ro,/alice:alice:rw,/bob:bob:rw,/shared:* !charlie:rw
```

### Storage Backends

**Local filesystem**
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
# For MinIO:
S3_ENDPOINT=http://localhost:9000
S3_FORCE_PATH_STYLE=true
```

**Google Cloud Storage**
```env
STORAGE_TYPE=gcs
GCS_BUCKET=my-bucket
GOOGLE_APPLICATION_CREDENTIALS=/run/secrets/sa.json
```

### Authentication

**Basic Auth** (default)
```env
BASIC_AUTH_ENABLED=true
BASIC_USERS="alice:alice123 bob:bob456"
```

**LDAP / Active Directory**
```env
LDAP_ENABLED=true
LDAP_URL=ldap://ldap.example.com:389
LDAP_BASE_DN=dc=example,dc=com
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=secret
```

**OpenID Connect (Bearer token)**
```env
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://accounts.example.com
OIDC_CLIENT_ID=my-client
```

---

## Development

**Prerequisites:** Go 1.22+, Docker (for MinIO integration tests)

```sh
# Run the server locally
make run

# Run all tests (with race detector)
make test

# Start a local MinIO instance for S3 testing
make minio-up

# Stop MinIO
make minio-down
```

**Test coverage:**

| Package | Coverage |
|---|---|
| `internal/config` | Permission config parsing |
| `internal/permissions` | Access control logic & path matching |
| `internal/server` | HTTP middleware + integration tests |

---


## References

- [vaggeliskls/webdav-server](https://github.com/vaggeliskls/webdav-server) — the original Apache httpd-based WebDAV server that inspired this project. Supports Basic, LDAP, OAuth/OIDC, and per-folder access control via Docker.
- [golang.org/x/net/webdav](https://pkg.go.dev/golang.org/x/net/webdav) — Go standard WebDAV handler
- [WebDAV RFC 4918](https://www.rfc-editor.org/rfc/rfc4918)
