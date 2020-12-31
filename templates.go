package main

import "html/template"

const (
	placeTemplate = `<html>
<body>
<h1>Places!</h1>
<form method="post">
<fieldset>
<input type="text" placeholder="Search for place" name="place">
</fieldset>
</form>
</body>
</html>
`
	placeResultsTemplate = `<html>
<body>
<h1>Places!</h1>
<form method="post">
<fieldset>
<input type="text" placeholder="Search for place" name="place">
</fieldset>
</form>
<h2>Results</h2>
<table>

<tr>
<th></th>
<th>Date</th>
<th>Distance</th>
<th></th>
</tr>

{{ range $i, $result := .results }}
<tr>
<td>{{$i}}</td>
<td>{{$result.Date.Format "2006-01-02"}}</td>
<td>{{printf "%.2f" $result.Distance}}km</td>
<td><a href="/where/ui/?start={{$result.Date.Format "2006-01-02"}}T00%3A00%3A00&end={{$result.Date.Format "2006-01-02"}}T23%3A59%3A59&layers=last,line,points" title="Map">Map</a></td>
</tr>
{{ end }}
</table>
</body>
</html>
`
)

func BuildTemplates() *template.Template {
	t := template.Must(template.New("place").Parse(placeTemplate))
	t = template.Must(t.New("placeResults").Parse(placeResultsTemplate))
	return t
}
