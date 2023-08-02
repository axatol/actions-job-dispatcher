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

{{- define "actions-job-dispatcher.githubAuthSecretName" -}}
{{ default (printf "%s-github-credentials" .Release.Name) .Values.github.authSecret.name }}
{{- end -}}

{{- define "actions-job-dispatcher.dispatcherConfigName" -}}
{{ default .Release.Name .Values.dispatcher.config.name }}
{{- end -}}

{{- define "actions-job-dispatcher.serviceAccountName" -}}
{{ default .Release.Name .Values.rbac.serviceAccountName }}
{{- end -}}

{{- define "actions-job-dispatcher.serviceName" -}}
{{ default .Release.Name .Values.service.name }}
{{- end -}}
