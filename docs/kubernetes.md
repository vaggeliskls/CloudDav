# Kubernetes / Helm

The Helm chart is published to GHCR as an OCI artifact at `oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server`. It deploys a Deployment, Service, ConfigMap, Secret, optional PVC, and optional Ingress.

## Install

```sh
helm install wd oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z>
```

Find published versions on the [chart's GHCR page](https://github.com/vaggeliskls/cloud-webdav-server/pkgs/container/charts%2Fcloud-webdav-server). Always pin `--version` in production — omitting it falls back to whatever Helm resolves as latest.

By default the chart deploys with a local-filesystem backend backed by a 10 GiB PVC. Override anything via `--set` flags or a custom `values.yaml`.

## Pulling and inspecting the chart

```sh
# Show default values
helm show values oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z>

# Save a copy locally to edit
helm pull oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z> --untar
```

## Common configurations

> Each example uses `<X.Y.Z>` as a placeholder — replace with the version you want to pin. The OCI URL stays the same.

### Local filesystem (with PVC)

```sh
helm install wd oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z> \
  --set persistence.enabled=true \
  --set persistence.size=50Gi \
  --set basicAuth.users="alice:alice123 bob:bob456"
```

The data lives on a `PersistentVolumeClaim` named `<release>-data`. Use a `RWO` volume — the Deployment uses `Recreate` strategy so two replicas never share the volume.

### S3

```sh
helm install wd oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z> \
  --set storage.type=s3 \
  --set storage.s3.bucket=my-webdav-bucket \
  --set storage.s3.region=us-east-1 \
  --set storage.s3.accessKeyId=$AWS_KEY \
  --set storage.s3.secretAccessKey=$AWS_SECRET \
  --set persistence.enabled=false
```

### Google Cloud Storage

GCS uses a service-account JSON file, mounted from a Kubernetes Secret you create out-of-band:

```sh
kubectl create secret generic gcs-credentials \
  --from-file=sa.json=/path/to/service-account.json

helm install wd oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z> \
  --set storage.type=gcs \
  --set storage.gcs.bucket=my-webdav-bucket \
  --set storage.gcs.serviceAccountSecret=gcs-credentials \
  --set persistence.enabled=false
```

The chart mounts the secret at `/secrets/gcs/sa.json` and sets `GOOGLE_APPLICATION_CREDENTIALS` automatically.

### Azure Blob Storage

```sh
helm install wd oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z> \
  --set storage.type=azure \
  --set storage.azure.account=mystorageacct \
  --set storage.azure.container=webdav \
  --set storage.azure.key=$AZURE_KEY \
  --set persistence.enabled=false
```

> The `endpoint` value stays empty for public Azure. Set it only for sovereign clouds (Azure Government, Azure China) or Azure Stack.

## Authentication

### Basic auth (default)

```yaml
basicAuth:
  enabled: true
  users: "alice:alice123 bob:bob456"
```

The `users` value is stored in the chart's Secret. Rotate by updating the value and `helm upgrade`.

### LDAP

```yaml
ldap:
  enabled: true
  url: "ldaps://ldap.example.com"
  baseDN: "ou=users,dc=example,dc=com"
  bindDN: "uid=searchuser,ou=users,dc=example,dc=com"
  bindPassword: "..."   # in Secret
  attribute: "uid"      # or "sAMAccountName" for AD
```

### OAuth 2.0 / OpenID Connect (Bearer)

```yaml
oauth:
  enabled: true
  providerURL: "https://keycloak.example.com/realms/myrealm"
  clientID: "webdav-client"
  clientSecret: "..."   # in Secret
```

## Ingress

```yaml
ingress:
  enabled: true
  host: "files.example.com"
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "0"   # allow large uploads
    nginx.ingress.kubernetes.io/proxy-read-timeout: "1800"
  tls:
    enabled: true
    secretName: "files-example-com-tls"
```

Per ingress controller examples (paste into `annotations`):

| Controller | Annotation                                                                |
|------------|---------------------------------------------------------------------------|
| nginx      | `nginx.ingress.kubernetes.io/proxy-body-size: "0"` (uncapped uploads)     |
| Traefik    | `traefik.ingress.kubernetes.io/router.entrypoints: web,websecure`         |
| GKE        | `kubernetes.io/ingress.class: "gce"`                                      |

> WebDAV uses non-standard methods (`PROPFIND`, `MKCOL`, `LOCK`, `UNLOCK`, `PROPPATCH`, `COPY`, `MOVE`). Most controllers pass them through, but a few require explicit allow-lists. Check your controller's docs if `LOCK` returns 405.

## Health probes

The Deployment defines both `livenessProbe` and `readinessProbe` against `/_health`. They use `httpGet`, so no additional config is needed.

## Resource defaults

```yaml
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

The server is small (~10 MB binary) and CPU-bound only during large file transfers. Scale `limits.memory` if you expect frequent multi-GiB uploads on the in-memory buffering path (planned to switch to streaming).

## Upgrade / uninstall

```sh
helm upgrade wd oci://ghcr.io/vaggeliskls/charts/cloud-webdav-server --version <X.Y.Z> -f my-values.yaml
helm uninstall wd
```

The PVC is **not** deleted automatically — `kubectl delete pvc <release>-data` if you want to release the storage.
