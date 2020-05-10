package template

var Html = fromString("Feed", feedTpl)

//noinspection HtmlDeprecatedAttribute,HtmlUnknownTarget
const feedTpl = `{{- /*gotype:github.com/Necoro/feed2imap-go/internal/feed.feeditem*/ -}}
{{define "bottomLine"}}
  {{if .content}}
    <tr>
      <td style="text-align: right; padding: 0">
        <span style="color: #ababab">{{.descr}}</span>&nbsp;&nbsp;
      </td>
      <td style="padding: 0">
        <span style="color: #ababab">{{.content}}</span>
      </td>
    </tr>
  {{end}}
{{end}}
<table style="border: 2px black groove; background: #ededed; width: 100%; margin-bottom: 5px">
    <tr>
        <td style="text-align: right; padding: 4px"><strong>Feed</strong></td>
        <td style="width: 100%; padding: 4px">
            {{with .Feed.Link}}<a href="{{.}}">{{end}}
                <strong>{{or .Feed.Title .Feed.Link "Unnammed feed"}}</strong>
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
  {{template "bottomLine" (dict "descr" "Author:" "content" .Creator)}}
  {{template "bottomLine" (dict "descr" "Filed under:" "content" (join ", " .Categories))}}
  {{with .FeedLink}}
    {{template "bottomLine" (dict "descr" "Feed-Link:" "content" (print "<a style=\"color: #ababab;\" href=\"" . "\">" . "</a>" | html))}}
  {{end}}
</table>`
