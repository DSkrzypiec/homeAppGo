package front

import "html/template"

type CountersData struct{}

func Counters() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/counters.html")...))
}
