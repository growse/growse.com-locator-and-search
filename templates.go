package main

import "html/template"

const (
	includesTemplate = `{{define "header"}} <html>
	<head>
	<link rel="stylesheet" href="https://unpkg.com/purecss@2.0.3/build/pure-min.css" integrity="sha384-cg6SkqEOCV1NbJoCu11+bm0NvBRc8IYLRGXkmNrqUBfTjmMYwNKPWBTIKyw9mHNJ" crossorigin="anonymous">
<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
<style>
    body {
        margin:1rem;
    }
</style>
{{ end }}
{{define "form"}}
<div class="pure-g">
<div class="pure-u-1">
<h1>Places!</h1>
</div>
</div>
<div class="pure-g">
<div class="pure-u-1">
<form method="post" class="pure-form">
<fieldset>
<input type="text" placeholder="Search for place" name="place" value="{{.place}}">
</fieldset>
</form>
</div>
</div>
{{end}}
{{define "footer"}}
</body>
</html>
{{end}}`
	placeTemplate = `{{template "header"}}
{{template "form"}}
{{template "footer"}}
`
	placeResultsTemplate = `{{template "header"}}
{{template "form"}}
<div class="pure-g">
<div class="pure-u-1">
<h2>Results for "{{.place}}"</h2>
<p>{{.formatted}}</p>
</div>
</div>
<div class="pure-g">
<div class="pure-u-1">
<table class="pure-table">

<tr>
<th>Date</th>
<th>Number of locations</th>
<th>Link</th>
</tr>

{{ range $i, $result := .results }}
<tr>
<td>{{$result.Date.Format "2 January 2006"}}</td>
<td>{{$result.LocationCount}}</td>
<td><a href="/where/ui/?start={{$result.Date.Format "2006-01-02"}}T00%3A00%3A00&end={{$result.Date.Format "2006-01-02"}}T23%3A59%3A59&layers=last,line,points" title="Map">Map</a></td>
</tr>
{{ else }}
<tr><td colspan="3">No results</td></tr>
{{ end }}

</table>
</div>
</div>
{{template "footer"}}
`
)

func BuildTemplates() *template.Template {
	t := template.Must(template.New("place").Parse(placeTemplate))
	t = template.Must(t.New("placeResults").Parse(placeResultsTemplate))
	t = template.Must(t.New("placeResults").Parse(includesTemplate))
	return t
}
