{{- define "kfleet-hub.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kfleet-hub.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "kfleet-hub.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kfleet-hub.labels" -}}
helm.sh/chart: {{ include "kfleet-hub.chart" . }}
{{ include "kfleet-hub.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "kfleet-hub.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kfleet-hub.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "kfleet-hub.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "kfleet-hub.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Fail the install when configuration cannot produce a runnable, safe hub.
Called once from deployment.yaml so every rendering path hits it exactly once.
*/}}
{{- define "kfleet-hub.validateValues" -}}
{{- if and (not .Values.registration.token) (not .Values.registration.existingSecret) -}}
{{- fail "kfleet-hub: registration.token is empty and registration.existingSecret is not set. Set registration.token or point registration.existingSecret at a Secret holding the agent registration token. Installing with an empty token would leave agent registration open." -}}
{{- end -}}
{{- if gt (int .Values.replicaCount) 1 -}}
{{- fail "kfleet-hub: the hub uses SQLite and must run as a single replica (set replicaCount to 1)" -}}
{{- end -}}
{{- end -}}

{{/*
Secret holding the agent registration token: an external Secret when
registration.existingSecret is set, otherwise the chart-managed Secret.
*/}}
{{- define "kfleet-hub.registration.secretName" -}}
{{- default (include "kfleet-hub.fullname" .) .Values.registration.existingSecret -}}
{{- end -}}

{{- define "kfleet-hub.registration.secretKey" -}}
{{- default "registration-token" .Values.registration.existingSecretKey -}}
{{- end -}}

{{/*
Secret holding the bootstrap admin credentials: an external Secret when
auth.bootstrapAdmin.existingSecret is set, otherwise the chart-managed Secret.
*/}}
{{- define "kfleet-hub.bootstrapAdmin.secretName" -}}
{{- default (include "kfleet-hub.fullname" .) .Values.auth.bootstrapAdmin.existingSecret -}}
{{- end -}}

{{- define "kfleet-hub.bootstrapAdmin.usernameKey" -}}
{{- default "bootstrap-admin-username" .Values.auth.bootstrapAdmin.existingSecretUsernameKey -}}
{{- end -}}

{{- define "kfleet-hub.bootstrapAdmin.emailKey" -}}
{{- default "bootstrap-admin-email" .Values.auth.bootstrapAdmin.existingSecretEmailKey -}}
{{- end -}}

{{- define "kfleet-hub.bootstrapAdmin.passwordKey" -}}
{{- default "bootstrap-admin-password" .Values.auth.bootstrapAdmin.existingSecretPasswordKey -}}
{{- end -}}
