{{- range . -}}
This product includes {{ .Name }} {{ .Version }} licensed under {{ .LicenseName }} license.

{{ .LicenseText }}

{{ end }}
