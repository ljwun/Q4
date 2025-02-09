{{/*
此 helper 函數 "utils.envValue" 接受一個物件，該物件可包含以下欄位：
  - name（字串，用以指定環境變數的名稱）
  - data (物件，包含以下欄位)：
    - secretName（字串，用以指定 secret 的名稱）
    - configMapName（字串，用以指定 configMap 的名稱）
    - key（字串，用以指定 secret 或 configMap 中的 key）
    - value（字串，用以指定環境變數的值）
  - required（布林值，用以標示是否必須設定上述其中一種）
  - default（在非 required 狀況下，當上述欄位都不存在時，將用此預設值）

使用範例：
  {{ include "utils.envValue" (dict "data" .Values.api.oidc.issuerUrl "required" true "default" "defaultUrl" ) }}
注意：
  - 如果 required 為 true，必須至少提供下列其中一組：
      * secretName 與 key
      * configMapName 與 key
      * value
  - 如果 required 為 false，且前述欄位皆未設定，但提供了 default，則會設成：
      value: "<default的內容>"
*/}}
{{- define "utils.envValue" -}}
{{- /* 若變數設定為必填，則必須提供至少一組有效設定 */ -}}
{{- if .required }}
  {{- if not (or (and .data.secretName .data.key) (and .data.configMapName .data.key) .data.value) }}
    {{- fail (print "環境變數" .name "是必要設定，但未提供 secretName/key、configMapName/key 或 value") }}
  {{- end }}
{{- end }}

{{- /* 依優先順序檢查並輸出：先檢查 secret，再檢查 configMap，最後為字面值 */ -}}
- name: {{ .name | quote }}
  {{- if .data.secretName }}
  valueFrom:
    secretKeyRef:
      name: {{ .data.secretName | quote }}
      key: {{ .data.key | quote }}
  {{- else if .data.configMapName }}
  valueFrom:
    configMapKeyRef:
      name: {{ .data.configMapName | quote }}
      key: {{ .data.key | quote }}
  {{- else if .data.value }}
  value: {{ .data.value | quote }}
  {{- else if .default }}
  value: {{ .default | quote }}
  {{- end }}
{{- end }}