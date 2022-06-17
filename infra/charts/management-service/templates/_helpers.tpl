{{/*
Expand the name of the chart.
*/}}
{{- define "xp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "xp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "xp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "xp.environment" -}}
{{- .Values.global.environment | default .Values.xpManagement.environment | default "dev" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "xp.labels" -}}
helm.sh/chart: {{ include "xp.chart" . }}
{{- with .Values.xpManagement.extraLabels }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{ include "xp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "xp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "xp.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "xp.serviceAccountName" -}}
{{- if .Values.xpManagement.serviceAccount.create }}
{{- default (include "xp.fullname" .) .Values.xpManagement.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.xpManagement.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "xp.sentry.enabled" -}}
{{ eq ((.Values.xpManagement.apiConfig.sentryConfig).enabled | toString) "true" }}
{{- end -}}

{{- define "xp.sentry.dsn" -}}
{{- .Values.global.sentry.dsn | default .Values.xpManagement.sentry.dsn -}}
{{- end -}}

{{- define "xp.ui.defaultConfig" -}}
{{- if .Values.xpManagement.uiConfig -}}
appConfig:
  environment: {{ .Values.xpManagement.uiConfig.appConfig.environment | default (include "xp.environment" .) }}
authConfig:
  oauthClientId: {{ .Values.global.oauthClientId | default .Values.xpManagement.uiConfig.authConfig.oauthClientId | quote }}
{{- if (include "xp.sentry.enabled" .) }}
sentryConfig:
  environment: {{ .Values.xpManagement.uiConfig.sentryConfig.environment | default (include "xp.environment" .) }}
  dsn: {{ .Values.xpManagement.uiConfig.sentryConfig.dsn | default (include "xp.sentry.dsn" .) | quote }}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "xp.ui.config" -}}
{{- $defaultConfig := include "xp.ui.defaultConfig" . | fromYaml -}}
{{ .Values.xpManagement.uiConfig | merge $defaultConfig | toPrettyJson }}
{{- end -}}
