package front

import "html/template"

type HomeData struct{}

func Home() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/home.html")...))
}
