package front

import "html/template"

func Documents() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/documents.html")...))
}

func DocumentsNewForm() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/documents_form.html")...))
}
