{{/*
Expand the name of the chart.
*/}}
{{- define "traktor.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "traktor.fullname" -}}
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
{{- define "traktor.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "traktor.labels" -}}
helm.sh/chart: {{ include "traktor.chart" . }}
{{ include "traktor.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "traktor.selectorLabels" -}}
app.kubernetes.io/name: {{ include "traktor.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "traktor.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "traktor.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the image path
*/}}
{{- define "traktor.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
Create the namespace name
*/}}
{{- define "traktor.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride }}
{{- end }}

{{/*
Metrics service name
*/}}
{{- define "traktor.metricsServiceName" -}}
{{- printf "%s-metrics-service" (include "traktor.fullname" .) }}
{{- end }}

{{/*
Controller manager name
*/}}
{{- define "traktor.controllerManagerName" -}}
{{- printf "%s-controller-manager" (include "traktor.fullname" .) }}
{{- end }}

{{/*
Leader election role name
*/}}
{{- define "traktor.leaderElectionRoleName" -}}
{{- printf "%s-leader-election-role" (include "traktor.fullname" .) }}
{{- end }}

{{/*
Manager role name
*/}}
{{- define "traktor.managerRoleName" -}}
{{- printf "%s-manager-role" (include "traktor.fullname" .) }}
{{- end }}

{{/*
Metrics reader role name
*/}}
{{- define "traktor.metricsReaderRoleName" -}}
{{- printf "%s-metrics-reader" (include "traktor.fullname" .) }}
{{- end }}

{{/*
Proxy role name
*/}}
{{- define "traktor.proxyRoleName" -}}
{{- printf "%s-proxy-role" (include "traktor.fullname" .) }}
{{- end }}