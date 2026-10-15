{{- /*gotype:github.com/Necoro/feed2imap-go/internal/feed.Item*/ -}}
{{define "bottomLineStart"}}
    <tr>
      <td style="text-align: right; padding: 0">
        <span style="color: #ababab">{{.}}</span>&nbsp;&nbsp;
      </td>
      <td style="padding: 0">
        <span style="color: #ababab">
{{- end}}
{{define "bottomLineEnd" -}}
        </span>
      </td>
    </tr>
{{end}}
{{define "bottomLine"}}
  {{if .content}}
    {{- template "bottomLineStart" .descr}}{{.content}}{{template "bottomLineEnd"}}
  {{- end}}
{{end}}
{{define "link"}}<a style="color: #ababab;" href="{{.url}}">{{.text}}</a>{{end}}
<table style="border: 2px black groove; background: #ededed; width: 100%; margin-bottom: 5px">
  <tr>
    <td style="text-align: right; padding: 4px"><strong>Feed</strong></td>
    <td style="width: 100%; padding: 4px">
      {{with .Feed.Link}}<a href="{{.}}">{{end}}
        <strong>{{or .Feed.Title .Feed.Link "Unnamed feed"}}</strong>
        {{if .Feed.Link}}</a>{{end}}
    </td>
  </tr>
  <tr>
    <td style="text-align: right; padding: 4px"><strong>Item</strong></td>
    <td style="width: 100%; padding: 4px">
      {{with .Item.Link}}<a href="{{.}}">{{end}}
        <strong>{{or .Item.Title .Item.Link}}</strong>
        {{if .Item.Link}}</a>{{end}}
    </td>
  </tr>
</table>
{{with .Body}}
  {{html .}}
{{end}}
{{with .Item.Enclosures}}
  <table style="border: 2px black groove; background: #ededed; width: 100%; margin-top: 5px">
    <tr>
      <td style="width: 100%"><strong>Files:</strong></td>
    </tr>
    {{range .}}
      <tr>
        <td>
          &nbsp;&nbsp;&nbsp;
          <a href={{.URL}}>{{.URL | lastUrlPart}}</a> ({{with .Length}}{{. | byteCount}}, {{end}}{{.Type}})
        </td>
      </tr>
    {{end}}
  </table>
{{end}}
<hr style="width: 100%"/>
<table style="width: 100%; border-spacing: 0">
  {{template "bottomLine" (dict "descr" "Date:" "content" .Date)}}
  {{with .AuthorList}}
    {{- template "bottomLineStart" "Author:"}}
    {{- range $i, $p := .}}
      {{- if $i}}, {{end}}
      {{- if .URL}}{{template "link" (dict "url" .URL "text" (or .Name .URL))}}{{else}}{{.Name}}{{end}}
    {{- end}}
    {{- template "bottomLineEnd"}}
  {{- end}}
  {{template "bottomLine" (dict "descr" "Filed under:" "content" (join ", " .Categories))}}
  {{with .FeedLink}}
    {{- template "bottomLineStart" "Feed-Link:"}}{{template "link" (dict "url" . "text" .)}}{{template "bottomLineEnd"}}
  {{- end}}
</table>