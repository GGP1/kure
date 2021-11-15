For more information about a command, its flags and examples please visit the [commands folder](https://github.com/GGP1/kure/tree/master/docs/commands).

{{ range .Commands -}}
- [{{ .Name }}](#{{ .Name }})
{{ end }}
{{ range .Commands }}
### {{ .Name }}
```
{{ cmdAndFlags . -}}
{{ subCmds . "" }}
```

---
{{ end }}