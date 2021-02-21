{{- /*gotype:github.com/Necoro/feed2imap-go/internal/feed.feeditem*/ -}}
{{- with .Item.Link -}}
<{{.}}>

{{ end -}}
{{- with .TextBody -}}
{{.}}
{{ end -}}
{{- with .Item.Enclosures}}
Files:
{{- range . }}
  {{ .URL}} ({{with .Length}}{{. | byteCount}}, {{end}}{{.Type}})
{{- end -}}
{{- end}}
-- 
Feed: {{ with .Feed.Title -}}{{.}}{{- end }}
{{ with .Feed.Link -}}
  <{{.}}>
{{end -}}
Item: {{ with .Item.Title -}}
  {{.}}
{{- end }}
{{ with .Item.Link -}}
  <{{.}}>
{{end -}}
{{ with .Date -}}
  Date: {{.}}
{{ end -}}
{{ with .Creator -}}
  Author: {{.}}
{{ end -}}
{{ with (join ", " .Categories) -}}
  Filed under: {{.}}
{{ end -}}
{{ with .FeedLink -}}
  Feed-Link: {{.}}
{{ end -}}