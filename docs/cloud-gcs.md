# Google Cloud Storage

The GCS backend uses [`cloud.google.com/go/storage`](https://pkg.go.dev/cloud.google.com/go/storage), the official Google Cloud Storage Go client.

## 1. Create the bucket

```sh
gcloud storage buckets create gs://my-webdav-bucket \
  --location=us-central1 \
  --uniform-bucket-level-access
```

Use **uniform bucket-level access** — fine-grained ACLs are legacy and harder to reason about. The server doesn't rely on object-level ACLs anyway.

## 2. Create a service account with restricted permissions

Don't reuse a personal account or one with broad `storage.admin`. Create a dedicated service account and grant it just `storage.objectAdmin` on the bucket:

```sh
SA_NAME=cloud-webdav
PROJECT=$(gcloud config get-value project)

gcloud iam service-accounts create $SA_NAME \
  --display-name="Cloud WebDAV Server"

gcloud storage buckets add-iam-policy-binding gs://my-webdav-bucket \
  --member="serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com" \
  --role=roles/storage.objectAdmin
```

`storage.objectAdmin` covers list/read/write/delete on objects within the bucket. It does **not** allow bucket configuration changes — that's intentional.

## 3a. Local / Docker: download a JSON key

For local dev or non-GCP container hosts, download a JSON key file:

```sh
gcloud iam service-accounts keys create sa.json \
  --iam-account=${SA_NAME}@${PROJECT}.iam.gserviceaccount.com
```

Configure the server:

```env
STORAGE_TYPE=gcs
GCS_BUCKET=my-webdav-bucket
GCS_PREFIX=webdav/                              # optional
GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json
```

Mount the key into the container:

```sh
docker run -d --name webdav \
  -p 8080:8080 \
  -e STORAGE_TYPE=gcs \
  -e GCS_BUCKET=my-webdav-bucket \
  -e GOOGLE_APPLICATION_CREDENTIALS=/secrets/sa.json \
  -v /path/to/sa.json:/secrets/sa.json:ro \
  ghcr.io/vaggeliskls/cloud-webdav-server:latest
```

> JSON keys are long-lived bearer credentials. Rotate them, store them in a secret manager, never commit them.

## 3b. Kubernetes: mount the key from a Secret

```sh
kubectl create secret generic gcs-credentials \
  --from-file=sa.json=/path/to/sa.json
```

```yaml
storage:
  type: gcs
  gcs:
    bucket: "my-webdav-bucket"
    prefix: "webdav/"
    serviceAccountSecret: "gcs-credentials"
```

The chart mounts the secret at `/secrets/gcs/sa.json` and sets `GOOGLE_APPLICATION_CREDENTIALS` automatically.

## 3c. GKE Workload Identity (preferred for production on GKE)

On a Workload-Identity-enabled GKE cluster, you can skip JSON keys entirely:

```sh
# Bind the GCP service account to the K8s service account used by the pod
gcloud iam service-accounts add-iam-policy-binding \
  ${SA_NAME}@${PROJECT}.iam.gserviceaccount.com \
  --role=roles/iam.workloadIdentityUser \
  --member="serviceAccount:${PROJECT}.svc.id.goog[<namespace>/<ksa-name>]"
```

Annotate the K8s ServiceAccount and reference it from the Deployment. With Workload Identity active, leave `serviceAccountSecret` empty and don't set `GOOGLE_APPLICATION_CREDENTIALS` — the SDK uses Application Default Credentials transparently.

> The chart doesn't yet generate a ServiceAccount or apply the WI annotation. PR welcome.

## Notes on behavior

- **No native rename**: WebDAV `MOVE` runs as a server-side copy + delete via `Object.CopierFrom`.
- **Directory markers**: same pattern as S3 — empty objects ending in `/`.
- **Auth fallback**: when `GOOGLE_APPLICATION_CREDENTIALS` is empty, the SDK falls back to Application Default Credentials (metadata server on GCE/GKE, `gcloud auth application-default login` locally, etc).
