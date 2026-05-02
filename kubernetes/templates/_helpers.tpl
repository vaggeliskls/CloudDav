{{/*
Expand the name of the chart.
*/}}
{{- define "cloud-webdav-server.name" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cloud-webdav-server.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "cloud-webdav-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cloud-webdav-server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cloud-webdav-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Render folderPermissions array as the comma-separated /path:users:mode string
the server expects via FOLDER_PERMISSIONS.
*/}}
{{- define "cloud-webdav-server.folderPermissions" -}}
{{- $rules := list -}}
{{- range .Values.folderPermissions -}}
{{- $rules = append $rules (printf "%s:%s:%s" .path .users .mode) -}}
{{- end -}}
{{- join "," $rules -}}
{{- end }}
