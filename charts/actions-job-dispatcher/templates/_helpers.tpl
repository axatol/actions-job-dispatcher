{{/*
Common labels
*/}}
{{- define "actions-job-dispatcher.labels" -}}
helm.sh/chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ include "actions-job-dispatcher.selectors" . }}
{{- end -}}

{{/*
Common selectors
*/}}
{{- define "actions-job-dispatcher.selectors" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
