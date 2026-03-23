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
