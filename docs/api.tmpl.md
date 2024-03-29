<!-- This file is auto-generated. Please do not modify it yourself. -->
<!-- markdown-link-check-disable -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents
{{range .Files}}
{{$file_name := .Name}}- [{{.Name}}](#{{.Name}})
  {{- if .Messages }}
  {{range .Messages}}  - [{{.LongName}}](#{{.FullName}})
  {{end}}
  {{- end -}}
  {{- if .Enums }}
  {{range .Enums}}  - [{{.LongName}}](#{{.FullName}})
  {{end}}
  {{- end -}}
  {{- if .Extensions }}
  {{range .Extensions}}  - [File-level Extensions](#{{$file_name}}-extensions)
  {{end}}
  {{- end -}}
  {{- if .Services }}
  {{range .Services}}  - [{{.Name}}](#{{.FullName}})
  {{end}}
  {{- end -}}
{{end}}
- [Scalar Value Types](#scalar-value-types)

{{range .Files}}
{{$file_name := .Name}}
<a name="{{.Name}}"></a>
<p align="right"><a href="#top">Top</a></p>

## {{.Name}}
{{if .Description}}
```
{{.Description}}
```
{{end}}

{{range .Messages}}
<a name="{{.FullName}}"></a>

### {{.LongName}}
{{if .Description}}
```
{{.Description}}
```
{{end}}

{{if .HasFields}}
| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
{{range .Fields -}}
  | `{{.Name}}` | [{{.LongType}}](#{{.FullType}}) | {{.Label}} | {{if (index .Options "deprecated"|default false)}}**Deprecated.** {{end}} {{if .Description}}`{{.Description | replace "\n" " " | replace "\\`" ""}}`{{end}} {{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}
{{end}}

{{if .HasExtensions}}
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
{{range .Extensions -}}
  | `{{.Name}}` | {{.LongType}} | {{.ContainingLongType}} | {{.Number}} | {{if .Description}}`{{.Description | replace "\n" " " | replace "/`" ""}}`{{end}} {{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}
{{end}}

{{end}} <!-- end messages -->

{{range .Enums}}
<a name="{{.FullName}}"></a>

### {{.LongName}}
{{if .Description}}
```
{{.Description}}
```
{{end}}


| Name | Number | Description |
| ---- | ------ | ----------- |
{{range .Values -}}
  | {{.Name}} | {{.Number}} | {{if .Description}}`{{.Description | replace "\n" " " | replace "\\`" ""}}`{{end}} |
{{end}}

{{end}} <!-- end enums -->

{{if .HasExtensions}}
<a name="{{$file_name}}-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
{{range .Extensions -}}
  | `{{.Name}}` | {{.LongType}} | {{.ContainingLongType}} | {{.Number}} | {{if .Description}}`{{.Description | replace "\n" " " | replace "\\`" ""}}`{{end}} {{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}
{{end}} <!-- end HasExtensions -->

{{range .Services}}
<a name="{{.FullName}}"></a>

### {{.Name}}
{{if .Description}}
```
{{.Description}}
```
{{end}}

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
{{range .Methods -}}
  | `{{.Name}}` | [{{.RequestLongType}}](#{{.RequestFullType}}){{if .RequestStreaming}} stream{{end}} | [{{.ResponseLongType}}](#{{.ResponseFullType}}){{if .ResponseStreaming}} stream{{end}} | {{if .Description}}`{{.Description | replace "\n" " " | replace "\\`" ""}}`{{end}} | {{with (index .Options "google.api.http")}}{{range .Rules}}{{.Method}}|{{.Pattern}}{{end}}{{end}} |
{{end}}
{{end}} <!-- end services -->

{{end}}

## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
{{range .Scalars -}}
  | <a name="{{.ProtoType}}" /> {{.ProtoType}} | {{.Notes}} | {{.CppType}} | {{.JavaType}} | {{.PythonType}} | {{.GoType}} | {{.CSharp}} | {{.PhpType}} | {{.RubyType}} |
{{end}}

<!-- markdown-link-check-enable -->
