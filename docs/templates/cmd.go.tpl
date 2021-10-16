## Use

`{{ replace (printf "%s " .UseLine) "[flags]" "" -1 -}}
{{ visitFlags . -}}` 

{{ if ne (len .Aliases) 0 -}}
{{ $aliases := .Aliases -}}
*Aliases*: {{ range $i, $v := $aliases -}}
		{{ $v -}}
		{{ if ne $i (sub (len $aliases) 1) -}} , {{ end -}}
	{{ end }}.
{{ end }}
## Description

{{ if ne .Long "" -}}
	{{ .Long }}
{{ else -}}
	{{ .Short }}.
{{ end -}}

{{ if .HasSubCommands }}
## Subcommands
	{{ $url := getURL . -}} 
	{{ range .Commands }}
		{{- $name := .Name -}}
		{{ if not .HasSubCommands -}}
			{{ $name = printf "%s.md" .Name }}
		{{- end }}
- [`{{ .CommandPath }}`]({{ $url }}{{ $name }}): {{ .Short }}.
	{{- end }}
{{ end }}
## Flags

{{ if .Flags.HasFlags -}}
| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
{{ visitFlagsTable . -}}
{{- else -}}
No flags.
{{ end }}
{{ if ne .Example "" -}}
## Examples
{{ $examples := split .Example "\n" -}}
	{{- range $i, $v := $examples -}}
		{{- if hasPrefix $v "*" -}}
			{{- replace $v "* " "" 1 -}}:
		{{- else if ne $v "" -}}
```
{{ $v }}
```
{{- end }}
{{ end -}}
{{- end }}