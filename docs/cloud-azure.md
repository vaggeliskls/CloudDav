# Azure Blob Storage

The Azure backend uses [`github.com/Azure/azure-sdk-for-go/sdk/storage/azblob`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob), the official Azure Blob Storage Go SDK.

## Azure terminology

If you're coming from S3 or GCS:

| S3 / GCS    | Azure                                     |
|-------------|-------------------------------------------|
| Bucket      | **Container** (lives inside an *Account*) |
| Object key  | Blob name                                 |
| Region      | Set on the storage account at create time |

URLs look like `https://<account>.blob.core.windows.net/<container>/<blob>`.

## 1. Create a storage account

```sh
RG=my-webdav-rg
ACCT=mywebdavacct

az group create --name $RG --location eastus

az storage account create \
  --name $ACCT \
  --resource-group $RG \
  --location eastus \
  --sku Standard_LRS \
  --kind StorageV2 \
  --min-tls-version TLS1_2 \
  --allow-blob-public-access false
```

`Standard_LRS` = locally redundant. Use `Standard_GRS` if you want geo-redundancy (more expensive). `--allow-blob-public-access false` prevents accidental public buckets.

## 2. Create a container

```sh
KEY=$(az storage account keys list -g $RG -n $ACCT --query '[0].value' -o tsv)

az storage container create \
  --account-name $ACCT \
  --account-key "$KEY" \
  --name webdav
```

The container is the equivalent of an S3 bucket — it's where blobs live.

## 3. Configure the server

Shared Key auth (account name + access key) is the simplest production pattern:

```env
STORAGE_TYPE=azure
AZURE_CONTAINER=webdav
AZURE_PREFIX=webdav/                # optional
AZURE_STORAGE_ACCOUNT=mywebdavacct
AZURE_STORAGE_KEY=<account key>     # az storage account keys list ...
```

Where to find these in the Azure Portal:

1. **Account name** — top of the *Storage Account* resource page.
2. **Container** — *Data storage* → *Containers*.
3. **Access key** — *Security + networking* → *Access keys*. Two keys (`key1`, `key2`); use either, rotate periodically.

## 4. Kubernetes / Helm

```yaml
storage:
  type: azure
  azure:
    container: "webdav"
    prefix: "webdav/"
    account: "mywebdavacct"
    key: "<account key>"      # in the chart's Secret
```

Or via `--set`:

```sh
helm install wd ./kubernetes \
  --set storage.type=azure \
  --set storage.azure.account=mywebdavacct \
  --set storage.azure.container=webdav \
  --set storage.azure.key="$AZURE_KEY" \
  --set persistence.enabled=false
```

## Sovereign clouds and Azure Stack

Set `AZURE_STORAGE_ENDPOINT` (or `storage.azure.endpoint` in the chart) when the default `*.blob.core.windows.net` is wrong:

| Cloud                  | Endpoint                                          |
|------------------------|---------------------------------------------------|
| Azure Government       | `https://<account>.blob.core.usgovcloudapi.net/`  |
| Azure China (21Vianet) | `https://<account>.blob.core.chinacloudapi.cn/`   |
| Azure Stack            | `https://<account>.blob.<your-stack-fqdn>/`       |

Public Azure: leave it empty.

## AKS Workload Identity (future)

Microsoft's recommended production pattern is **Workload Identity** — pods authenticate as a Microsoft Entra ID managed identity, no shared keys involved. Implementing this in this server requires switching from `azblob.NewClientWithSharedKeyCredential` to `azblob.NewClient` with `azidentity.NewDefaultAzureCredential()`.

Currently the server only supports Shared Key. Open an issue or PR if you need Workload Identity / DefaultAzureCredential support.

## Notes on behavior

- **No native rename**: WebDAV `MOVE` runs as a same-account `StartCopyFromURL` + `DeleteBlob`. Same-account copies are sync (instant for blobs in the same container under Shared Key auth).
- **Directory markers**: empty blobs with names ending in `/`.
- **Authorization**: the chart's `AZURE_STORAGE_KEY` is stored in a Kubernetes `Secret`. Rotate it by updating the value and `helm upgrade`.
