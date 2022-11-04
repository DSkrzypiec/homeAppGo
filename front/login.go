package front

import "html/template"

func Login() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/login.html")...))
}
