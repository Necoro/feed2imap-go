package template

var Feed = fromString("Feed", feedTpl)

//noinspection HtmlDeprecatedAttribute,HtmlUnknownTarget
const feedTpl = `{{- /*gotype:github.com/Necoro/feed2imap-go/internal/feed.feeditem*/ -}}
{{define "bottomLine"}}
  {{if .content}}
    <tr>
      <td align="right">
        <span style="color: #ababab; ">{{.descr}}</span>&nbsp;&nbsp;
      </td>
      <td>
        <span style="color: #ababab; ">{{.content}}</span>
      </td>
    </tr>
  {{end}}
{{end}}
<table border="1" width="100%" cellpadding="0" cellspacing="0" style="border-spacing: 0; ">
  <tr>
    <td>
      <table width="100%" bgcolor="#EDEDED" cellpadding="4" cellspacing="2">
        <tr>
          <td align="right"><b>Feed</b></td>
          <td width="100%">
            {{with .Feed.Link}}<a href="{{.}}">{{end}}
              <b>{{or .Feed.Title .Feed.Link "Unnammed feed"}}</b>
            {{if .Feed.Link}}</a>{{end}}
          </td>
        </tr>
        <tr>
          <td align="right"><b>Item</b></td>
          <td width="100%">
            {{with .Item.Link}}<a href="{{.}}">{{end}}
              <b>{{or .Item.Title .Item.Link}}</b>
            {{if .Item.Link}}</a>{{end}}
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
{{with .Body}}
  {{html .}}
{{end}}
{{with .Item.Enclosures}}
  <table border="1" width="100%" cellpadding="0" cellspacing="0" style="border-spacing: 0; ">
    <tr>
      <td>
        <table width="100%" bgcolor="#EDEDED" cellpadding="2" cellspacing="2">
          <tr><td width="100%"><b>Files:</b></td></tr>
          {{range .}}
              <tr>
                <td>
                  &nbsp;&nbsp;&nbsp;
                  <a href={{.URL}}>{{.URL | lastUrlPart}}</a> ({{with .Length}}{{. | byteCount}}, {{end}}{{.Type}})
                </td>
              </tr>
          {{end}}
        </table>
      </td>
    </tr>
  </table>
{{end}}
<hr width="100%"/>
<table width="100%" cellpadding="0" cellspacing="0">
  {{template "bottomLine" (dict "descr" "Date:" "content" .Item.Published)}}
  {{template "bottomLine" (dict "descr" "Author:" "content" .Creator)}}
  {{template "bottomLine" (dict "descr" "Filed under:" "content" (join ", " .Item.Categories))}}
  {{with .Feed.FeedLink}}
    {{template "bottomLine" (dict "descr" "Feed-Link:" "content" (print "<a style=\"color: #ababab;\" href=\"" . "\">" . "</a>" | html))}}
  {{end}}
</table>`
