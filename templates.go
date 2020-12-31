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
<ol>
{{ range .results }}
<li>
{{.Date.Format "2006-02-01"}}: {{printf "%.2f" .Distance}}km (<a href="/where/ui/?start={{.Date.Format "2006-02-01"}}T00%3A00%3A00&end={{.Date.Format "2006-02-01"}}T23%3A59%3A59&layers=last,line,points" title="Map">Map</a>)
</li>
{{ end }}
</ol>
</body>
</html>
`
)

func BuildTemplates() *template.Template {
	t := template.Must(template.New("place").Parse(placeTemplate))
	t = template.Must(t.New("placeResults").Parse(placeResultsTemplate))
	return t
}
